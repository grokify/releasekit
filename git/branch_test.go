package git

import "testing"

func TestParseAheadBehind(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		ahead, behind int
		wantErr       bool
	}{
		{"normal", "3\t5\n", 3, 5, false},
		{"zeros", "0\t0\n", 0, 0, false},
		{"spaces", "  2  4  \n", 2, 4, false},
		{"invalid", "abc\n", 0, 0, true},
		{"single value", "3\n", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ahead, behind, err := parseAheadBehind(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ahead != tt.ahead {
				t.Errorf("ahead = %d, want %d", ahead, tt.ahead)
			}
			if behind != tt.behind {
				t.Errorf("behind = %d, want %d", behind, tt.behind)
			}
		})
	}
}
