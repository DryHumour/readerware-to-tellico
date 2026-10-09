package strutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuoteJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "plain ASCII", input: "hello", expected: `"hello"`},
		{name: "quote backslash and control", input: "a\"b\\c\x01", expected: `"a\"b\\c\u0001"`},
		{name: "HTML characters not escaped", input: "<a>&b", expected: `"<a>&b"`},
		{name: "invalid UTF-8 becomes U+FFFD", input: "a\xffb", expected: "\"a�b\""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, QuoteJSON(tt.input), "QuoteJSON(%q)", tt.input)
		})
	}
}

func TestUnquoteJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
		ok       bool
	}{
		{name: "valid literal with surrounding space", input: `  "abc" `, expected: "abc", ok: true},
		{name: "unicode escape decodes", input: `"\u0039"`, expected: "9", ok: true},
		{name: "bare digits are not a string", input: "123", ok: false},
		{name: "leading quote but malformed", input: `"abc`, ok: false},
		{name: "empty string", input: "", ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := UnquoteJSON(tt.input)
			assert.Equal(t, tt.ok, ok, "UnquoteJSON(%q) ok", tt.input)
			assert.Equal(t, tt.expected, got, "UnquoteJSON(%q) value", tt.input)
		})
	}
}
