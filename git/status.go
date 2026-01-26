package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/grokify/releasekit/run"
)

// Status represents working tree state.
type Status struct {
	Staged    []FileStatus
	Unstaged  []FileStatus
	Untracked []string
}

// FileStatus represents a file's git status.
type FileStatus struct {
	Path   string
	Status string // M, A, D, R, C
}

// Status returns the working tree status.
func (r *Repo) Status() (*Status, error) {
	cmd := run.Git(r.Dir, "status", "--porcelain")
	result := r.Runner.Run(context.Background(), cmd)
	if !result.Passed() {
		return nil, fmt.Errorf("git status: %s", result.Output())
	}
	return parseStatus(result.Stdout), nil
}

// IsClean returns true if the working tree has no changes.
func (r *Repo) IsClean() (bool, error) {
	status, err := r.Status()
	if err != nil {
		return false, err
	}
	return len(status.Staged) == 0 &&
		len(status.Unstaged) == 0 &&
		len(status.Untracked) == 0, nil
}

func parseStatus(output string) *Status {
	s := &Status{}
	for _, line := range strings.Split(output, "\n") {
		if len(line) < 3 {
			continue
		}
		x := line[0] // index status
		y := line[1] // worktree status
		path := strings.TrimSpace(line[3:])

		if x == '?' && y == '?' {
			s.Untracked = append(s.Untracked, path)
			continue
		}

		if x != ' ' && x != '?' {
			s.Staged = append(s.Staged, FileStatus{
				Path:   path,
				Status: string(x),
			})
		}
		if y != ' ' && y != '?' {
			s.Unstaged = append(s.Unstaged, FileStatus{
				Path:   path,
				Status: string(y),
			})
		}
	}
	return s
}
