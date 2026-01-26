package git

import (
	"fmt"
	"testing"
)

func TestParseLog(t *testing.T) {
	// Simulate git log output with our format (subject on first line, body on next)
	block := fmt.Sprintf(
		"abc123\tabc\tJohn\tjohn@example.com\t2024-06-15T10:30:00Z\tfeat: add feature\nsome body text%s%s",
		bodyEnd, commitDelimiter,
	)

	commits := parseLog(block, false)
	if len(commits) != 1 {
		t.Fatalf("got %d commits, want 1", len(commits))
	}

	c := commits[0]
	if c.Hash != "abc123" {
		t.Errorf("Hash = %q, want %q", c.Hash, "abc123")
	}
	if c.ShortHash != "abc" {
		t.Errorf("ShortHash = %q, want %q", c.ShortHash, "abc")
	}
	if c.Author != "John" {
		t.Errorf("Author = %q, want %q", c.Author, "John")
	}
	if c.Email != "john@example.com" {
		t.Errorf("Email = %q, want %q", c.Email, "john@example.com")
	}
	if c.Subject != "feat: add feature" {
		t.Errorf("Subject = %q, want %q", c.Subject, "feat: add feature")
	}
	if c.Body != "some body text" {
		t.Errorf("Body = %q, want %q", c.Body, "some body text")
	}
}

func TestParseLogMultiple(t *testing.T) {
	block := fmt.Sprintf(
		"aaa\ta\tAlice\ta@x.com\t2024-01-01T00:00:00Z\tfirst commit\n%s%s"+
			"bbb\tb\tBob\tb@x.com\t2024-01-02T00:00:00Z\tsecond commit\n%s%s",
		bodyEnd, commitDelimiter,
		bodyEnd, commitDelimiter,
	)

	commits := parseLog(block, false)
	if len(commits) != 2 {
		t.Fatalf("got %d commits, want 2", len(commits))
	}
	if commits[0].Subject != "first commit" {
		t.Errorf("first subject = %q", commits[0].Subject)
	}
	if commits[1].Subject != "second commit" {
		t.Errorf("second subject = %q", commits[1].Subject)
	}
}

func TestParseNumstat(t *testing.T) {
	input := "10\t5\tfile1.go\n3\t0\tfile2.go\n"
	stats := parseNumstat(input)
	if len(stats) != 2 {
		t.Fatalf("got %d stats, want 2", len(stats))
	}
	if stats[0].Path != "file1.go" {
		t.Errorf("path = %q, want %q", stats[0].Path, "file1.go")
	}
	if stats[0].Insertions != 10 || stats[0].Deletions != 5 {
		t.Errorf("stats = +%d/-%d, want +10/-5", stats[0].Insertions, stats[0].Deletions)
	}
}

func TestParseLogEmpty(t *testing.T) {
	commits := parseLog("", false)
	if len(commits) != 0 {
		t.Errorf("got %d commits for empty input, want 0", len(commits))
	}
}
