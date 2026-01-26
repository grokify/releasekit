package git

import "testing"

func TestCompareSemver(t *testing.T) {
	tests := []struct {
		a, b string
		want int // positive = a > b, negative = a < b, 0 = equal
	}{
		{"v1.0.0", "v0.9.0", 1},
		{"v0.9.0", "v1.0.0", -1},
		{"v1.0.0", "v1.0.0", 0},
		{"v1.2.3", "v1.2.2", 1},
		{"v2.0.0", "v1.99.99", 1},
		{"v0.1.0", "v0.0.9", 1},
		{"sdk/go/v1.0.0", "sdk/go/v0.9.0", 1},
	}

	for _, tt := range tests {
		t.Run(tt.a+" vs "+tt.b, func(t *testing.T) {
			got := compareSemver(tt.a, tt.b)
			if (tt.want > 0 && got <= 0) || (tt.want < 0 && got >= 0) || (tt.want == 0 && got != 0) {
				t.Errorf("compareSemver(%q, %q) = %d, want sign %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestParseTags(t *testing.T) {
	input := "v1.0.0\tabc1234\t2024-01-15 10:00:00 -0800\nv0.1.0\tdef5678\t2024-01-01 09:00:00 -0800\n"
	tags := parseTags(input)
	if len(tags) != 2 {
		t.Fatalf("got %d tags, want 2", len(tags))
	}
	if tags[0].Name != "v1.0.0" {
		t.Errorf("first tag = %q, want %q", tags[0].Name, "v1.0.0")
	}
	if tags[0].Commit != "abc1234" {
		t.Errorf("first commit = %q, want %q", tags[0].Commit, "abc1234")
	}
	if tags[1].Name != "v0.1.0" {
		t.Errorf("second tag = %q, want %q", tags[1].Name, "v0.1.0")
	}
}

func TestSplitVersion(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"v1.2.3", []string{"1", "2", "3"}},
		{"1.0.0", []string{"1", "0", "0"}},
		{"v1.2.3-beta", []string{"1", "2", "3"}},
		{"sdk/go/v1.2.3", []string{"1", "2", "3"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := splitVersion(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d", len(got), len(tt.want))
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("[%d] = %q, want %q", i, v, tt.want[i])
				}
			}
		})
	}
}
