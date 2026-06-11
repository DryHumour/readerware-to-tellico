package isbn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ISBN
	}{
		{
			name:     "valid ISBN-10",
			input:    "0306406152",
			expected: ISBN("0306406152"),
		},
		{
			name:     "valid ISBN-10 with X",
			input:    "080442957X",
			expected: ISBN("080442957X"),
		},
		{
			name:     "valid ISBN-13",
			input:    "9780306406157",
			expected: ISBN("9780306406157"),
		},
		{
			name:     "valid ISBN-13 with isbn: tag",
			input:    "isbn: 9780306406157",
			expected: ISBN("9780306406157"),
		},
		{
			name:     "valid ISBN-10 with ISBN-10: tag",
			input:    "ISBN-10: 0306406152",
			expected: ISBN("0306406152"),
		},
		{
			name:     "valid ISBN-13 with isbn 13 : tag",
			input:    "isbn 13 : 9780306406157",
			expected: ISBN("9780306406157"),
		},
		{
			name:     "valid ISBN-10 with leading spaces and tag",
			input:    "  isbn : 080442957X",
			expected: ISBN("080442957X"),
		},
		{
			name:     "valid ISBN-13 with uppercase tag",
			input:    "ISBN-13: 9780306406157",
			expected: ISBN("9780306406157"),
		},
		{
			name:     "invalid ISBN-13 with ISBN-13 tag",
			input:    "ISBN-13: 9780306406158",
			expected: "",
		},
		{
			name:     "invalid ISBN-10 with ISBN-10 tag",
			input:    "ISBN-10: 0306406153",
			expected: "",
		},
		{
			name:     "ISBN-13 with ISBN-10 tag",
			input:    "ISBN-10: 9780306406157",
			expected: "",
		},
		{
			name:     "ISBN-10 with ISBN-13 tag",
			input:    "ISBN-13: 0306406152",
			expected: "",
		},
		{
			name:     "invalid ISBN-10 checksum",
			input:    "0306406153",
			expected: "",
		},
		{
			name:     "invalid ISBN-13 checksum",
			input:    "9780306406158",
			expected: "",
		},
		{
			name:     "invalid length",
			input:    "123",
			expected: "",
		},
		{
			name:     "invalid characters",
			input:    "030640615A",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "invalid tag number",
			input:    "ISBN-14: 0306406152",
			expected: "",
		},
		{
			name:     "bare ISBN tag with valid ISBN-10",
			input:    "ISBN: 0306406152",
			expected: ISBN("0306406152"),
		},
		{
			name:     "bare ISBN tag with valid ISBN-13",
			input:    "ISBN: 9780306406157",
			expected: ISBN("9780306406157"),
		},
		{
			name:     "bare ISBN tag with invalid ISBN",
			input:    "ISBN: 0306406153",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := New(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN10(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ISBN
	}{
		{
			name:     "valid ISBN-10",
			input:    "0306406152",
			expected: ISBN("0306406152"),
		},
		{
			name:     "valid ISBN-10 with X",
			input:    "080442957X",
			expected: ISBN("080442957X"),
		},
		{
			name:     "invalid checksum",
			input:    "0306406153",
			expected: "",
		},
		{
			name:     "invalid length",
			input:    "123456789",
			expected: "",
		},
		{
			name:     "invalid characters",
			input:    "030640615A",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ISBN10(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN13(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ISBN
	}{
		{
			name:     "valid ISBN-13",
			input:    "9780306406157",
			expected: ISBN("9780306406157"),
		},
		{
			name:     "invalid checksum",
			input:    "9780306406158",
			expected: "",
		},
		{
			name:     "invalid length",
			input:    "123456789012",
			expected: "",
		},
		{
			name:     "invalid characters",
			input:    "978030640615A",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ISBN13(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN_IsISBN10(t *testing.T) {
	tests := []struct {
		name     string
		isbn     ISBN
		expected bool
	}{
		{
			name:     "valid ISBN-10",
			isbn:     ISBN("0306406152"),
			expected: true,
		},
		{
			name:     "valid ISBN-10 with X",
			isbn:     ISBN("080442957X"),
			expected: true,
		},
		{
			name:     "invalid checksum",
			isbn:     ISBN("0306406153"),
			expected: false,
		},
		{
			name:     "wrong length",
			isbn:     ISBN("030640615"),
			expected: false,
		},
		{
			name:     "invalid characters",
			isbn:     ISBN("030640615A"),
			expected: false,
		},
		{
			name:     "valid ISBN-10 with ISBN-10 tag",
			isbn:     ISBN("ISBN-10: 0306406152"),
			expected: true,
		},
		{
			name:     "valid ISBN-10 with ISBN-13 tag",
			isbn:     ISBN("ISBN-13: 0306406152"),
			expected: false,
		},
		{
			name:     "valid ISBN-10 with invalid tag",
			isbn:     ISBN("ISBN-14: 0306406152"),
			expected: false,
		},
		{
			name:     "valid ISBN-10 with bare ISBN tag",
			isbn:     ISBN("ISBN: 0306406152"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.isbn.IsISBN10()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN_IsISBN13(t *testing.T) {
	tests := []struct {
		name     string
		isbn     ISBN
		expected bool
	}{
		{
			name:     "valid ISBN-13",
			isbn:     ISBN("9780306406157"),
			expected: true,
		},
		{
			name:     "invalid checksum",
			isbn:     ISBN("9780306406158"),
			expected: false,
		},
		{
			name:     "wrong length",
			isbn:     ISBN("978030640615"),
			expected: false,
		},
		{
			name:     "invalid characters",
			isbn:     ISBN("978030640615A"),
			expected: false,
		},
		{
			name:     "valid ISBN-13 with ISBN-13 tag",
			isbn:     ISBN("ISBN-13: 9780306406157"),
			expected: true,
		},
		{
			name:     "valid ISBN-13 with ISBN-10 tag",
			isbn:     ISBN("ISBN-10: 9780306406157"),
			expected: false,
		},
		{
			name:     "valid ISBN-13 with invalid tag",
			isbn:     ISBN("ISBN-14: 9780306406157"),
			expected: false,
		},
		{
			name:     "valid ISBN-13 with bare ISBN tag",
			isbn:     ISBN("ISBN: 9780306406157"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.isbn.IsISBN13()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN_IsISBN(t *testing.T) {
	tests := []struct {
		name     string
		isbn     ISBN
		expected bool
	}{
		{
			name:     "valid ISBN-10",
			isbn:     ISBN("0306406152"),
			expected: true,
		},
		{
			name:     "valid ISBN-13",
			isbn:     ISBN("9780306406157"),
			expected: true,
		},
		{
			name:     "invalid ISBN-10",
			isbn:     ISBN("0306406153"),
			expected: false,
		},
		{
			name:     "invalid ISBN-13",
			isbn:     ISBN("9780306406158"),
			expected: false,
		},
		{
			name:     "empty",
			isbn:     ISBN(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.isbn.IsISBN()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN_String(t *testing.T) {
	tests := []struct {
		name     string
		isbn     ISBN
		expected string
	}{
		{
			name:     "ISBN-10",
			isbn:     ISBN("0306406152"),
			expected: "0306406152",
		},
		{
			name:     "ISBN-10 with X",
			isbn:     ISBN("080442957X"),
			expected: "080442957X",
		},
		{
			name:     "ISBN-13",
			isbn:     ISBN("9780306406157"),
			expected: "9780306406157",
		},
		{
			name:     "with hyphens",
			isbn:     ISBN("978-0-306-40615-7"),
			expected: "9780306406157",
		},
		{
			name:     "with spaces",
			isbn:     ISBN("978 0 306 40615 7"),
			expected: "9780306406157",
		},
		{
			name:     "lower case",
			isbn:     ISBN("080442957x"),
			expected: "080442957X",
		},
		{
			name:     "with isbn tag prefix",
			isbn:     ISBN("isbn: 9780306406157"),
			expected: "9780306406157",
		},
		{
			name:     "with ISBN-10 tag prefix",
			isbn:     ISBN("ISBN-10: 0306406152"),
			expected: "0306406152",
		},
		{
			name:     "with ISBN 13 spaced tag",
			isbn:     ISBN("isbn 13 : 9780306406157"),
			expected: "9780306406157",
		},
		{
			name:     "invalid checksum ISBN-10",
			isbn:     ISBN("0306406153"),
			expected: "0306406153",
		},
		{
			name:     "invalid length ISBN-10",
			isbn:     ISBN("030640615"),
			expected: "030640615",
		},
		{
			name:     "invalid characters",
			isbn:     ISBN("030640615A"),
			expected: "030640615A",
		},
		{
			name:     "invalid checksum ISBN-13",
			isbn:     ISBN("9780306406158"),
			expected: "9780306406158",
		},
		{
			name:     "invalid length ISBN-13",
			isbn:     ISBN("978030640615"),
			expected: "978030640615",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.isbn.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN_Canonical(t *testing.T) {
	tests := []struct {
		name     string
		isbn     ISBN
		expected string
	}{
		{
			name:     "ISBN-10",
			isbn:     ISBN("0306406152"),
			expected: "0-306-40615-2",
		},
		{
			name:     "ISBN-13",
			isbn:     ISBN("9780306406157"),
			expected: "978-0-306-40615-7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.isbn.Canonical()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN_ISBN13(t *testing.T) {
	tests := []struct {
		name     string
		isbn     ISBN
		expected ISBN
	}{
		{
			name:     "ISBN-10 to ISBN-13",
			isbn:     ISBN("0306406152"),
			expected: ISBN("9780306406157"),
		},
		{
			name:     "already ISBN-13",
			isbn:     ISBN("9780306406157"),
			expected: ISBN("9780306406157"),
		},
		{
			name:     "invalid ISBN-10",
			isbn:     ISBN("0306406153"),
			expected: "",
		},
		{
			name:     "ISBN-13 with non-978 prefix",
			isbn:     ISBN("9791032702284"),
			expected: ISBN("9791032702284"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.isbn.ISBN13()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN_ISBN10(t *testing.T) {
	tests := []struct {
		name     string
		isbn     ISBN
		expected ISBN
	}{
		{
			name:     "ISBN-13 to ISBN-10",
			isbn:     ISBN("9780306406157"),
			expected: ISBN("0306406152"),
		},
		{
			name:     "already ISBN-10",
			isbn:     ISBN("0306406152"),
			expected: ISBN("0306406152"),
		},
		{
			name:     "invalid ISBN-13",
			isbn:     ISBN("9780306406158"),
			expected: "",
		},
		{
			name:     "ISBN-13 with non-978 prefix",
			isbn:     ISBN("9791032702284"),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.isbn.ISBN10()
			assert.Equal(t, tt.expected, result)
		})
	}
}
