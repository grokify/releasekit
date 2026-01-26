package output

import "encoding/json"

type jsonFormatter struct{}

func (f *jsonFormatter) Format(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
