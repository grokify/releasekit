package run

import (
	"context"
	"testing"
)

func TestRunAllSuccess(t *testing.T) {
	r := NewRunner()
	cmds := []Command{
		{Name: "echo 1", Args: []string{"echo", "1"}},
		{Name: "echo 2", Args: []string{"echo", "2"}},
	}
	rs := RunAll(context.Background(), r, cmds, true)
	if !rs.Passed() {
		t.Error("expected all passed")
	}
	if len(rs.Results) != 2 {
		t.Fatalf("results = %d, want 2", len(rs.Results))
	}
}

func TestRunAllStopOnFail(t *testing.T) {
	r := NewRunner()
	cmds := []Command{
		{Name: "false", Args: []string{"false"}},
		{Name: "echo after", Args: []string{"echo", "after"}},
	}
	rs := RunAll(context.Background(), r, cmds, true)
	if rs.Passed() {
		t.Error("expected failure")
	}
	if len(rs.Results) != 1 {
		t.Fatalf("results = %d, want 1 (stopped on fail)", len(rs.Results))
	}
}

func TestRunAllContinueOnFail(t *testing.T) {
	r := NewRunner()
	cmds := []Command{
		{Name: "false", Args: []string{"false"}},
		{Name: "echo after", Args: []string{"echo", "after"}},
	}
	rs := RunAll(context.Background(), r, cmds, false)
	if rs.Passed() {
		t.Error("expected failure")
	}
	if len(rs.Results) != 2 {
		t.Fatalf("results = %d, want 2 (continued past fail)", len(rs.Results))
	}
}

func TestRunSetSummary(t *testing.T) {
	rs := &RunSet{
		Results: []*Result{
			{ExitCode: 0},
			{ExitCode: 1},
			{ExitCode: 0},
		},
	}
	s := rs.Summary()
	if s.Total != 3 {
		t.Errorf("Total = %d, want 3", s.Total)
	}
	if s.Passed != 2 {
		t.Errorf("Passed = %d, want 2", s.Passed)
	}
	if s.Failed != 1 {
		t.Errorf("Failed = %d, want 1", s.Failed)
	}
}
