// Package output provides formatters for rendering structured data
// as TOON, JSON, or plain text.
package output

import "fmt"

// Format identifies the output format.
type Format string

const (
	FormatTOON Format = "toon"
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// Formatter writes structured data in the specified format.
type Formatter interface {
	Format(v any) ([]byte, error)
}

// NewFormatter returns a Formatter for the given format.
func NewFormatter(format Format) (Formatter, error) {
	switch format {
	case FormatTOON:
		return &toonFormatter{}, nil
	case FormatJSON:
		return &jsonFormatter{}, nil
	case FormatText:
		return &textFormatter{}, nil
	default:
		return nil, fmt.Errorf("unsupported output format: %q", format)
	}
}
