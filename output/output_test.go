package output

import (
	"encoding/json"
	"strings"
	"testing"
)

type testData struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Count   int    `json:"count"`
}

func TestNewFormatterValid(t *testing.T) {
	for _, format := range []Format{FormatTOON, FormatJSON, FormatText} {
		f, err := NewFormatter(format)
		if err != nil {
			t.Errorf("NewFormatter(%q) error: %v", format, err)
		}
		if f == nil {
			t.Errorf("NewFormatter(%q) returned nil", format)
		}
	}
}

func TestNewFormatterInvalid(t *testing.T) {
	_, err := NewFormatter("xml")
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestJSONFormatter(t *testing.T) {
	f, _ := NewFormatter(FormatJSON)
	data := testData{Name: "test", Version: "1.0.0", Count: 42}

	out, err := f.Format(data)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	// Verify it's valid JSON
	var decoded testData
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if decoded.Name != "test" || decoded.Version != "1.0.0" || decoded.Count != 42 {
		t.Errorf("decoded = %+v, want original data", decoded)
	}
}

func TestTOONFormatter(t *testing.T) {
	f, _ := NewFormatter(FormatTOON)
	data := testData{Name: "test", Version: "1.0.0", Count: 42}

	out, err := f.Format(data)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	// Basic sanity: TOON output should contain the field values
	s := string(out)
	if !strings.Contains(s, "test") {
		t.Errorf("TOON output missing 'test': %s", s)
	}
	if !strings.Contains(s, "1.0.0") {
		t.Errorf("TOON output missing '1.0.0': %s", s)
	}
}

func TestTextFormatter(t *testing.T) {
	f, _ := NewFormatter(FormatText)
	data := testData{Name: "test", Version: "1.0.0", Count: 42}

	out, err := f.Format(data)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	s := string(out)
	if !strings.Contains(s, "name: test") {
		t.Errorf("text output missing 'name: test': %s", s)
	}
	if !strings.Contains(s, "version: 1.0.0") {
		t.Errorf("text output missing 'version: 1.0.0': %s", s)
	}
}

func TestTextFormatterNil(t *testing.T) {
	f, _ := NewFormatter(FormatText)
	out, err := f.Format(nil)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}
	if string(out) != "" {
		t.Errorf("expected empty output for nil, got %q", string(out))
	}
}
