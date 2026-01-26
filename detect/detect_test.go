package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func mustWriteFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.Mkdir(path, 0755); err != nil {
		t.Fatal(err)
	}
}

func TestDetectGo(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "go.mod"), []byte("module test\n"))

	detections, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !HasLanguage(detections, Go) {
		t.Error("expected Go detection")
	}
}

func TestDetectTypeScript(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "package.json"), []byte("{}"))
	mustWriteFile(t, filepath.Join(dir, "tsconfig.json"), []byte("{}"))

	detections, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !HasLanguage(detections, TypeScript) {
		t.Error("expected TypeScript detection")
	}
	if HasLanguage(detections, JavaScript) {
		t.Error("should not detect JavaScript when tsconfig.json exists")
	}
}

func TestDetectJavaScript(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "package.json"), []byte("{}"))

	detections, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !HasLanguage(detections, JavaScript) {
		t.Error("expected JavaScript detection")
	}
}

func TestDetectPython(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "pyproject.toml"), []byte(""))

	detections, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !HasLanguage(detections, Python) {
		t.Error("expected Python detection")
	}
}

func TestDetectRust(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "Cargo.toml"), []byte(""))

	detections, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !HasLanguage(detections, Rust) {
		t.Error("expected Rust detection")
	}
}

func TestDetectMultiple(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "go.mod"), []byte("module test\n"))
	sub := filepath.Join(dir, "web")
	mustMkdir(t, sub)
	mustWriteFile(t, filepath.Join(sub, "package.json"), []byte("{}"))

	detections, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(detections) != 2 {
		t.Fatalf("expected 2 detections, got %d", len(detections))
	}
	if !HasLanguage(detections, Go) {
		t.Error("expected Go")
	}
	if !HasLanguage(detections, JavaScript) {
		t.Error("expected JavaScript")
	}
}

func TestDetectSkipsVendor(t *testing.T) {
	dir := t.TempDir()
	vendor := filepath.Join(dir, "vendor")
	mustMkdir(t, vendor)
	mustWriteFile(t, filepath.Join(vendor, "go.mod"), []byte("module vendor\n"))

	detections, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(detections) != 0 {
		t.Errorf("expected 0 detections (vendor skipped), got %d", len(detections))
	}
}

func TestGetByLanguage(t *testing.T) {
	detections := []Detection{
		{Language: Go, Path: ""},
		{Language: Go, Path: "sdk"},
		{Language: TypeScript, Path: "web"},
	}
	goDetections := GetByLanguage(detections, Go)
	if len(goDetections) != 2 {
		t.Errorf("expected 2 Go detections, got %d", len(goDetections))
	}
}

func TestDetectEmpty(t *testing.T) {
	dir := t.TempDir()
	detections, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(detections) != 0 {
		t.Errorf("expected 0 detections for empty dir, got %d", len(detections))
	}
}
