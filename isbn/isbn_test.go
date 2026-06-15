package isbn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ni(s string) ISBN {
	return ISBN{s: s}
}

func TestNew(t *testing.T) {
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
			name:     "valid ISBN-13",
			input:    "9780306406157",
			expected: ni("9780306406157"),
		},
		{
			name:     "ISBN-13 with ISBN: tag",
			input:    "ISBN: 9780306406157",
			expected: ISBN{},
			errIs:    ErrInvalidISBN,
		},
		{
			name:     "ISBN-10 with ISBN-10: tag",
			input:    "ISBN-10: 0306406152",
			expected: ISBN{},
			errIs:    ErrInvalidISBN,
		},
		{
			name:     "ISBN-13 with ISBN-13: tag",
			input:    "ISBN-13: 9780306406157",
			expected: ISBN{},
			errIs:    ErrInvalidISBN,
		},
		{
			name:     "invalid ISBN-10 checksum",
			input:    "0306406153",
			expected: ni("0306406153"),
			errIs:    ErrInvalidCheckDigit,
		},
		{
			name:     "invalid ISBN-13 checksum",
			input:    "9780306406158",
			expected: ni("9780306406158"),
			errIs:    ErrInvalidCheckDigit,
		},
		{
			name:     "ISBN-10 with 13 and invalid check",
			input:    "1304010000",
			expected: ni("1304010000"),
			errIs:    ErrInvalidCheckDigit,
		},
		{
			name:     "invalid length",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := New(tt.input)
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

func TestISBN_Is10(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		isbn     ISBN
		expected bool
	}{
		{
			name:     "valid ISBN-10",
			isbn:     ni("0306406152"),
			expected: true,
		},
		{
			name:     "valid ISBN-10 with X",
			isbn:     ni("080442957X"),
			expected: true,
		},
		{
			name:     "invalid checksum",
			isbn:     ni("0306406153"),
			expected: true, // Still 10 characters, so Is10 returns true
		},
		{
			name:     "ISBN-13",
			isbn:     ni("9780306406157"),
			expected: false,
		},
		{
			name:     "zero",
			isbn:     ISBN{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.isbn.Is10()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN_Is13(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		isbn     ISBN
		expected bool
	}{
		{
			name:     "valid ISBN-13",
			isbn:     ni("9780306406157"),
			expected: true,
		},
		{
			name:     "invalid checksum",
			isbn:     ni("9780306406158"),
			expected: true, // Still 13 characters, so Is13 returns true
		},
		{
			name:     "ISBN-10",
			isbn:     ni("0306406152"),
			expected: false,
		},
		{
			name:     "zero",
			isbn:     ISBN{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.isbn.Is13()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN_IsValid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		isbn     ISBN
		expected bool
	}{
		{
			name:     "valid ISBN-10",
			isbn:     ni("0306406152"),
			expected: true,
		},
		{
			name:     "valid ISBN-13",
			isbn:     ni("9780306406157"),
			expected: true,
		},
		{
			name:     "invalid ISBN-10",
			isbn:     ni("0306406153"),
			expected: false,
		},
		{
			name:     "invalid ISBN-13",
			isbn:     ni("9780306406158"),
			expected: false,
		},
		{
			name:     "zero",
			isbn:     ISBN{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.isbn.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		isbn     ISBN
		expected string
	}{
		{
			name:     "ISBN-10",
			isbn:     ni("0306406152"),
			expected: "0306406152",
		},
		{
			name:     "ISBN-10 with X",
			isbn:     ni("080442957X"),
			expected: "080442957X",
		},
		{
			name:     "ISBN-13",
			isbn:     ni("9780306406157"),
			expected: "9780306406157",
		},
		{
			name:     "zero",
			isbn:     ISBN{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.isbn.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN_To13(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		isbn     ISBN
		expected ISBN
	}{
		{
			name:     "ISBN-10 to ISBN-13",
			isbn:     ni("0306406152"),
			expected: ni("9780306406157"),
		},
		{
			name:     "already ISBN-13",
			isbn:     ni("9780306406157"),
			expected: ni("9780306406157"),
		},
		{
			name:     "invalid ISBN-10",
			isbn:     ni("0306406153"),
			expected: ni("9780306406157"),
		},
		{
			name:     "ISBN-13 with non-978 prefix",
			isbn:     ni("9791032702284"),
			expected: ni("9791032702284"),
		},
		{
			name:     "zero",
			isbn:     ISBN{},
			expected: ISBN{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.isbn.To13()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN_To10(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		isbn     ISBN
		expected ISBN
		errIs    error
	}{
		{
			name:     "ISBN-13 to ISBN-10",
			isbn:     ni("9780306406157"),
			expected: ni("0306406152"),
		},
		{
			name:     "already ISBN-10",
			isbn:     ni("0306406152"),
			expected: ni("0306406152"),
		},
		{
			name:     "invalid ISBN-13",
			isbn:     ni("9780306406158"),
			expected: ni("0306406152"), // Converts to ISBN-10 with correct check digit
		},
		{
			name:  "ISBN-13 with non-978 prefix",
			isbn:  ni("9791032702284"),
			errIs: ErrInvalidConversion,
		},
		{
			name:     "zero",
			isbn:     ISBN{},
			expected: ISBN{},
			errIs:    ErrInvalidISBN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := tt.isbn.To10()
			if tt.errIs != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.errIs)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN_Split(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		isbn       ISBN
		body       string
		checkDigit string
	}{
		{
			name:       "valid ISBN-10",
			isbn:       ni("0306406152"),
			body:       "030640615",
			checkDigit: "2",
		},
		{
			name:       "valid ISBN-10 with X",
			isbn:       ni("080442957X"),
			body:       "080442957",
			checkDigit: "X",
		},
		{
			name:       "valid ISBN-13",
			isbn:       ni("9780306406157"),
			body:       "978030640615",
			checkDigit: "7",
		},
		{
			name:       "zero",
			isbn:       ISBN{},
			body:       "",
			checkDigit: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			body, checkDigit := tt.isbn.Split()
			assert.Equal(t, tt.body, body)
			assert.Equal(t, tt.checkDigit, checkDigit)
		})
	}
}

func TestISBN_Body(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		isbn     ISBN
		expected string
	}{
		{
			name:     "valid ISBN-10",
			isbn:     ni("0306406152"),
			expected: "030640615",
		},
		{
			name:     "valid ISBN-10 with X",
			isbn:     ni("080442957X"),
			expected: "080442957",
		},
		{
			name:     "valid ISBN-13",
			isbn:     ni("9780306406157"),
			expected: "978030640615",
		},
		{
			name:     "zero",
			isbn:     ISBN{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.isbn.Body()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN_CheckDigit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		isbn     ISBN
		expected byte
	}{
		{
			name:     "valid ISBN-10",
			isbn:     ni("0306406152"),
			expected: '2',
		},
		{
			name:     "valid ISBN-10 with X",
			isbn:     ni("080442957X"),
			expected: 'X',
		},
		{
			name:     "valid ISBN-13",
			isbn:     ni("9780306406157"),
			expected: '7',
		},
		{
			name:     "zero",
			isbn:     ISBN{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.isbn.CheckDigit()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN_ExpectedCheckDigit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		isbn     ISBN
		expected byte
	}{
		{
			name:     "valid ISBN-10",
			isbn:     ni("0306406152"),
			expected: '2',
		},
		{
			name:     "valid ISBN-10 with X",
			isbn:     ni("080442957X"),
			expected: 'X',
		},
		{
			name:     "valid ISBN-13",
			isbn:     ni("9780306406157"),
			expected: '7',
		},
		{
			name:     "invalid ISBN-10 with wrong check digit",
			isbn:     ni("0306406153"),
			expected: '2',
		},
		{
			name:     "invalid ISBN-13 with wrong check digit",
			isbn:     ni("9780306406158"),
			expected: '7',
		},
		{
			name:     "zero",
			isbn:     ISBN{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.isbn.ExpectedCheckDigit()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN_Len(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		isbn     ISBN
		expected int
	}{
		{
			name:     "valid ISBN-10",
			isbn:     ni("0306406152"),
			expected: 10,
		},
		{
			name:     "valid ISBN-10 with X",
			isbn:     ni("080442957X"),
			expected: 10,
		},
		{
			name:     "valid ISBN-13",
			isbn:     ni("9780306406157"),
			expected: 13,
		},
		{
			name:     "zero",
			isbn:     ISBN{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.isbn.Len()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestISBN_IsZero(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		isbn     ISBN
		expected bool
	}{
		{
			name:     "valid ISBN-10",
			isbn:     ni("0306406152"),
			expected: false,
		},
		{
			name:     "valid ISBN-13",
			isbn:     ni("9780306406157"),
			expected: false,
		},
		{
			name:     "zero",
			isbn:     ISBN{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.isbn.IsZero()
			assert.Equal(t, tt.expected, result)
		})
	}
}
