package commits

import "testing"

func TestSuggestCategory(t *testing.T) {
	tests := []struct {
		name    string
		cc      *ConventionalCommit
		wantCat string
		wantNil bool
		minConf float64
	}{
		{"nil input", nil, "", true, 0},
		{"feat", &ConventionalCommit{Type: "feat"}, "Added", false, 0.9},
		{"fix", &ConventionalCommit{Type: "fix"}, "Fixed", false, 0.9},
		{"docs", &ConventionalCommit{Type: "docs"}, "Changed", false, 0.7},
		{"refactor", &ConventionalCommit{Type: "refactor"}, "Changed", false, 0.8},
		{"perf", &ConventionalCommit{Type: "perf"}, "Changed", false, 0.8},
		{"revert", &ConventionalCommit{Type: "revert"}, "Removed", false, 0.7},
		{"breaking", &ConventionalCommit{Type: "feat", Breaking: true}, "Changed", false, 1.0},
		{"unknown type", &ConventionalCommit{Type: "unknown"}, "Changed", false, 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SuggestCategory(tt.cc)
			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
				return
			}
			if result == nil {
				t.Fatal("expected non-nil")
			}
			if result.Category != tt.wantCat {
				t.Errorf("Category = %q, want %q", result.Category, tt.wantCat)
			}
			if result.Confidence < tt.minConf {
				t.Errorf("Confidence = %f, want >= %f", result.Confidence, tt.minConf)
			}
		})
	}
}
