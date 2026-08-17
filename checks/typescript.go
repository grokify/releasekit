package checks

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	multiagentspec "github.com/plexusone/multi-agent-spec/sdk/go"

	"github.com/grokify/releasekit/detect"
	"github.com/grokify/releasekit/run"
)

// TypeScriptChecker validates TypeScript/JavaScript projects before release.
type TypeScriptChecker struct {
	Opts Options

	// Language distinguishes a real TypeScript project (has tsconfig.json,
	// per detect.Detect) from a plain JavaScript one. Typecheck only makes
	// sense for the former — for a JavaScript-only directory, tsc has
	// nothing to check and shouldn't be invoked at all.
	Language detect.Language
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

	if c.Language != detect.TypeScript {
		return taskResult(id, multiagentspec.StatusSkip, "no tsconfig.json — not a TypeScript project")
	}

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

	if !hasESLintConfig(dir) {
		return taskResult(id, multiagentspec.StatusSkip, "eslint not configured")
	}

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

	if !hasPrettierConfig(dir) {
		return taskResult(id, multiagentspec.StatusSkip, "prettier not configured")
	}

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

// eslintConfigFiles are the file-based ESLint config names, flat config
// (v9+) first since that's what any newly-created project should use.
var eslintConfigFiles = []string{
	"eslint.config.js", "eslint.config.mjs", "eslint.config.cjs", "eslint.config.ts",
	".eslintrc.js", ".eslintrc.cjs", ".eslintrc.json", ".eslintrc.yml", ".eslintrc.yaml", ".eslintrc",
}

// hasESLintConfig reports whether dir has an ESLint config, checked
// up front rather than inferred from a failed run's stderr: npx auto-fetches
// an ad-hoc eslint/prettier from the registry when the tool isn't a local
// devDependency, so an unconfigured project still "runs" and fails on a
// missing-config error rather than a command-not-found one.
func hasESLintConfig(dir string) bool {
	for _, f := range eslintConfigFiles {
		if fileExists(filepath.Join(dir, f)) {
			return true
		}
	}
	return packageJSONHasKey(dir, "eslintConfig")
}

// prettierConfigFiles are the file-based Prettier config names.
var prettierConfigFiles = []string{
	".prettierrc", ".prettierrc.json", ".prettierrc.yml", ".prettierrc.yaml",
	".prettierrc.js", ".prettierrc.cjs", ".prettierrc.mjs",
	"prettier.config.js", "prettier.config.cjs", "prettier.config.mjs",
}

// hasPrettierConfig reports whether dir has a Prettier config. See
// hasESLintConfig for why this is checked up front instead of inferred.
func hasPrettierConfig(dir string) bool {
	for _, f := range prettierConfigFiles {
		if fileExists(filepath.Join(dir, f)) {
			return true
		}
	}
	return packageJSONHasKey(dir, "prettier")
}

// packageJSONHasKey reports whether dir/package.json has a top-level key
// (e.g. "eslintConfig" or "prettier") — some projects embed tool config
// there instead of a separate file. A missing or unparseable package.json
// is treated as not having the key, not as an error: this is an existence
// check, and neither case should abort validation.
func packageJSONHasKey(dir, key string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return false
	}
	var pkg map[string]any
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}
	_, ok := pkg[key]
	return ok
}
