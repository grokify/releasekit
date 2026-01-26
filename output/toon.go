package output

import toon "github.com/toon-format/toon-go"

type toonFormatter struct{}

func (f *toonFormatter) Format(v any) ([]byte, error) {
	return toon.Marshal(v, toon.WithIndent(2))
}
