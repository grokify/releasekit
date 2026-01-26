package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	multiagentspec "github.com/agentplexus/multi-agent-spec/sdk/go"
	"github.com/spf13/cobra"

	"github.com/grokify/releasekit/checks"
	"github.com/grokify/releasekit/detect"
	"github.com/grokify/releasekit/run"
)

var (
	validateDryRun   bool
	validateVerbose  bool
	validateCoverage bool
	validateNoLint   bool
	validateNoTest   bool
)

var validateCmd = &cobra.Command{
	Use:   "validate [dir]",
	Short: "Run pre-release checks (auto-detects language)",
	Long: `Validate runs language-specific pre-release checks against a project directory.
It auto-detects languages (Go, TypeScript, JavaScript, Python, Rust) and runs
appropriate checks: build, test, lint, format, and more.

Output conforms to the multi-agent-spec AgentResult schema.

Exit codes:
  0 - all checks passed (GO)
  1 - error running checks
  2 - one or more checks failed (NO-GO)`,
	RunE: runValidate,
}

func init() {
	validateCmd.Flags().BoolVar(&validateDryRun, "dry-run", false, "Show what would be run without executing")
	validateCmd.Flags().BoolVar(&validateVerbose, "verbose", false, "Show verbose output")
	validateCmd.Flags().BoolVar(&validateCoverage, "coverage", false, "Include coverage check")
	validateCmd.Flags().BoolVar(&validateNoLint, "no-lint", false, "Skip lint checks")
	validateCmd.Flags().BoolVar(&validateNoTest, "no-test", false, "Skip test checks")
}

func runValidate(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	// Detect languages
	detections, err := detect.Detect(absDir)
	if err != nil {
		return fmt.Errorf("detecting languages: %w", err)
	}

	if len(detections) == 0 {
		return fmt.Errorf("no supported languages detected in %s", absDir)
	}

	// Build options
	opts := checks.DefaultOptions()
	opts.Verbose = validateVerbose
	opts.Coverage = validateCoverage
	if validateNoLint {
		opts.Lint = false
	}
	if validateNoTest {
		opts.Test = false
	}

	// Build runner
	var runner run.Runner
	if validateDryRun {
		runner = run.NewDryRunner()
	} else {
		runner = run.NewRunner()
	}

	ctx := context.Background()
	startTime := time.Now()
	var allTasks []multiagentspec.TaskResult

	// Run checks for each detected language/path
	for _, d := range detections {
		checkDir := absDir
		if d.Path != "" {
			checkDir = filepath.Join(absDir, d.Path)
		}

		var checker checks.Checker
		switch d.Language {
		case detect.Go:
			checker = &checks.GoChecker{Opts: opts}
		case detect.TypeScript, detect.JavaScript:
			checker = &checks.TypeScriptChecker{Opts: opts}
		default:
			continue
		}

		results := checker.Check(ctx, runner, checkDir)

		// Prefix task IDs with path context if multi-module
		if d.Path != "" {
			for i := range results {
				results[i].ID = d.Path + "/" + results[i].ID
			}
		}

		allTasks = append(allTasks, results...)
	}

	duration := time.Since(startTime)

	// Build AgentResult
	agentResult := multiagentspec.AgentResult{
		Schema:     "https://raw.githubusercontent.com/agentplexus/multi-agent-spec/main/schema/report/agent-result.schema.json",
		AgentID:    "language-validation",
		StepID:     "releasekit-validate",
		Tasks:      allTasks,
		ExecutedAt: time.Now().UTC(),
		Duration:   duration.Round(time.Millisecond).String(),
	}

	if validateDryRun {
		agentResult.Outputs = map[string]interface{}{
			"dry_run": true,
			"dir":     absDir,
		}
	}

	agentResult.Status = agentResult.ComputeStatus()

	if err := printFormatted(agentResult); err != nil {
		return err
	}

	if agentResult.Status == multiagentspec.StatusNoGo {
		os.Exit(2)
	}

	return nil
}
