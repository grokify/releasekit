// Package run provides a unified command execution layer for all external
// commands (git, go, ls, gh, linters, etc.). All packages that run external
// processes use this layer rather than calling os/exec directly.
package run

import "time"

// Command represents any executable command for release analysis.
type Command struct {
	Name    string        // Human-readable label (e.g., "go build", "git status")
	Args    []string      // Full command and arguments (e.g., ["git", "status", "--porcelain"])
	Dir     string        // Working directory (empty = cwd)
	Env     []string      // Additional environment variables (KEY=VALUE)
	Timeout time.Duration // Per-command timeout (0 = use context deadline)
}

// Result captures the outcome of running a command.
type Result struct {
	Command  Command
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
	Err      error // Non-nil if command failed to start or timed out
}

// Passed returns true if the command exited with code 0 and no error.
func (r *Result) Passed() bool {
	return r.Err == nil && r.ExitCode == 0
}

// Output returns Stdout if non-empty, otherwise Stderr.
func (r *Result) Output() string {
	if r.Stdout != "" {
		return r.Stdout
	}
	return r.Stderr
}
