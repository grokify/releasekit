package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/grokify/releasekit/git"
	"github.com/grokify/releasekit/run"
)

var modulesCmd = &cobra.Command{
	Use:   "modules",
	Short: "Show status of Go submodules in a repo",
	RunE:  runModules,
}

// ModulesOutput is the structured output for the modules command.
type ModulesOutput struct {
	Root    string         `json:"root"`
	Modules []ModuleStatus `json:"modules"`
}

// ModuleStatus represents a Go module's release status.
type ModuleStatus struct {
	Path         string `json:"path"`
	LatestTag    string `json:"latest_tag,omitempty"`
	CommitsSince int    `json:"commits_since"`
}

func runModules(cmd *cobra.Command, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolving path: %w", err)
	}

	// Find all go.mod files
	modPaths, err := findGoModFiles(absDir)
	if err != nil {
		return fmt.Errorf("finding go.mod files: %w", err)
	}

	runner := run.NewRunner()
	repo := git.NewRepo(absDir, runner)

	out := ModulesOutput{Root: absDir}

	for _, modPath := range modPaths {
		relPath, _ := filepath.Rel(absDir, filepath.Dir(modPath))
		if relPath == "." {
			relPath = ""
		}

		ms := ModuleStatus{Path: relPath}

		// Determine tag prefix for this module
		prefix := ""
		if relPath != "" {
			prefix = relPath + "/v"
		}

		tag, err := repo.LatestTag(prefix)
		if err == nil && tag != nil {
			ms.LatestTag = tag.Name

			var logCommits []git.CommitInfo
			if relPath != "" {
				logCommits, _ = repo.CommitsSincePath(tag.Name, relPath)
			} else {
				logCommits, _ = repo.CommitsSince(tag.Name)
			}
			ms.CommitsSince = len(logCommits)
		}

		out.Modules = append(out.Modules, ms)
	}

	return printFormatted(out)
}

func findGoModFiles(root string) ([]string, error) {
	var mods []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Name() == "go.mod" {
			mods = append(mods, path)
		}
		return nil
	})
	return mods, err
}
