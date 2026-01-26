package run

import "context"

// DryRunner records commands without executing them.
// Useful for previewing what a workflow or check suite would do.
type DryRunner struct {
	Recorded []Command
}

// NewDryRunner returns a runner that records commands without executing them.
func NewDryRunner() *DryRunner {
	return &DryRunner{}
}

func (d *DryRunner) Run(_ context.Context, cmd Command) *Result {
	d.Recorded = append(d.Recorded, cmd)
	return &Result{
		Command:  cmd,
		ExitCode: 0,
	}
}
