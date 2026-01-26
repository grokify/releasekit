package run

// RunSet represents a batch of commands run together (e.g., all checks).
type RunSet struct {
	Commands []Command
	Results  []*Result
}

// Passed returns true if all results passed.
func (rs *RunSet) Passed() bool {
	for _, r := range rs.Results {
		if !r.Passed() {
			return false
		}
	}
	return true
}

// Summary returns a pass/fail/skip breakdown.
func (rs *RunSet) Summary() Summary {
	s := Summary{Total: len(rs.Results)}
	for _, r := range rs.Results {
		if r.Passed() {
			s.Passed++
		} else {
			s.Failed++
		}
	}
	return s
}

// Summary holds a pass/fail/skip breakdown.
type Summary struct {
	Total   int
	Passed  int
	Failed  int
	Skipped int
}
