package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/grokify/releasekit/commits"
	"github.com/grokify/releasekit/git"
	"github.com/grokify/releasekit/run"
)

var (
	commitsSince   string
	commitsUntil   string
	commitsLast    int
	commitsPath    string
	commitsNoMerge bool
)

var commitsCmd = &cobra.Command{
	Use:   "commits",
	Short: "Parse commits with conventional commit analysis",
	RunE:  runCommits,
}

func init() {
	commitsCmd.Flags().StringVar(&commitsSince, "since", "", "Start ref (tag or commit)")
	commitsCmd.Flags().StringVar(&commitsUntil, "until", "", "End ref (default: HEAD)")
	commitsCmd.Flags().IntVar(&commitsLast, "last", 0, "Limit to N most recent")
	commitsCmd.Flags().StringVar(&commitsPath, "path", "", "Filter by path")
	commitsCmd.Flags().BoolVar(&commitsNoMerge, "no-merges", false, "Exclude merge commits")
}

// CommitsOutput is the structured output for the commits command.
type CommitsOutput struct {
	Total   int              `json:"total"`
	Commits []CommitAnalysis `json:"commits"`
	Summary map[string]int   `json:"summary"`
}

// CommitAnalysis represents a commit with its conventional commit analysis.
type CommitAnalysis struct {
	Hash       string  `json:"hash"`
	Subject    string  `json:"subject"`
	Author     string  `json:"author"`
	Type       string  `json:"type,omitempty"`
	Scope      string  `json:"scope,omitempty"`
	Breaking   bool    `json:"breaking,omitempty"`
	Category   string  `json:"category,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

func runCommits(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	runner := run.NewRunner()
	repo := git.NewRepo(dir, runner)

	opts := git.LogOptions{
		Since:    commitsSince,
		Until:    commitsUntil,
		Last:     commitsLast,
		Path:     commitsPath,
		NoMerges: commitsNoMerge,
	}

	logCommits, err := repo.Log(opts)
	if err != nil {
		return fmt.Errorf("getting commits: %w", err)
	}

	out := CommitsOutput{
		Total:   len(logCommits),
		Summary: make(map[string]int),
	}

	for _, c := range logCommits {
		entry := CommitAnalysis{
			Hash:    c.ShortHash,
			Subject: c.Subject,
			Author:  c.Author,
		}

		cc, _ := commits.ParseConventional(c.Subject)
		if cc != nil {
			entry.Type = cc.Type
			entry.Scope = cc.Scope
			entry.Breaking = cc.Breaking
			out.Summary[cc.Type]++

			suggestion := commits.SuggestCategory(cc)
			if suggestion != nil {
				entry.Category = suggestion.Category
				entry.Confidence = suggestion.Confidence
			}
		} else {
			out.Summary["other"]++
		}

		out.Commits = append(out.Commits, entry)
	}

	return printFormatted(out)
}
