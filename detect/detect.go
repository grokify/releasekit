// Package detect provides language detection for repositories by scanning
// for manifest files (go.mod, package.json, Cargo.toml, etc.).
package detect

import (
	"os"
	"path/filepath"
)

// Language represents a detected programming language.
type Language string

const (
	Go         Language = "go"
	TypeScript Language = "typescript"
	JavaScript Language = "javascript"
	Python     Language = "python"
	Rust       Language = "rust"
	Swift      Language = "swift"
)

// Detection holds information about a detected language in a directory.
type Detection struct {
	Language Language `json:"language"`
	Path     string   `json:"path"`  // Directory where detected
	Files    []string `json:"files"` // Indicator files found
}

// Detect scans a directory and returns all detected languages.
func Detect(dir string) ([]Detection, error) {
	var detections []Detection

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}

		if d.IsDir() {
			name := d.Name()
			if name != "." && (name[0] == '.' || name == "node_modules" || name == "vendor" || name == "__pycache__" || name == "testdata") {
				return filepath.SkipDir
			}
			return nil
		}

		relDir, _ := filepath.Rel(dir, filepath.Dir(path))
		if relDir == "." {
			relDir = ""
		}

		switch d.Name() {
		case "go.mod":
			detections = appendIfNew(detections, Detection{
				Language: Go,
				Path:     relDir,
				Files:    []string{d.Name()},
			})
		case "package.json":
			lang := JavaScript
			tsConfig := filepath.Join(filepath.Dir(path), "tsconfig.json")
			if _, err := os.Stat(tsConfig); err == nil {
				lang = TypeScript
			}
			detections = appendIfNew(detections, Detection{
				Language: lang,
				Path:     relDir,
				Files:    []string{d.Name()},
			})
		case "Cargo.toml":
			detections = appendIfNew(detections, Detection{
				Language: Rust,
				Path:     relDir,
				Files:    []string{d.Name()},
			})
		case "Package.swift":
			detections = appendIfNew(detections, Detection{
				Language: Swift,
				Path:     relDir,
				Files:    []string{d.Name()},
			})
		case "pyproject.toml", "setup.py", "requirements.txt":
			detections = appendIfNew(detections, Detection{
				Language: Python,
				Path:     relDir,
				Files:    []string{d.Name()},
			})
		}

		return nil
	})

	return detections, err
}

// HasLanguage checks if a specific language was detected.
func HasLanguage(detections []Detection, lang Language) bool {
	for _, d := range detections {
		if d.Language == lang {
			return true
		}
	}
	return false
}

// GetByLanguage returns all detections for a specific language.
func GetByLanguage(detections []Detection, lang Language) []Detection {
	var result []Detection
	for _, d := range detections {
		if d.Language == lang {
			result = append(result, d)
		}
	}
	return result
}

func appendIfNew(detections []Detection, d Detection) []Detection {
	for i, existing := range detections {
		if existing.Language == d.Language && existing.Path == d.Path {
			detections[i].Files = append(existing.Files, d.Files...)
			return detections
		}
	}
	return append(detections, d)
}
