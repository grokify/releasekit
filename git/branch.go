package git

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/grokify/releasekit/run"
)

// BranchAheadBehind returns how many commits the current branch is ahead
// and behind its upstream tracking branch.
func (r *Repo) BranchAheadBehind() (ahead, behind int, err error) {
	// Get current branch tracking ref
	branchCmd := run.Git(r.Dir, "rev-parse", "--abbrev-ref", "HEAD")
	branchResult := r.Runner.Run(context.Background(), branchCmd)
	if !branchResult.Passed() {
		return 0, 0, fmt.Errorf("git rev-parse HEAD: %s", branchResult.Output())
	}
	branch := strings.TrimSpace(branchResult.Stdout)

	// Get upstream for this branch
	upstreamCmd := run.Git(r.Dir, "rev-parse", "--abbrev-ref", branch+"@{upstream}")
	upstreamResult := r.Runner.Run(context.Background(), upstreamCmd)
	if !upstreamResult.Passed() {
		return 0, 0, fmt.Errorf("no upstream configured for %s", branch)
	}
	upstream := strings.TrimSpace(upstreamResult.Stdout)

	// Count ahead/behind
	countCmd := run.Git(r.Dir, "rev-list", "--left-right", "--count", branch+"..."+upstream)
	countResult := r.Runner.Run(context.Background(), countCmd)
	if !countResult.Passed() {
		return 0, 0, fmt.Errorf("git rev-list: %s", countResult.Output())
	}

	return parseAheadBehind(countResult.Stdout)
}

func parseAheadBehind(output string) (ahead, behind int, err error) {
	parts := strings.Fields(strings.TrimSpace(output))
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected rev-list output: %q", output)
	}
	ahead, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("parse ahead count: %w", err)
	}
	behind, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("parse behind count: %w", err)
	}
	return ahead, behind, nil
}
