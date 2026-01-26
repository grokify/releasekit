package checks

import (
	"context"
	"testing"

	multiagentspec "github.com/agentplexus/multi-agent-spec/sdk/go"

	"github.com/grokify/releasekit/run"
)

func TestAllGo(t *testing.T) {
	passing := []multiagentspec.TaskResult{
		{ID: "a", Status: multiagentspec.StatusGo},
		{ID: "b", Status: multiagentspec.StatusWarn},
		{ID: "c", Status: multiagentspec.StatusSkip},
	}
	if !AllGo(passing) {
		t.Error("AllGo should be true when no NO-GO")
	}

	failing := []multiagentspec.TaskResult{
		{ID: "a", Status: multiagentspec.StatusGo},
		{ID: "b", Status: multiagentspec.StatusNoGo},
	}
	if AllGo(failing) {
		t.Error("AllGo should be false when NO-GO exists")
	}

	if !AllGo(nil) {
		t.Error("AllGo should be true for empty results")
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if !opts.Build {
		t.Error("Build should be true by default")
	}
	if !opts.Test {
		t.Error("Test should be true by default")
	}
	if !opts.Lint {
		t.Error("Lint should be true by default")
	}
	if !opts.Format {
		t.Error("Format should be true by default")
	}
	if !opts.ModTidy {
		t.Error("ModTidy should be true by default")
	}
	if opts.Coverage {
		t.Error("Coverage should be false by default")
	}
}

func TestGoCheckerDryRun(t *testing.T) {
	checker := &GoChecker{Opts: DefaultOptions()}
	if checker.Name() != "Go" {
		t.Errorf("Name() = %q, want %q", checker.Name(), "Go")
	}

	r := run.NewDryRunner()
	ctx := context.Background()
	results := checker.Check(ctx, r, "/tmp/fake")

	if len(results) < 5 {
		t.Errorf("expected at least 5 results, got %d", len(results))
	}

	ids := make(map[string]bool)
	for _, res := range results {
		ids[res.ID] = true
	}
	expected := []string{"go:no-local-replace", "go:build", "go:test"}
	for _, id := range expected {
		if !ids[id] {
			t.Errorf("expected check %q in results", id)
		}
	}
}

func TestTypeScriptCheckerDryRun(t *testing.T) {
	checker := &TypeScriptChecker{Opts: DefaultOptions()}
	if checker.Name() != "TypeScript" {
		t.Errorf("Name() = %q, want %q", checker.Name(), "TypeScript")
	}

	r := run.NewDryRunner()
	ctx := context.Background()
	results := checker.Check(ctx, r, "/tmp/fake")

	if len(results) != 4 {
		t.Errorf("expected 4 results, got %d", len(results))
	}

	ids := make(map[string]bool)
	for _, res := range results {
		ids[res.ID] = true
	}
	expected := []string{"ts:typecheck", "ts:lint", "ts:format", "ts:test"}
	for _, id := range expected {
		if !ids[id] {
			t.Errorf("expected check %q in results", id)
		}
	}
}

func TestDetectPackageManager(t *testing.T) {
	pm := detectPackageManager("/tmp/nonexistent-dir-xyz")
	if pm != "npm" {
		t.Errorf("detectPackageManager for empty dir = %q, want %q", pm, "npm")
	}
}

func TestTruncateOutput(t *testing.T) {
	short := "hello"
	if got := truncateOutput(short, 100); got != short {
		t.Errorf("truncateOutput short = %q, want %q", got, short)
	}

	long := "abcdefghij"
	got := truncateOutput(long, 5)
	if got != "abcde\n... (truncated)" {
		t.Errorf("truncateOutput long = %q", got)
	}
}

func TestParseCoverageFromOutput(t *testing.T) {
	input := `ok      github.com/example/pkg  0.5s  coverage: 85.3% of statements
ok      github.com/example/other  0.2s  coverage: 72.1% of statements`

	got := parseCoverageFromOutput(input)
	if got != "coverage: 72.1% of statements" {
		t.Errorf("parseCoverageFromOutput = %q", got)
	}

	if got := parseCoverageFromOutput("no coverage here"); got != "unknown" {
		t.Errorf("parseCoverageFromOutput empty = %q, want %q", got, "unknown")
	}
}

func TestTaskResultHelpers(t *testing.T) {
	r := taskResult("test:id", multiagentspec.StatusGo, "passed")
	if r.ID != "test:id" || r.Status != multiagentspec.StatusGo || r.Detail != "passed" {
		t.Errorf("taskResult mismatch: %+v", r)
	}

	rOut := taskResultWithOutput("test:id2", multiagentspec.StatusNoGo, "failed", "error details")
	if rOut.Metadata == nil {
		t.Fatal("expected metadata to be set")
	}
	if rOut.Metadata["output"] != "error details" {
		t.Errorf("expected output in metadata, got %v", rOut.Metadata["output"])
	}

	// Empty output should not set metadata
	rEmpty := taskResultWithOutput("test:id3", multiagentspec.StatusGo, "ok", "")
	if rEmpty.Metadata != nil {
		t.Errorf("expected nil metadata for empty output, got %v", rEmpty.Metadata)
	}
}
