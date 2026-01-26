package run

import "context"

// RunRecord captures a command and its result for audit trails.
type RunRecord struct {
	Command Command
	Result  *Result
}

// RecordingRunner wraps a real runner and records all commands + results.
// Useful for audit trails and LLM-optimized output.
type RecordingRunner struct {
	Inner   Runner
	History []RunRecord
}

// NewRecordingRunner wraps a real runner and records all commands + results.
func NewRecordingRunner(inner Runner) *RecordingRunner {
	return &RecordingRunner{Inner: inner}
}

func (r *RecordingRunner) Run(ctx context.Context, cmd Command) *Result {
	result := r.Inner.Run(ctx, cmd)
	r.History = append(r.History, RunRecord{
		Command: cmd,
		Result:  result,
	})
	return result
}
