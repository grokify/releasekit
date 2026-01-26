package run

import (
	"reflect"
	"testing"
)

func TestGitHelper(t *testing.T) {
	cmd := Git("/tmp/repo", "status", "--porcelain")
	if cmd.Dir != "/tmp/repo" {
		t.Errorf("Dir = %q, want %q", cmd.Dir, "/tmp/repo")
	}
	expected := []string{"git", "status", "--porcelain"}
	if !reflect.DeepEqual(cmd.Args, expected) {
		t.Errorf("Args = %v, want %v", cmd.Args, expected)
	}
	if cmd.Name != "git status" {
		t.Errorf("Name = %q, want %q", cmd.Name, "git status")
	}
}

func TestGoHelper(t *testing.T) {
	cmd := Go("/tmp/repo", "build", "./...")
	expected := []string{"go", "build", "./..."}
	if !reflect.DeepEqual(cmd.Args, expected) {
		t.Errorf("Args = %v, want %v", cmd.Args, expected)
	}
	if cmd.Name != "go build" {
		t.Errorf("Name = %q, want %q", cmd.Name, "go build")
	}
}

func TestShellHelper(t *testing.T) {
	cmd := Shell("/tmp", "ls", "-la")
	expected := []string{"ls", "-la"}
	if !reflect.DeepEqual(cmd.Args, expected) {
		t.Errorf("Args = %v, want %v", cmd.Args, expected)
	}
	if cmd.Name != "ls" {
		t.Errorf("Name = %q, want %q", cmd.Name, "ls")
	}
}

func TestGitHelperEmptyArgs(t *testing.T) {
	cmd := Git("/tmp/repo")
	if cmd.Name != "git " {
		t.Errorf("Name = %q, want %q", cmd.Name, "git ")
	}
}
