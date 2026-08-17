package checks

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	multiagentspec "github.com/plexusone/multi-agent-spec/sdk/go"

	"github.com/grokify/releasekit/run"
)

// GoChecker validates Go projects before release.
type GoChecker struct {
	Opts Options
}

func (c *GoChecker) Name() string { return "Go" }

func (c *GoChecker) Check(ctx context.Context, r run.Runner, dir string) []multiagentspec.TaskResult {
	var results []multiagentspec.TaskResult

	// Always check for local replace directives
	results = append(results, c.checkNoLocalReplace(ctx, r, dir))

	if c.Opts.ModTidy {
		results = append(results, c.checkModTidy(ctx, r, dir))
	}
	if c.Opts.Build {
		results = append(results, c.checkBuild(ctx, r, dir))
	}
	if c.Opts.Format {
		results = append(results, c.checkFormat(ctx, r, dir))
	}
	if c.Opts.Lint {
		results = append(results, c.checkLint(ctx, r, dir))
	}
	if c.Opts.Test {
		results = append(results, c.checkTest(ctx, r, dir))
	}

	// Vulnerability check (always runs if available)
	results = append(results, c.checkVulncheck(ctx, r, dir))

	// Error handling compliance (always runs)
	results = append(results, c.checkErrorHandling(dir))

	// Soft checks (warnings only)
	results = append(results, c.checkUntrackedReferences(ctx, r, dir))

	if c.Opts.Coverage {
		results = append(results, c.checkCoverage(ctx, r, dir))
	}

	return results
}

func (c *GoChecker) checkNoLocalReplace(ctx context.Context, r run.Runner, dir string) multiagentspec.TaskResult {
	id := "go:no-local-replace"

	cmd := run.Shell(dir, "grep", "-n", "replace.*=>.*\\./\\|replace.*=>.*\\.\\./", filepath.Join(dir, "go.mod"))
	res := r.Run(ctx, cmd)

	if res.ExitCode == 0 && strings.TrimSpace(res.Stdout) != "" {
		return taskResultWithOutput(id, multiagentspec.StatusNoGo,
			"go.mod contains local replace directives",
			strings.TrimSpace(res.Stdout))
	}

	return taskResult(id, multiagentspec.StatusGo, "no local replace directives")
}

func (c *GoChecker) checkModTidy(ctx context.Context, r run.Runner, dir string) multiagentspec.TaskResult {
	id := "go:mod-tidy"

	// Try go mod tidy -diff first (Go 1.21+)
	diffCmd := run.Go(dir, "mod", "tidy", "-diff")
	res := r.Run(ctx, diffCmd)

	// If -diff is not supported, fall back to running tidy and checking git diff
	if res.ExitCode != 0 && strings.Contains(res.Stderr, "unknown flag") {
		return c.checkModTidyFallback(ctx, r, dir, id)
	}

	output := strings.TrimSpace(res.Stdout)
	if output != "" {
		return taskResultWithOutput(id, multiagentspec.StatusNoGo,
			"go.mod or go.sum needs updating (run 'go mod tidy')",
			truncateOutput(output, 1000))
	}

	return taskResult(id, multiagentspec.StatusGo, "go.mod is tidy")
}

func (c *GoChecker) checkModTidyFallback(ctx context.Context, r run.Runner, dir, id string) multiagentspec.TaskResult {
	tidyCmd := run.Go(dir, "mod", "tidy")
	res := r.Run(ctx, tidyCmd)
	if res.Err != nil {
		return taskResultWithOutput(id, multiagentspec.StatusNoGo, "go mod tidy failed", res.Stderr)
	}

	diffCmd := run.Git(dir, "diff", "--name-only", "go.mod", "go.sum")
	diffRes := r.Run(ctx, diffCmd)
	if diffRes.ExitCode == 0 && strings.TrimSpace(diffRes.Stdout) != "" {
		restoreCmd := run.Git(dir, "checkout", "--", "go.mod", "go.sum")
		r.Run(ctx, restoreCmd)

		return taskResultWithOutput(id, multiagentspec.StatusNoGo,
			"go.mod or go.sum needs updating (run 'go mod tidy')",
			strings.TrimSpace(diffRes.Stdout))
	}

	return taskResult(id, multiagentspec.StatusGo, "go.mod is tidy")
}

func (c *GoChecker) checkBuild(ctx context.Context, r run.Runner, dir string) multiagentspec.TaskResult {
	id := "go:build"

	cmd := run.Go(dir, "build", "./...")
	res := r.Run(ctx, cmd)
	if res.Err != nil || res.ExitCode != 0 {
		output := res.Stderr
		if output == "" {
			output = res.Stdout
		}
		return taskResultWithOutput(id, multiagentspec.StatusNoGo, "build failed", output)
	}

	return taskResult(id, multiagentspec.StatusGo, "build succeeded")
}

func (c *GoChecker) checkFormat(ctx context.Context, r run.Runner, dir string) multiagentspec.TaskResult {
	id := "go:format"

	cmd := run.Shell(dir, "gofmt", "-l", ".")
	res := r.Run(ctx, cmd)
	if res.Err != nil {
		if !commandExists("gofmt") {
			return taskResult(id, multiagentspec.StatusSkip, "gofmt not found")
		}
		return taskResultWithOutput(id, multiagentspec.StatusNoGo, "gofmt check failed", res.Stderr)
	}

	unformatted := strings.TrimSpace(res.Stdout)
	if unformatted != "" {
		return taskResultWithOutput(id, multiagentspec.StatusNoGo, "files need formatting", unformatted)
	}

	return taskResult(id, multiagentspec.StatusGo, "all files formatted")
}

func (c *GoChecker) checkLint(ctx context.Context, r run.Runner, dir string) multiagentspec.TaskResult {
	id := "go:lint"

	if !commandExists("golangci-lint") {
		return taskResult(id, multiagentspec.StatusSkip, "golangci-lint not installed")
	}

	cmd := run.Shell(dir, "golangci-lint", "run", "./...")
	res := r.Run(ctx, cmd)
	if res.ExitCode != 0 {
		output := res.Stdout
		if output == "" {
			output = res.Stderr
		}
		return taskResultWithOutput(id, multiagentspec.StatusNoGo, "linting issues found", truncateOutput(output, 2000))
	}

	return taskResult(id, multiagentspec.StatusGo, "no lint issues")
}

func (c *GoChecker) checkTest(ctx context.Context, r run.Runner, dir string) multiagentspec.TaskResult {
	id := "go:test"

	args := []string{"test", "./..."}
	if c.Opts.Verbose {
		args = []string{"test", "-v", "./..."}
	}

	cmd := run.Go(dir, args...)
	res := r.Run(ctx, cmd)
	if res.ExitCode != 0 {
		output := res.Stdout
		if output == "" {
			output = res.Stderr
		}
		return taskResultWithOutput(id, multiagentspec.StatusNoGo, "tests failed", truncateOutput(output, 3000))
	}

	return taskResult(id, multiagentspec.StatusGo, "all tests passed")
}

func (c *GoChecker) checkVulncheck(ctx context.Context, r run.Runner, dir string) multiagentspec.TaskResult {
	id := "go:vulncheck"

	if !commandExists("govulncheck") {
		return taskResult(id, multiagentspec.StatusSkip,
			"govulncheck not installed (go install golang.org/x/vuln/cmd/govulncheck@latest)")
	}

	cmd := run.Shell(dir, "govulncheck", "./...")
	res := r.Run(ctx, cmd)
	if res.ExitCode != 0 {
		output := res.Stdout
		if output == "" {
			output = res.Stderr
		}
		return taskResultWithOutput(id, multiagentspec.StatusNoGo, "vulnerabilities found", truncateOutput(output, 3000))
	}

	return taskResult(id, multiagentspec.StatusGo, "no vulnerabilities found")
}

func (c *GoChecker) checkCoverage(ctx context.Context, r run.Runner, dir string) multiagentspec.TaskResult {
	id := "go:coverage"

	args := []string{"test", "-coverprofile=coverage.out"}
	if c.Opts.GoExcludeCoverage != "" {
		listCmd := run.Go(dir, "list", "./...")
		listRes := r.Run(ctx, listCmd)
		if listRes.Err != nil {
			return taskResultWithOutput(id, multiagentspec.StatusWarn, "could not list packages for coverage", listRes.Stderr)
		}

		var pkgs []string
		scanner := bufio.NewScanner(strings.NewReader(listRes.Stdout))
		for scanner.Scan() {
			pkg := scanner.Text()
			if !strings.Contains(pkg, "/"+c.Opts.GoExcludeCoverage+"/") &&
				!strings.HasSuffix(pkg, "/"+c.Opts.GoExcludeCoverage) {
				pkgs = append(pkgs, pkg)
			}
		}

		if len(pkgs) == 0 {
			return taskResult(id, multiagentspec.StatusSkip, "no packages to test after exclusion")
		}

		args = append([]string{"test", "-coverprofile=coverage.out"}, pkgs...)
	} else {
		args = append(args, "./...")
	}

	cmd := run.Go(dir, args...)
	res := r.Run(ctx, cmd)
	if res.ExitCode != 0 {
		return taskResultWithOutput(id, multiagentspec.StatusWarn, "coverage run failed", truncateOutput(res.Stderr, 1000))
	}

	coverage := parseCoverageFromOutput(res.Stdout)
	return taskResult(id, multiagentspec.StatusGo, fmt.Sprintf("coverage: %s", coverage))
}

// checkErrorHandling scans Go source files for improper error handling patterns.
func (c *GoChecker) checkErrorHandling(dir string) multiagentspec.TaskResult {
	id := "go:error-handling"

	discardPattern := regexp.MustCompile(`_\s*=\s*err\b`)
	discardFuncPattern := regexp.MustCompile(`_\s*=\s*\w+\([^)]*\)`)
	discardPkgFuncPattern := regexp.MustCompile(`_\s*=\s*\w+\.\w+\([^)]*\)`)

	var violations []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer func() { _ = file.Close() }() // read-only file, close error is non-actionable

		scanner := bufio.NewScanner(file)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") {
				continue
			}

			if discardPattern.MatchString(line) {
				relPath, _ := filepath.Rel(dir, path)
				violations = append(violations, fmt.Sprintf("%s:%d: _ = err (error discarded)", relPath, lineNum))
				continue
			}

			if (discardFuncPattern.MatchString(line) || discardPkgFuncPattern.MatchString(line)) &&
				!strings.Contains(line, "//") {
				relPath, _ := filepath.Rel(dir, path)
				violations = append(violations, fmt.Sprintf("%s:%d: potential error discard without comment", relPath, lineNum))
			}
		}

		return nil
	})

	if err != nil {
		return taskResultWithOutput(id, multiagentspec.StatusNoGo, "error scanning files", err.Error())
	}

	if len(violations) > 0 {
		output := violations
		if len(output) > 10 {
			output = output[:10]
			output = append(output, fmt.Sprintf("... and %d more violations", len(violations)-10))
		}
		return taskResultWithOutput(id, multiagentspec.StatusNoGo,
			fmt.Sprintf("%d error handling violations found", len(violations)),
			strings.Join(output, "\n"))
	}

	return taskResult(id, multiagentspec.StatusGo, "no error handling violations")
}

// checkUntrackedReferences checks if tracked Go files reference untracked files.
func (c *GoChecker) checkUntrackedReferences(ctx context.Context, r run.Runner, dir string) multiagentspec.TaskResult {
	id := "go:untracked-refs"

	untrackedCmd := run.Git(dir, "ls-files", "--others", "--exclude-standard")
	untrackedRes := r.Run(ctx, untrackedCmd)
	if untrackedRes.Err != nil {
		return taskResult(id, multiagentspec.StatusWarn, "could not list untracked files")
	}

	untrackedFiles := strings.Split(strings.TrimSpace(untrackedRes.Stdout), "\n")
	if len(untrackedFiles) == 0 || (len(untrackedFiles) == 1 && untrackedFiles[0] == "") {
		return taskResult(id, multiagentspec.StatusGo, "no untracked files")
	}

	trackedCmd := run.Git(dir, "ls-files")
	trackedRes := r.Run(ctx, trackedCmd)
	if trackedRes.Err != nil {
		return taskResult(id, multiagentspec.StatusWarn, "could not list tracked files")
	}

	trackedFiles := strings.Split(strings.TrimSpace(trackedRes.Stdout), "\n")

	var references []string
	for _, tracked := range trackedFiles {
		if !strings.HasSuffix(tracked, ".go") && tracked != "go.mod" {
			continue
		}

		for _, untracked := range untrackedFiles {
			if untracked == "" || !strings.HasSuffix(untracked, ".go") {
				continue
			}
			baseName := strings.TrimSuffix(filepath.Base(untracked), ".go")
			if baseName == "main" || baseName == "test" || baseName == "doc" {
				continue
			}

			grepCmd := run.Shell(dir, "grep", "-l", baseName, tracked)
			grepRes := r.Run(ctx, grepCmd)
			if grepRes.ExitCode == 0 && strings.TrimSpace(grepRes.Stdout) != "" {
				references = append(references, fmt.Sprintf("%s may reference untracked %s", tracked, untracked))
			}
		}
	}

	if len(references) > 0 {
		output := references
		if len(output) > 10 {
			output = output[:10]
			output = append(output, fmt.Sprintf("... and %d more", len(references)-10))
		}
		return taskResultWithOutput(id, multiagentspec.StatusWarn,
			fmt.Sprintf("%d possible untracked references", len(references)),
			strings.Join(output, "\n"))
	}

	return taskResult(id, multiagentspec.StatusGo, "no untracked references detected")
}

// parseCoverageFromOutput extracts coverage percentage from go test output.
func parseCoverageFromOutput(output string) string {
	scanner := bufio.NewScanner(strings.NewReader(output))
	var lastCoverage string
	for scanner.Scan() {
		line := scanner.Text()
		if idx := strings.Index(line, "coverage:"); idx >= 0 {
			part := line[idx:]
			lastCoverage = part
		}
	}
	if lastCoverage == "" {
		return "unknown"
	}
	return lastCoverage
}

// truncateOutput limits output to maxLen characters.
func truncateOutput(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n... (truncated)"
}
