package checks

import (
	"context"
	"path/filepath"
	"strings"

	multiagentspec "github.com/agentplexus/multi-agent-spec/sdk/go"

	"github.com/grokify/releasekit/run"
)

// TypeScriptChecker validates TypeScript/JavaScript projects before release.
type TypeScriptChecker struct {
	Opts Options
}

func (c *TypeScriptChecker) Name() string { return "TypeScript" }

func (c *TypeScriptChecker) Check(ctx context.Context, r run.Runner, dir string) []multiagentspec.TaskResult {
	var results []multiagentspec.TaskResult

	pm := detectPackageManager(dir)

	if c.Opts.Build {
		results = append(results, c.checkTypeCheck(ctx, r, dir, pm))
	}
	if c.Opts.Lint {
		results = append(results, c.checkLint(ctx, r, dir, pm))
	}
	if c.Opts.Format {
		results = append(results, c.checkFormat(ctx, r, dir, pm))
	}
	if c.Opts.Test {
		results = append(results, c.checkTest(ctx, r, dir, pm))
	}

	return results
}

func (c *TypeScriptChecker) checkTypeCheck(ctx context.Context, r run.Runner, dir string, pm string) multiagentspec.TaskResult {
	id := "ts:typecheck"

	cmd := run.Shell(dir, pm, "exec", "tsc", "--noEmit")
	if pm == "npm" {
		cmd = run.Shell(dir, "npx", "tsc", "--noEmit")
	}

	res := r.Run(ctx, cmd)
	if res.ExitCode != 0 {
		output := res.Stdout
		if output == "" {
			output = res.Stderr
		}
		return taskResultWithOutput(id, multiagentspec.StatusNoGo, "type errors found", truncateOutput(output, 2000))
	}

	return taskResult(id, multiagentspec.StatusGo, "no type errors")
}

func (c *TypeScriptChecker) checkLint(ctx context.Context, r run.Runner, dir string, pm string) multiagentspec.TaskResult {
	id := "ts:lint"

	cmd := run.Shell(dir, pm, "exec", "eslint", ".")
	if pm == "npm" {
		cmd = run.Shell(dir, "npx", "eslint", ".")
	}

	res := r.Run(ctx, cmd)
	if res.ExitCode != 0 {
		if strings.Contains(res.Stderr, "not found") || strings.Contains(res.Stderr, "Cannot find") {
			return taskResult(id, multiagentspec.StatusSkip, "eslint not configured")
		}
		output := res.Stdout
		if output == "" {
			output = res.Stderr
		}
		return taskResultWithOutput(id, multiagentspec.StatusNoGo, "lint issues found", truncateOutput(output, 2000))
	}

	return taskResult(id, multiagentspec.StatusGo, "no lint issues")
}

func (c *TypeScriptChecker) checkFormat(ctx context.Context, r run.Runner, dir string, pm string) multiagentspec.TaskResult {
	id := "ts:format"

	cmd := run.Shell(dir, pm, "exec", "prettier", "--check", ".")
	if pm == "npm" {
		cmd = run.Shell(dir, "npx", "prettier", "--check", ".")
	}

	res := r.Run(ctx, cmd)
	if res.ExitCode != 0 {
		if strings.Contains(res.Stderr, "not found") || strings.Contains(res.Stderr, "Cannot find") {
			return taskResult(id, multiagentspec.StatusSkip, "prettier not configured")
		}
		return taskResultWithOutput(id, multiagentspec.StatusNoGo, "formatting issues found", truncateOutput(res.Stdout, 2000))
	}

	return taskResult(id, multiagentspec.StatusGo, "all files formatted")
}

func (c *TypeScriptChecker) checkTest(ctx context.Context, r run.Runner, dir string, pm string) multiagentspec.TaskResult {
	id := "ts:test"

	cmd := run.Shell(dir, pm, "test")
	if pm == "npm" {
		cmd = run.Shell(dir, "npm", "test", "--", "--passWithNoTests")
	}

	res := r.Run(ctx, cmd)
	if res.ExitCode != 0 {
		if strings.Contains(res.Stderr, "Missing script") ||
			strings.Contains(res.Stderr, "no test specified") {
			return taskResult(id, multiagentspec.StatusSkip, "no test script configured")
		}
		output := res.Stdout
		if output == "" {
			output = res.Stderr
		}
		return taskResultWithOutput(id, multiagentspec.StatusNoGo, "tests failed", truncateOutput(output, 3000))
	}

	return taskResult(id, multiagentspec.StatusGo, "all tests passed")
}

// detectPackageManager determines which package manager to use based on lock files.
func detectPackageManager(dir string) string {
	if fileExists(filepath.Join(dir, "pnpm-lock.yaml")) {
		return "pnpm"
	}
	if fileExists(filepath.Join(dir, "yarn.lock")) {
		return "yarn"
	}
	if fileExists(filepath.Join(dir, "bun.lockb")) || fileExists(filepath.Join(dir, "bun.lock")) {
		return "bun"
	}
	return "npm"
}
