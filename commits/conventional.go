// Package commits provides conventional commit parsing and changelog category
// suggestion for git commit messages.
package commits

import (
	"regexp"
	"strings"
)

// ConventionalCommit represents a parsed conventional commit message.
type ConventionalCommit struct {
	Type     string // feat, fix, docs, etc.
	Scope    string // optional scope
	Subject  string // description after the colon
	Breaking bool   // breaking change flag (! or BREAKING CHANGE footer)
	Issues   []int  // referenced issues
	PRs      []int  // referenced PRs
}

// conventionalRe matches: type(scope)!: description
// Groups: 1=type, 2=scope (with parens), 3=scope (without parens), 4=breaking, 5=subject
var conventionalRe = regexp.MustCompile(`^(\w+)(\(([^)]*)\))?(!)?\s*:\s*(.+)$`)

// ParseConventional parses a commit subject line into a ConventionalCommit.
// Returns (nil, nil) if the subject does not match conventional commit format.
func ParseConventional(subject string) (*ConventionalCommit, error) {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return nil, nil
	}

	matches := conventionalRe.FindStringSubmatch(subject)
	if matches == nil {
		return nil, nil
	}

	cc := &ConventionalCommit{
		Type:     strings.ToLower(matches[1]),
		Scope:    matches[3],
		Breaking: matches[4] == "!",
		Subject:  strings.TrimSpace(matches[5]),
	}

	cc.Issues = ExtractIssueRefs(cc.Subject)
	cc.PRs = ExtractPRRefs(cc.Subject)

	return cc, nil
}
