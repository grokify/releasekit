package commits

import "testing"

func TestSuggestNextVersion(t *testing.T) {
	tests := []struct {
		name    string
		current string
		commits []*ConventionalCommit
		want    string
	}{
		{
			"patch bump",
			"v1.2.3",
			[]*ConventionalCommit{{Type: "fix"}},
			"v1.2.4",
		},
		{
			"minor bump",
			"v1.2.3",
			[]*ConventionalCommit{{Type: "feat"}},
			"v1.3.0",
		},
		{
			"major bump",
			"v1.2.3",
			[]*ConventionalCommit{{Type: "feat", Breaking: true}},
			"v2.0.0",
		},
		{
			"major wins over minor",
			"v1.2.3",
			[]*ConventionalCommit{
				{Type: "feat"},
				{Type: "fix", Breaking: true},
			},
			"v2.0.0",
		},
		{
			"minor wins over patch",
			"v1.2.3",
			[]*ConventionalCommit{
				{Type: "fix"},
				{Type: "feat"},
				{Type: "docs"},
			},
			"v1.3.0",
		},
		{
			"from zero",
			"v0.0.0",
			[]*ConventionalCommit{{Type: "feat"}},
			"v0.1.0",
		},
		{
			"no v prefix",
			"1.2.3",
			[]*ConventionalCommit{{Type: "fix"}},
			"v1.2.4",
		},
		{
			"nil commits skipped",
			"v1.0.0",
			[]*ConventionalCommit{nil, {Type: "fix"}},
			"v1.0.1",
		},
		{
			"empty commits",
			"v1.0.0",
			nil,
			"v1.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SuggestNextVersion(tt.current, tt.commits)
			if got != tt.want {
				t.Errorf("SuggestNextVersion(%q) = %q, want %q", tt.current, got, tt.want)
			}
		})
	}
}

func TestParseSemver(t *testing.T) {
	tests := []struct {
		input               string
		major, minor, patch int
	}{
		{"v1.2.3", 1, 2, 3},
		{"0.1.0", 0, 1, 0},
		{"v2.0.0-beta", 2, 0, 0},
		{"1", 1, 0, 0},
		{"1.2", 1, 2, 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			major, minor, patch := parseSemver(tt.input)
			if major != tt.major || minor != tt.minor || patch != tt.patch {
				t.Errorf("parseSemver(%q) = %d.%d.%d, want %d.%d.%d",
					tt.input, major, minor, patch, tt.major, tt.minor, tt.patch)
			}
		})
	}
}
