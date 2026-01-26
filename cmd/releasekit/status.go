package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grokify/releasekit/commits"
	"github.com/grokify/releasekit/git"
	"github.com/grokify/releasekit/run"
)

var statusPath string

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show release status: latest tag, commits since, working tree state",
	RunE:  runStatus,
}

func init() {
	statusCmd.Flags().StringVar(&statusPath, "path", "", "Path prefix for nested module")
}

// StatusOutput is the structured output for the status command.
type StatusOutput struct {
	LatestTag    string        `json:"latest_tag"`
	TagCommit    string        `json:"tag_commit,omitempty"`
	CommitsSince int           `json:"commits_since"`
	NextVersion  string        `json:"next_version,omitempty"`
	Commits      []CommitEntry `json:"commits,omitempty"`
	WorkingTree  WorkingTree   `json:"working_tree"`
}

// CommitEntry represents a commit in the status output.
type CommitEntry struct {
	Hash    string `json:"hash"`
	Subject string `json:"subject"`
	Type    string `json:"type,omitempty"`
	Scope   string `json:"scope,omitempty"`
}

// WorkingTree represents working tree state.
type WorkingTree struct {
	Clean     bool `json:"clean"`
	Staged    int  `json:"staged"`
	Unstaged  int  `json:"unstaged"`
	Untracked int  `json:"untracked"`
}

func runStatus(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	runner := run.NewRunner()
	repo := git.NewRepo(dir, runner)

	out := StatusOutput{}

	// Get latest tag
	tag, err := repo.LatestTag(statusPath)
	if err != nil {
		return fmt.Errorf("getting latest tag: %w", err)
	}
	if tag != nil {
		out.LatestTag = tag.Name
		out.TagCommit = tag.Commit
	}

	// Get commits since last tag
	if out.LatestTag != "" {
		var logCommits []git.CommitInfo
		if statusPath != "" {
			logCommits, err = repo.CommitsSincePath(out.LatestTag, statusPath)
		} else {
			logCommits, err = repo.CommitsSince(out.LatestTag)
		}
		if err != nil {
			return fmt.Errorf("getting commits: %w", err)
		}
		out.CommitsSince = len(logCommits)

		var parsed []*commits.ConventionalCommit
		for _, c := range logCommits {
			entry := CommitEntry{Hash: c.ShortHash, Subject: c.Subject}
			cc, _ := commits.ParseConventional(c.Subject)
			if cc != nil {
				entry.Type = cc.Type
				entry.Scope = cc.Scope
				parsed = append(parsed, cc)
			}
			out.Commits = append(out.Commits, entry)
		}

		if len(parsed) > 0 {
			out.NextVersion = commits.SuggestNextVersion(out.LatestTag, parsed)
		}
	}

	// Get working tree status
	status, err := repo.Status()
	if err != nil {
		return fmt.Errorf("getting status: %w", err)
	}
	out.WorkingTree = WorkingTree{
		Clean:     len(status.Staged) == 0 && len(status.Unstaged) == 0 && len(status.Untracked) == 0,
		Staged:    len(status.Staged),
		Unstaged:  len(status.Unstaged),
		Untracked: len(status.Untracked),
	}

	return printFormatted(out)
}
