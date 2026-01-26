package git

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/grokify/releasekit/run"
)

const (
	commitDelimiter = "---COMMIT_DELIMITER---"
	bodyEnd         = "---END_BODY---"
)

// gitLogFormat is the format string for git log output.
// Subject is on the first tab-delimited line; body follows on subsequent lines.
var gitLogFormat = fmt.Sprintf("%%H\t%%h\t%%an\t%%ae\t%%aI\t%%s%%n%%b%s%s", bodyEnd, commitDelimiter)

// CommitInfo represents a parsed git commit.
type CommitInfo struct {
	Hash      string
	ShortHash string
	Author    string
	Email     string
	Date      time.Time
	Subject   string
	Body      string
	Files     []FileStat
}

// FileStat represents file-level change statistics.
type FileStat struct {
	Path       string
	Insertions int
	Deletions  int
}

// LogOptions configures git log queries.
type LogOptions struct {
	Since    string // Start ref (tag or commit)
	Until    string // End ref (default: HEAD)
	Path     string // Filter by path
	Last     int    // Limit to N most recent
	NoMerges bool   // Exclude merge commits
	NumStat  bool   // Include file statistics
}

// Log returns commits matching the given options.
func (r *Repo) Log(opts LogOptions) ([]CommitInfo, error) {
	args := []string{"log", "--format=" + gitLogFormat}

	if opts.NoMerges {
		args = append(args, "--no-merges")
	}
	if opts.Last > 0 {
		args = append(args, fmt.Sprintf("-%d", opts.Last))
	}
	if opts.NumStat {
		args = append(args, "--numstat")
	}
	if opts.Since != "" {
		ref := opts.Since
		if opts.Until != "" {
			ref += ".." + opts.Until
		} else {
			ref += "..HEAD"
		}
		args = append(args, ref)
	}
	if opts.Path != "" {
		args = append(args, "--", opts.Path)
	}

	cmd := run.Git(r.Dir, args...)
	result := r.Runner.Run(context.Background(), cmd)
	if !result.Passed() {
		return nil, fmt.Errorf("git log: %s", result.Output())
	}

	return parseLog(result.Stdout, opts.NumStat), nil
}

// CommitsSince returns all commits since the given tag.
func (r *Repo) CommitsSince(tag string) ([]CommitInfo, error) {
	return r.Log(LogOptions{Since: tag})
}

// CommitsSincePath returns commits since tag filtered by path.
func (r *Repo) CommitsSincePath(tag, path string) ([]CommitInfo, error) {
	return r.Log(LogOptions{Since: tag, Path: path})
}

func parseLog(output string, numstat bool) []CommitInfo {
	var commits []CommitInfo

	blocks := strings.Split(output, commitDelimiter)
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}

		ci := parseCommitBlock(block, numstat)
		if ci.Hash != "" {
			commits = append(commits, ci)
		}
	}
	return commits
}

func parseCommitBlock(block string, numstat bool) CommitInfo {
	var ci CommitInfo

	// Split into header+body and numstat sections
	bodyIdx := strings.Index(block, bodyEnd)
	var mainPart, statPart string
	if bodyIdx >= 0 {
		mainPart = block[:bodyIdx]
		statPart = strings.TrimSpace(block[bodyIdx+len(bodyEnd):])
	} else {
		mainPart = block
	}

	// Parse header line (first line with tabs)
	lines := strings.SplitN(mainPart, "\n", 2)
	if len(lines) == 0 {
		return ci
	}

	fields := strings.SplitN(lines[0], "\t", 6)
	if len(fields) < 6 {
		return ci
	}

	ci.Hash = fields[0]
	ci.ShortHash = fields[1]
	ci.Author = fields[2]
	ci.Email = fields[3]
	ci.Date, _ = time.Parse(time.RFC3339, fields[4])
	ci.Subject = fields[5]

	// Body is everything after the first line
	if len(lines) > 1 {
		ci.Body = strings.TrimSpace(lines[1])
	}

	// Parse numstat if present
	if numstat && statPart != "" {
		ci.Files = parseNumstat(statPart)
	}

	return ci
}

func parseNumstat(output string) []FileStat {
	var stats []FileStat
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		ins, _ := strconv.Atoi(parts[0])
		del, _ := strconv.Atoi(parts[1])
		stats = append(stats, FileStat{
			Path:       parts[2],
			Insertions: ins,
			Deletions:  del,
		})
	}
	return stats
}
