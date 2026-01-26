package git

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/grokify/releasekit/run"
)

// Tag represents a git tag with metadata.
type Tag struct {
	Name   string
	Commit string
	Date   time.Time
}

// Tags returns all tags in the repository.
func (r *Repo) Tags() ([]Tag, error) {
	cmd := run.Git(r.Dir, "tag", "--format=%(refname:short)\t%(objectname:short)\t%(creatordate:iso8601)")
	result := r.Runner.Run(context.Background(), cmd)
	if !result.Passed() {
		return nil, fmt.Errorf("git tag: %s", result.Output())
	}
	return parseTags(result.Stdout), nil
}

// LatestTag returns the most recent semver tag with the given prefix.
// Tags are sorted by semver, not by date.
func (r *Repo) LatestTag(prefix string) (*Tag, error) {
	tags, err := r.Tags()
	if err != nil {
		return nil, err
	}

	var filtered []Tag
	for _, t := range tags {
		if prefix == "" || strings.HasPrefix(t.Name, prefix) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) == 0 {
		return nil, nil
	}

	sort.Slice(filtered, func(i, j int) bool {
		return compareSemver(filtered[i].Name, filtered[j].Name) > 0
	})

	return &filtered[0], nil
}

// CreateTag creates a lightweight tag at HEAD.
func (r *Repo) CreateTag(name string) error {
	cmd := run.Git(r.Dir, "tag", name)
	result := r.Runner.Run(context.Background(), cmd)
	if !result.Passed() {
		return fmt.Errorf("git tag %s: %s", name, result.Output())
	}
	return nil
}

// PushTag pushes a tag to the origin remote.
func (r *Repo) PushTag(name string) error {
	cmd := run.Git(r.Dir, "push", "origin", name)
	result := r.Runner.Run(context.Background(), cmd)
	if !result.Passed() {
		return fmt.Errorf("git push origin %s: %s", name, result.Output())
	}
	return nil
}

func parseTags(output string) []Tag {
	var tags []Tag
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		tag := Tag{Name: parts[0]}
		if len(parts) >= 2 {
			tag.Commit = parts[1]
		}
		if len(parts) >= 3 {
			tag.Date, _ = time.Parse("2006-01-02 15:04:05 -0700", parts[2])
		}
		tags = append(tags, tag)
	}
	return tags
}

// compareSemver compares two version strings by semver.
// Returns positive if a > b, negative if a < b, 0 if equal.
func compareSemver(a, b string) int {
	aParts := splitVersion(a)
	bParts := splitVersion(b)

	for i := 0; i < 3; i++ {
		var av, bv int
		if i < len(aParts) {
			av, _ = strconv.Atoi(aParts[i])
		}
		if i < len(bParts) {
			bv, _ = strconv.Atoi(bParts[i])
		}
		if av != bv {
			return av - bv
		}
	}
	return 0
}

func splitVersion(v string) []string {
	// Remove any prefix path (e.g., sdk/go/v1.2.3 -> v1.2.3)
	if idx := strings.LastIndex(v, "/v"); idx >= 0 {
		v = v[idx+1:]
	}
	v = strings.TrimPrefix(v, "v")
	// Remove pre-release suffix
	if idx := strings.Index(v, "-"); idx >= 0 {
		v = v[:idx]
	}
	return strings.Split(v, ".")
}
