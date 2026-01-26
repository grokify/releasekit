package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/grokify/releasekit/git"
	"github.com/grokify/releasekit/run"
)

var changelogPath string

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate version consistency between tags and CHANGELOG.json",
	RunE:  runCheck,
}

func init() {
	checkCmd.Flags().StringVar(&changelogPath, "changelog", "CHANGELOG.json", "Path to CHANGELOG.json")
}

// CheckOutput is the structured output for the check command.
type CheckOutput struct {
	Status       string       `json:"status"` // pass, fail
	TagCount     int          `json:"tag_count"`
	ChangelogVer int          `json:"changelog_versions"`
	Issues       []CheckIssue `json:"issues,omitempty"`
}

// CheckIssue represents a validation issue.
type CheckIssue struct {
	Type    string `json:"type"` // missing_changelog, missing_tag, invalid_version
	Version string `json:"version"`
	Message string `json:"message"`
}

func runCheck(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	runner := run.NewRunner()
	repo := git.NewRepo(dir, runner)

	// Get tags
	tags, err := repo.Tags()
	if err != nil {
		return fmt.Errorf("getting tags: %w", err)
	}

	tagSet := make(map[string]bool)
	for _, t := range tags {
		tagSet[t.Name] = true
	}

	// Read changelog if it exists
	changelogVersions, err := readChangelogVersions(changelogPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading changelog: %w", err)
	}

	out := CheckOutput{
		Status:       "pass",
		TagCount:     len(tags),
		ChangelogVer: len(changelogVersions),
	}

	// Check for tags without changelog entries
	for _, t := range tags {
		if len(changelogVersions) > 0 && !changelogVersions[t.Name] {
			out.Issues = append(out.Issues, CheckIssue{
				Type:    "missing_changelog",
				Version: t.Name,
				Message: fmt.Sprintf("tag %s has no CHANGELOG.json entry", t.Name),
			})
		}
	}

	// Check for changelog versions without tags
	for v := range changelogVersions {
		if !tagSet[v] && !tagSet["v"+v] {
			out.Issues = append(out.Issues, CheckIssue{
				Type:    "missing_tag",
				Version: v,
				Message: fmt.Sprintf("CHANGELOG.json version %s has no corresponding tag", v),
			})
		}
	}

	if len(out.Issues) > 0 {
		out.Status = "fail"
	}

	return printFormatted(out)
}

// readChangelogVersions reads a CHANGELOG.json and extracts version strings.
// Returns a map of version -> true for quick lookup.
func readChangelogVersions(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Minimal struct to extract versions from structured-changelog format
	var changelog struct {
		VersionHistory []struct {
			Version string `json:"version"`
		} `json:"version_history"`
	}

	if err := json.Unmarshal(data, &changelog); err != nil {
		return nil, fmt.Errorf("parsing CHANGELOG.json: %w", err)
	}

	versions := make(map[string]bool)
	for _, v := range changelog.VersionHistory {
		versions[v.Version] = true
		versions["v"+v.Version] = true
	}
	return versions, nil
}
