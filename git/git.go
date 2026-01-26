// Package git provides a structured abstraction over git CLI commands,
// using run.Runner for all command execution.
package git

import (
	"github.com/grokify/releasekit/run"
)

// Repo represents a git repository context.
type Repo struct {
	Dir    string     // Working directory (empty = cwd)
	Runner run.Runner // Command executor (nil = default real runner)
}

// NewRepo creates a Repo with the given directory and runner.
// If runner is nil, a real command executor is used.
func NewRepo(dir string, runner run.Runner) *Repo {
	if runner == nil {
		runner = run.NewRunner()
	}
	return &Repo{Dir: dir, Runner: runner}
}
