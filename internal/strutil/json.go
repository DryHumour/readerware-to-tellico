package strutil

import (
	json "encoding/json/v2"
	"encoding/json/jsontext"
	"strings"
)

// QuoteJSON returns s as a JSON string literal.  Invalid UTF-8 is replaced
// with U+FFFD rather than rejected; no HTML or JavaScript escaping is applied.
func QuoteJSON(s string) string {
	b, err := json.Marshal(s, jsontext.AllowInvalidUTF8(true))
	if err != nil {
		panic(err)
	}
	return string(b)
}

// UnquoteJSON reports whether s, after trimming surrounding whitespace, is a
// JSON string literal and, if so, returns its decoded value.
func UnquoteJSON(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, `"`) {
		return "", false
	}
	var v string
	if err := json.Unmarshal([]byte(s), &v, jsontext.AllowInvalidUTF8(true)); err != nil {
		return "", false
	}
	return v, true
}
