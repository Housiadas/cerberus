package jsonutil

import (
	"bytes"
	"encoding/json"
)

// Compact removes insignificant whitespace from JSON data.
// If the input is empty or invalid JSON, the original string is returned.
func Compact(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	var buf bytes.Buffer

	err := json.Compact(&buf, data)
	if err != nil {
		return string(data)
	}

	return buf.String()
}
