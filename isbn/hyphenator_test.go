package isbn

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHyphenator_Hyphenate(t *testing.T) {
	t.Parallel()

	h := DefaultHyphenator()

	cases := []struct {
		name       string
		hyphenator *Hyphenator
		isbn       ISBN
		expected   string
		errIs      error
	}{
		{
			name:       "isbn13 canonical",
			hyphenator: h,
			isbn:       ni("9780306406157"),
			expected:   "978-0-306-40615-7",
		},
		{
			name:       "isbn10 converted",
			hyphenator: h,
			isbn:       ni("0306406152"),
			expected:   "0-306-40615-2",
		},
		{
			name:       "979 prefix",
			hyphenator: h,
			isbn:       ni("9791032702284"),
			expected:   "979-10-327-0228-4",
		},
		{
			name:       "nil hyphenator",
			hyphenator: nil,
			isbn:       ni("9780306406157"),
			errIs:      ErrMissingISBNRangeData,
		},
		{
			name:       "unknown registration group",
			hyphenator: h,
			isbn:       appendCheck("977123456789", mod10("977123456789")),
			errIs:      ErrRegistrationGroupNotFound,
		},
		{
			name:       "3-digit registration group",
			hyphenator: h,
			isbn:       ni("9786131234569"),
			expected:   "978-613-1-23456-9",
		},
		{
			name:       "4-digit registration group",
			hyphenator: h,
			isbn:       ni("9789989123450"),
			expected:   "978-9989-123-45-0",
		},
		{
			name:       "5-digit registration group",
			hyphenator: h,
			isbn:       ni("9789993212348"),
			expected:   "978-99932-12-34-8",
		},
		{
			name:       "isbn10 with X check digit in 3-digit registration group",
			hyphenator: h,
			isbn:       ni("613123454X"),
			expected:   "613-1-23454-X",
		},
		{
			name:       "unassigned publisher range",
			hyphenator: h,
			isbn:       appendCheck("978611123456", mod10("978611123456")),
			errIs:      ErrPublisherRangeNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			out, err := tc.hyphenator.Hyphenate(tc.isbn)
			if tc.errIs != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.errIs)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, out)
		})
	}
}
