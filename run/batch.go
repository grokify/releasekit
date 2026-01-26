package run

import "context"

// RunAll executes a slice of commands sequentially, stopping on first failure
// if stopOnFail is true.
func RunAll(ctx context.Context, r Runner, cmds []Command, stopOnFail bool) *RunSet {
	rs := &RunSet{Commands: cmds}
	for _, cmd := range cmds {
		result := r.Run(ctx, cmd)
		rs.Results = append(rs.Results, result)
		if stopOnFail && !result.Passed() {
			break
		}
	}
	return rs
}
