package commits

import "testing"

func TestIsBreakingChange(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		body    string
		want    bool
	}{
		{"bang suffix", "feat!: new API", "", true},
		{"bang with scope", "feat(api)!: new endpoints", "", true},
		{"body footer", "feat: new API", "BREAKING CHANGE: removed old endpoints", true},
		{"body footer dash", "feat: new API", "BREAKING-CHANGE: removed old endpoints", true},
		{"not breaking", "feat: add feature", "", false},
		{"not breaking with body", "fix: patch", "some details", false},
		{"non-conventional", "update something", "", false},
		{"breaking in multiline body", "feat: update", "some context\n\nBREAKING CHANGE: this breaks things", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBreakingChange(tt.subject, tt.body)
			if got != tt.want {
				t.Errorf("IsBreakingChange(%q, %q) = %v, want %v", tt.subject, tt.body, got, tt.want)
			}
		})
	}
}
