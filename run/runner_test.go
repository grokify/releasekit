package run

import (
	"context"
	"testing"
	"time"
)

func TestRealRunnerEcho(t *testing.T) {
	r := NewRunner()
	result := r.Run(context.Background(), Command{
		Name: "echo test",
		Args: []string{"echo", "hello"},
	})
	if !result.Passed() {
		t.Fatalf("expected pass, got exit=%d err=%v", result.ExitCode, result.Err)
	}
	if result.Stdout != "hello\n" {
		t.Errorf("stdout = %q, want %q", result.Stdout, "hello\n")
	}
}

func TestRealRunnerNonZeroExit(t *testing.T) {
	r := NewRunner()
	result := r.Run(context.Background(), Command{
		Name: "false",
		Args: []string{"false"},
	})
	if result.Passed() {
		t.Fatal("expected failure")
	}
	if result.ExitCode == 0 {
		t.Error("expected non-zero exit code")
	}
}

func TestRealRunnerTimeout(t *testing.T) {
	r := NewRunner()
	result := r.Run(context.Background(), Command{
		Name:    "sleep",
		Args:    []string{"sleep", "10"},
		Timeout: 50 * time.Millisecond,
	})
	if result.Passed() {
		t.Fatal("expected timeout failure")
	}
}

func TestRealRunnerEmptyArgs(t *testing.T) {
	r := NewRunner()
	result := r.Run(context.Background(), Command{
		Name: "empty",
	})
	if result.Passed() {
		t.Fatal("expected failure for empty args")
	}
}

func TestDryRunnerRecords(t *testing.T) {
	d := NewDryRunner()
	cmd1 := Command{Name: "git status", Args: []string{"git", "status"}}
	cmd2 := Command{Name: "go build", Args: []string{"go", "build", "./..."}}

	r1 := d.Run(context.Background(), cmd1)
	r2 := d.Run(context.Background(), cmd2)

	if !r1.Passed() || !r2.Passed() {
		t.Error("dry runner should always return passed")
	}
	if len(d.Recorded) != 2 {
		t.Fatalf("recorded %d commands, want 2", len(d.Recorded))
	}
	if d.Recorded[0].Name != "git status" {
		t.Errorf("first recorded = %q, want %q", d.Recorded[0].Name, "git status")
	}
}

func TestRecordingRunnerRecords(t *testing.T) {
	inner := NewRunner()
	rec := NewRecordingRunner(inner)

	cmd := Command{Name: "echo", Args: []string{"echo", "recorded"}}
	result := rec.Run(context.Background(), cmd)

	if !result.Passed() {
		t.Fatal("expected pass")
	}
	if len(rec.History) != 1 {
		t.Fatalf("history len = %d, want 1", len(rec.History))
	}
	if rec.History[0].Result.Stdout != "recorded\n" {
		t.Errorf("recorded stdout = %q, want %q", rec.History[0].Result.Stdout, "recorded\n")
	}
}
