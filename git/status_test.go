package git

import (
	"testing"
)

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		staged    int
		unstaged  int
		untracked int
	}{
		{
			"empty",
			"",
			0, 0, 0,
		},
		{
			"modified staged",
			"M  file.go\n",
			1, 0, 0,
		},
		{
			"modified unstaged",
			" M file.go\n",
			0, 1, 0,
		},
		{
			"untracked",
			"?? newfile.go\n",
			0, 0, 1,
		},
		{
			"mixed",
			"M  staged.go\n M unstaged.go\n?? untracked.go\nA  added.go\n",
			2, 1, 1,
		},
		{
			"added and modified",
			"AM file.go\n",
			1, 1, 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := parseStatus(tt.input)
			if len(status.Staged) != tt.staged {
				t.Errorf("Staged = %d, want %d", len(status.Staged), tt.staged)
			}
			if len(status.Unstaged) != tt.unstaged {
				t.Errorf("Unstaged = %d, want %d", len(status.Unstaged), tt.unstaged)
			}
			if len(status.Untracked) != tt.untracked {
				t.Errorf("Untracked = %d, want %d", len(status.Untracked), tt.untracked)
			}
		})
	}
}
