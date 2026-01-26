package commits

import "strings"

// IsBreakingChange returns true if the commit indicates a breaking change,
// either via the ! suffix in the subject or a BREAKING CHANGE footer in the body.
func IsBreakingChange(subject, body string) bool {
	// Check for ! before the colon in conventional commit format
	if conventionalRe.MatchString(subject) {
		matches := conventionalRe.FindStringSubmatch(subject)
		if matches != nil && matches[4] == "!" {
			return true
		}
	}

	// Check for BREAKING CHANGE or BREAKING-CHANGE in body
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "BREAKING CHANGE:") ||
			strings.HasPrefix(trimmed, "BREAKING CHANGE #") ||
			strings.HasPrefix(trimmed, "BREAKING-CHANGE:") {
			return true
		}
	}

	return false
}
