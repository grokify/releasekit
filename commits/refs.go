package commits

import (
	"regexp"
	"strconv"
)

// issueRe matches issue references like #123 that are NOT inside parentheses.
// It uses a negative lookbehind approximation by matching start-of-string or non-paren char.
var issueRe = regexp.MustCompile(`(?:^|[^(])#(\d+)`)

// prRe matches PR references like (#123) — issue refs inside parentheses.
var prRe = regexp.MustCompile(`\(#(\d+)\)`)

// ExtractIssueRefs extracts issue references (#123) from text.
// References inside parentheses like (#123) are treated as PR refs, not issues.
func ExtractIssueRefs(text string) []int {
	// First, remove PR-style refs so they don't get double-counted
	cleaned := prRe.ReplaceAllString(text, "")
	matches := issueRe.FindAllStringSubmatch(cleaned, -1)
	return extractInts(matches)
}

// ExtractPRRefs extracts PR references (#123) that appear inside parentheses.
func ExtractPRRefs(text string) []int {
	matches := prRe.FindAllStringSubmatch(text, -1)
	return extractInts(matches)
}

func extractInts(matches [][]string) []int {
	if len(matches) == 0 {
		return nil
	}
	result := make([]int, 0, len(matches))
	for _, m := range matches {
		if len(m) >= 2 {
			n, err := strconv.Atoi(m[1])
			if err == nil {
				result = append(result, n)
			}
		}
	}
	return result
}
