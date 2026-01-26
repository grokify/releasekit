package commits

import "testing"

func TestParseConventional(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		wantNil  bool
		wantType string
		scope    string
		breaking bool
		subj     string
	}{
		{"simple feat", "feat: add login", false, "feat", "", false, "add login"},
		{"feat with scope", "feat(auth): add OAuth2", false, "feat", "auth", false, "add OAuth2"},
		{"fix", "fix: resolve crash", false, "fix", "", false, "resolve crash"},
		{"breaking", "refactor!: rename API", false, "refactor", "", true, "rename API"},
		{"breaking with scope", "feat(api)!: new endpoints", false, "feat", "api", true, "new endpoints"},
		{"docs", "docs: update README", false, "docs", "", false, "update README"},
		{"chore", "chore: update deps", false, "chore", "", false, "update deps"},
		{"case insensitive type", "FEAT: uppercase", false, "feat", "", false, "uppercase"},
		{"not conventional", "update something", true, "", "", false, ""},
		{"empty", "", true, "", "", false, ""},
		{"just colon", ":", true, "", "", false, ""},
		{"no space after colon", "feat:no space", false, "feat", "", false, "no space"},
		{"with issue ref", "fix: resolve #42", false, "fix", "", false, "resolve #42"},
		{"with PR ref", "feat: add feature (#99)", false, "feat", "", false, "add feature (#99)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc, err := ParseConventional(tt.subject)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if cc != nil {
					t.Errorf("expected nil, got %+v", cc)
				}
				return
			}
			if cc == nil {
				t.Fatal("expected non-nil result")
			}
			if cc.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", cc.Type, tt.wantType)
			}
			if cc.Scope != tt.scope {
				t.Errorf("Scope = %q, want %q", cc.Scope, tt.scope)
			}
			if cc.Breaking != tt.breaking {
				t.Errorf("Breaking = %v, want %v", cc.Breaking, tt.breaking)
			}
			if cc.Subject != tt.subj {
				t.Errorf("Subject = %q, want %q", cc.Subject, tt.subj)
			}
		})
	}
}

func TestParseConventionalIssueExtraction(t *testing.T) {
	cc, err := ParseConventional("fix: resolve #42 and #99")
	if err != nil {
		t.Fatal(err)
	}
	if len(cc.Issues) != 2 {
		t.Fatalf("Issues = %v, want 2 items", cc.Issues)
	}
	if cc.Issues[0] != 42 || cc.Issues[1] != 99 {
		t.Errorf("Issues = %v, want [42, 99]", cc.Issues)
	}
}

func TestParseConventionalPRExtraction(t *testing.T) {
	cc, err := ParseConventional("feat: add feature (#123)")
	if err != nil {
		t.Fatal(err)
	}
	if len(cc.PRs) != 1 || cc.PRs[0] != 123 {
		t.Errorf("PRs = %v, want [123]", cc.PRs)
	}
}
