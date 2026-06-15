package isbn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClean(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid ISBN-10 unchanged",
			input:    "0306406152",
			expected: "0306406152",
		},
		{
			name:     "valid ISBN-13 unchanged",
			input:    "9780306406157",
			expected: "9780306406157",
		},
		{
			name:     "hyphens to spaces",
			input:    "0-306-40615-2",
			expected: "0 306 40615 2",
		},
		{
			name:     "multiple hyphens to single space",
			input:    "0--306---40615--2",
			expected: "0 306 40615 2",
		},
		{
			name:     "spaces to single space",
			input:    "0  306   40615    2",
			expected: "0 306 40615 2",
		},
		{
			name:     "mixed hyphens and spaces",
			input:    "0-306 40615-2",
			expected: "0 306 40615 2",
		},
		{
			name:     "leading whitespace trimmed",
			input:    "   0306406152",
			expected: "0306406152",
		},
		{
			name:     "trailing whitespace trimmed",
			input:    "0306406152   ",
			expected: "0306406152",
		},
		{
			name:     "leading and trailing whitespace trimmed",
			input:    "   0306406152   ",
			expected: "0306406152",
		},
		{
			name:     "lowercase to uppercase",
			input:    "080442957x",
			expected: "080442957X",
		},
		{
			name:     "mixed case to uppercase",
			input:    "080442957Xx",
			expected: "080442957XX",
		},
		{
			name:     "en dash",
			input:    "0\u2013306\u201340615\u20132",
			expected: "0 306 40615 2",
		},
		{
			name:     "em dash",
			input:    "0\u2014306\u201440615\u20142",
			expected: "0 306 40615 2",
		},
		{
			name:     "figure dash",
			input:    "0\u2012306\u201240615\u20122",
			expected: "0 306 40615 2",
		},
		{
			name:     "non-breaking hyphen",
			input:    "0\u2011306\u201140615\u20112",
			expected: "0 306 40615 2",
		},
		{
			name:     "math minus sign",
			input:    "0\u2212306\u221240615\u22122",
			expected: "0 306 40615 2",
		},
		{
			name:     "tabs to spaces",
			input:    "0\t306\t40615\t2",
			expected: "0 306 40615 2",
		},
		{
			name:     "newlines to spaces",
			input:    "0\n306\n40615\n2",
			expected: "0 306 40615 2",
		},
		{
			name:     "mixed whitespace types",
			input:    "0 \t\n306 \t\n40615 \t\n2",
			expected: "0 306 40615 2",
		},
		{
			name:     "leading punctuation trimmed",
			input:    ",.0306406152",
			expected: "0306406152",
		},
		{
			name:     "trailing punctuation trimmed",
			input:    "0306406152,.",
			expected: "0306406152",
		},
		{
			name:     "short string unchanged",
			input:    "123",
			expected: "123",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only whitespace",
			input:    "   ",
			expected: "",
		},
		{
			name:     "only hyphens",
			input:    "---",
			expected: "",
		},
		{
			name:     "complex real-world case",
			input:    "ISBN-13: 978-0-306-40615-7",
			expected: "ISBN 13: 978 0 306 40615 7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := Clean(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseTagged(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected ISBN
		errIs    error
	}{
		{
			name:     "ISBN10 tag with valid ISBN-10",
			input:    "ISBN10: 0306406152",
			expected: ni("0306406152"),
		},
		{
			name:     "ISBN10 tag with valid ISBN-10 and X",
			input:    "ISBN10: 080442957X",
			expected: ni("080442957X"),
		},
		{
			name:     "ISBN13 tag with valid ISBN-13",
			input:    "ISBN13: 9780306406157",
			expected: ni("9780306406157"),
		},
		{
			name:     "ISBN 10 spaced tag with valid ISBN-10",
			input:    "ISBN 10: 0306406152",
			expected: ni("0306406152"),
		},
		{
			name:     "ISBN 13 spaced tag with valid ISBN-13",
			input:    "ISBN 13: 9780306406157",
			expected: ni("9780306406157"),
		},
		{
			name:     "weak 10: tag with valid ISBN-10",
			input:    "10: 0306406152",
			expected: ni("0306406152"),
		},
		{
			name:     "weak 13: tag with valid ISBN-13",
			input:    "13: 9780306406157",
			expected: ni("9780306406157"),
		},
		{
			name:     "weak 10 : spaced tag with valid ISBN-10",
			input:    "10 : 0306406152",
			expected: ni("0306406152"),
		},
		{
			name:     "weak 13 : spaced tag with valid ISBN-13",
			input:    "13 : 9780306406157",
			expected: ni("9780306406157"),
		},
		{
			name:     "bare ISBN prefix with valid ISBN-10",
			input:    "ISBN: 0306406152",
			expected: ni("0306406152"),
		},
		{
			name:     "bare ISBN prefix with valid ISBN-13",
			input:    "ISBN: 9780306406157",
			expected: ni("9780306406157"),
		},
		{
			name:     "ISBN10 tag with invalid check digit",
			input:    "ISBN10: 0306406153",
			expected: ni("0306406153"),
			errIs:    ErrInvalidCheckDigit,
		},
		{
			name:     "ISBN13 tag with invalid check digit",
			input:    "ISBN13: 9780306406158",
			expected: ni("9780306406158"),
			errIs:    ErrInvalidCheckDigit,
		},
		{
			name:     "ISBN10 tag with ISBN-13 length",
			input:    "ISBN10: 9780306406157",
			expected: ni("9780306406157"),
			errIs:    ErrInvalidISBN10,
		},
		{
			name:     "ISBN13 tag with ISBN-10 length",
			input:    "ISBN13: 0306406152",
			expected: ni("0306406152"),
			errIs:    ErrInvalidISBN13,
		},
		{
			name:     "no tag with valid ISBN-10",
			input:    "0306406152",
			expected: ni("0306406152"),
		},
		{
			name:     "no tag with valid ISBN-13",
			input:    "9780306406157",
			expected: ni("9780306406157"),
		},
		{
			name:     "colon split with valid ISBN-10",
			input:    "Notes: 0306406152",
			expected: ni("0306406152"),
		},
		{
			name:     "colon split with valid ISBN-13",
			input:    "Notes: 9780306406157",
			expected: ni("9780306406157"),
		},
		{
			name:     "too short",
			input:    "123",
			expected: ISBN{},
			errIs:    ErrInvalidISBN,
		},
		{
			name:     "completely invalid",
			input:    "not an isbn",
			expected: ISBN{},
			errIs:    ErrInvalidISBN,
		},
		{
			name:     "ISBN prefix with colon and valid ISBN-10",
			input:    "ISBN (Paperback) : 0306406152",
			expected: ni("0306406152"),
		},
		{
			name:     "multiple ISBN prefixes stripped",
			input:    "ISBN-10: ISBN 0306406152",
			expected: ni("0306406152"),
		},
		{
			name:     "weak tag without colon not treated as tag",
			input:    "10 copies of book",
			expected: ISBN{},
			errIs:    ErrInvalidISBN,
		},
		{
			name:     "hyphenated ISBN with tag",
			input:    "ISBN10: 0-306-40615-2",
			expected: ni("0306406152"),
		},
		{
			name:     "spaced ISBN with tag",
			input:    "ISBN13: 978 0 306 40615 7",
			expected: ni("9780306406157"),
		},
		{
			name:     "lowercase tag",
			input:    "isbn10: 0306406152",
			expected: ni("0306406152"),
		},
		{
			name:     "multiple tags",
			input:    "ISBN-10: ISBN: 10: ISBN-10: 0306406152",
			expected: ni("0306406152"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := ParseTagged(tt.input)
			assert.Equal(t, tt.expected, result)
			if tt.errIs != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.errIs)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestParse10(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected ISBN
		errIs    error
	}{
		{
			name:     "valid ISBN-10",
			input:    "0306406152",
			expected: ni("0306406152"),
		},
		{
			name:     "valid ISBN-10 with X",
			input:    "080442957X",
			expected: ni("080442957X"),
		},
		{
			name:     "valid ISBN-10 with hyphens",
			input:    "0-306-40615-2",
			expected: ni("0306406152"),
		},
		{
			name:     "valid ISBN-10 with spaces",
			input:    "0 306 40615 2",
			expected: ni("0306406152"),
		},
		{
			name:     "invalid check digit",
			input:    "0306406153",
			expected: ni("0306406153"),
			errIs:    ErrInvalidCheckDigit,
		},
		{
			name:     "ISBN-13 passed to Parse10",
			input:    "9780306406157",
			expected: ni("9780306406157"),
			errIs:    ErrInvalidISBN10,
		},
		{
			name:     "too short",
			input:    "123",
			expected: ISBN{},
			errIs:    ErrInvalidISBN,
		},
		{
			name:     "invalid characters",
			input:    "030640615A",
			expected: ISBN{},
			errIs:    ErrInvalidISBN,
		},
		{
			name:     "empty string",
			input:    "",
			expected: ISBN{},
			errIs:    ErrInvalidISBN,
		},
		{
			name:     "ISBN-13 with X",
			input:    "978030640615X",
			expected: ISBN{},
			errIs:    ErrInvalidISBN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := Parse10(tt.input)
			assert.Equal(t, tt.expected, result)
			if tt.errIs != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.errIs)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestParse13(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected ISBN
		errIs    error
	}{
		{
			name:     "valid ISBN-13",
			input:    "9780306406157",
			expected: ni("9780306406157"),
		},
		{
			name:     "valid ISBN-13 with hyphens",
			input:    "978-0-306-40615-7",
			expected: ni("9780306406157"),
		},
		{
			name:     "valid ISBN-13 with spaces",
			input:    "978 0 306 40615 7",
			expected: ni("9780306406157"),
		},
		{
			name:     "invalid check digit",
			input:    "9780306406158",
			expected: ni("9780306406158"),
			errIs:    ErrInvalidCheckDigit,
		},
		{
			name:     "ISBN-10 passed to Parse13",
			input:    "0306406152",
			expected: ni("0306406152"),
			errIs:    ErrInvalidISBN13,
		},
		{
			name:     "too short",
			input:    "123",
			expected: ISBN{},
			errIs:    ErrInvalidISBN,
		},
		{
			name:     "invalid characters",
			input:    "978030640615A",
			expected: ISBN{},
			errIs:    ErrInvalidISBN,
		},
		{
			name:     "empty string",
			input:    "",
			expected: ISBN{},
			errIs:    ErrInvalidISBN,
		},
		{
			name:     "ISBN-13 with X",
			input:    "978030640615X",
			expected: ISBN{},
			errIs:    ErrInvalidISBN13,
		},
		{
			name:     "ISBN-10 with X passed to Parse13",
			input:    "080442957X",
			expected: ni("080442957X"),
			errIs:    ErrInvalidISBN13,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := Parse13(tt.input)
			assert.Equal(t, tt.expected, result)
			if tt.errIs != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.errIs)
				return
			}
			require.NoError(t, err)
		})
	}
}
