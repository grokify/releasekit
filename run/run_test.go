package run

import "testing"

func TestResultPassed(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected bool
	}{
		{"zero exit", Result{ExitCode: 0}, true},
		{"non-zero exit", Result{ExitCode: 1}, false},
		{"error set", Result{ExitCode: 0, Err: exec_ErrNotFound}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.Passed(); got != tt.expected {
				t.Errorf("Passed() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestResultOutput(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected string
	}{
		{"stdout only", Result{Stdout: "out"}, "out"},
		{"stderr only", Result{Stderr: "err"}, "err"},
		{"both prefers stdout", Result{Stdout: "out", Stderr: "err"}, "out"},
		{"both empty", Result{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.Output(); got != tt.expected {
				t.Errorf("Output() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// sentinel error for testing
var exec_ErrNotFound = errSentinel("not found")

type errSentinel string

func (e errSentinel) Error() string { return string(e) }
