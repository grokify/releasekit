// Package checks provides a pluggable validation framework for pre-release
// checks. All checkers use run.Runner for command execution, enabling
// dry-run previews and audit trails. Results use the multi-agent-spec
// TaskResult format for interoperability with agent workflows.
package checks

import (
	"context"
	"os"
	"os/exec"

	multiagentspec "github.com/agentplexus/multi-agent-spec/sdk/go"

	"github.com/grokify/releasekit/run"
)

// Checker runs validation checks for a language/area.
type Checker interface {
	Name() string
	Check(ctx context.Context, r run.Runner, dir string) []multiagentspec.TaskResult
}

// Options configures which checks to run.
type Options struct {
	Build    bool `json:"build"`
	Test     bool `json:"test"`
	Lint     bool `json:"lint"`
	Format   bool `json:"format"`
	ModTidy  bool `json:"mod_tidy"`
	Coverage bool `json:"coverage"`
	Verbose  bool `json:"verbose"`

	GoExcludeCoverage string `json:"go_exclude_coverage,omitempty"`
}

// DefaultOptions returns the default check options with all standard checks enabled.
func DefaultOptions() Options {
	return Options{
		Build:             true,
		Test:              true,
		Lint:              true,
		Format:            true,
		ModTidy:           true,
		Coverage:          false,
		GoExcludeCoverage: "cmd",
	}
}

// taskResult is a convenience constructor for multiagentspec.TaskResult.
func taskResult(id string, status multiagentspec.Status, detail string) multiagentspec.TaskResult {
	return multiagentspec.TaskResult{
		ID:     id,
		Status: status,
		Detail: detail,
	}
}

// taskResultWithOutput creates a TaskResult with extended output in metadata.
func taskResultWithOutput(id string, status multiagentspec.Status, detail, output string) multiagentspec.TaskResult {
	r := multiagentspec.TaskResult{
		ID:     id,
		Status: status,
		Detail: detail,
	}
	if output != "" {
		r.Metadata = map[string]interface{}{
			"output": output,
		}
	}
	return r
}

// AllGo returns true if no tasks have NO-GO status.
func AllGo(tasks []multiagentspec.TaskResult) bool {
	for _, t := range tasks {
		if t.Status == multiagentspec.StatusNoGo {
			return false
		}
	}
	return true
}

// commandExists checks if a command is available in PATH.
func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
