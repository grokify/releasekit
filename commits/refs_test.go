package commits

import (
	"reflect"
	"testing"
)

func TestExtractIssueRefs(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []int
	}{
		{"single ref", "fix #42", []int{42}},
		{"multiple refs", "fix #42 and #99", []int{42, 99}},
		{"no refs", "just a message", nil},
		{"PR ref excluded", "feat (#123)", nil},
		{"mixed refs", "fix #42 (#123) #99", []int{42, 99}},
		{"start of line", "#42 is fixed", []int{42}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractIssueRefs(tt.text)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractIssueRefs(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

func TestExtractPRRefs(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []int
	}{
		{"single PR", "feat (#123)", []int{123}},
		{"multiple PRs", "feat (#1) (#2)", []int{1, 2}},
		{"no PRs", "fix #42", nil},
		{"mixed", "fix #42 (#99)", []int{99}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractPRRefs(tt.text)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractPRRefs(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}
