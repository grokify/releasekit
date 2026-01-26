package commits

import (
	"fmt"
	"strconv"
	"strings"
)

// SuggestNextVersion computes the next semver based on conventional commits
// since the last tag. Rules:
//   - BREAKING CHANGE → major bump
//   - feat → minor bump
//   - fix, etc. → patch bump
func SuggestNextVersion(current string, commits []*ConventionalCommit) string {
	major, minor, patch := parseSemver(current)

	bump := "patch"
	for _, cc := range commits {
		if cc == nil {
			continue
		}
		if cc.Breaking {
			bump = "major"
			break
		}
		if cc.Type == "feat" {
			bump = "minor"
		}
	}

	switch bump {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	}

	return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
}

func parseSemver(version string) (major, minor, patch int) {
	version = strings.TrimPrefix(version, "v")
	parts := strings.SplitN(version, ".", 3)

	if len(parts) >= 1 {
		major, _ = strconv.Atoi(parts[0])
	}
	if len(parts) >= 2 {
		minor, _ = strconv.Atoi(parts[1])
	}
	if len(parts) >= 3 {
		// Handle pre-release suffixes like 1.2.3-beta
		patchStr := strings.SplitN(parts[2], "-", 2)[0]
		patch, _ = strconv.Atoi(patchStr)
	}
	return
}
