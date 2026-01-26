package run

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

// Runner executes commands. Implementations support real execution, dry-run,
// and recording for audit/replay.
type Runner interface {
	Run(ctx context.Context, cmd Command) *Result
}

// NewRunner returns a real command executor using os/exec.
func NewRunner() Runner {
	return &realRunner{}
}

type realRunner struct{}

func (r *realRunner) Run(ctx context.Context, cmd Command) *Result {
	start := time.Now()

	if cmd.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cmd.Timeout)
		defer cancel()
	}

	if len(cmd.Args) == 0 {
		return &Result{
			Command:  cmd,
			ExitCode: -1,
			Duration: time.Since(start),
			Err:      exec.ErrNotFound,
		}
	}

	c := exec.CommandContext(ctx, cmd.Args[0], cmd.Args[1:]...) //nolint:gosec // G204: intentional command execution - this is the core runner functionality
	if cmd.Dir != "" {
		c.Dir = cmd.Dir
	}
	if len(cmd.Env) > 0 {
		c.Env = append(c.Environ(), cmd.Env...)
	}

	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Run()
	duration := time.Since(start)

	result := &Result{
		Command:  cmd,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	if err != nil {
		result.Err = err
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
			result.Err = nil // ExitError is not a run failure, just non-zero exit
		} else {
			result.ExitCode = -1
		}
	}

	return result
}
