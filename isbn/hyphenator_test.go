package isbn

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func isbn13WithCheckDigit(core12 string) ISBN {
	if len(core12) != 12 {
		return ""
	}
	sum := 0
	for i := 0; i < 12; i++ {
		d := int(core12[i] - '0')
		if d < 0 || d > 9 {
			return ""
		}
		if i%2 == 0 {
			sum += d
		} else {
			sum += d * 3
		}
	}
	check := (10 - (sum % 10)) % 10
	return ISBN(core12 + strconv.Itoa(check))
}

func TestHyphenator_Hyphenate(t *testing.T) {
	ranges, err := ParseISBNRangesJSON(isbnRangesJSON)
	require.NoError(t, err)

	h, err := NewHyphenator(ranges)
	require.NoError(t, err)

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
			isbn:       ISBN("9780306406157"),
			expected:   "978-0-306-40615-7",
		},
		{
			name:       "isbn13 already hyphenated",
			hyphenator: h,
			isbn:       ISBN("978-0-306-40615-7"),
			expected:   "978-0-306-40615-7",
		},
		{
			name:       "isbn13 with tag",
			hyphenator: h,
			isbn:       ISBN("ISBN-13: 9780306406157"),
			expected:   "978-0-306-40615-7",
		},
		{
			name:       "isbn10 converted",
			hyphenator: h,
			isbn:       ISBN("0306406152"),
			expected:   "0-306-40615-2",
		},
		{
			name:       "979 prefix",
			hyphenator: h,
			isbn:       ISBN("9791032702284"),
			expected:   "979-10-327-0228-4",
		},
		{
			name:       "nil hyphenator",
			hyphenator: nil,
			isbn:       ISBN("9780306406157"),
			errIs:      ErrMissingISBNRangeData,
		},
		{
			name:       "invalid isbn",
			hyphenator: h,
			isbn:       ISBN("9780306406158"),
			errIs:      ErrInvalidISBN,
		},
		{
			name:       "unknown registration group",
			hyphenator: h,
			isbn:       isbn13WithCheckDigit("977123456789"),
			errIs:      ErrRegistrationGroupNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
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

func TestHyphenator_Hyphenate_ExportCSV(t *testing.T) {
	t.Skip("Skipping export.csv test")

	path := filepath.Join("..", "export.csv")
	file, err := os.Open(path)
	if err != nil {
		t.Skipf("unable to read %s: %v", path, err)
	}
	defer file.Close()

	ranges, err := ParseISBNRangesJSON(isbnRangesJSON)
	require.NoError(t, err)

	h, err := NewHyphenator(ranges)
	require.NoError(t, err)

	r := csv.NewReader(file)
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	header, err := r.Read()
	require.NoError(t, err)

	isbnIdx := -1
	titleIdx := -1
	for i, h := range header {
		if strings.EqualFold(strings.TrimSpace(h), "ISBN") {
			isbnIdx = i
			continue
		}
		if strings.EqualFold(strings.TrimSpace(h), "TITLE") {
			titleIdx = i
			continue
		}
	}
	if isbnIdx < 0 {
		t.Skip("no ISBN column found in export.csv")
	}

	bad := make([]string, 0)
	badCount := 0
	checked := 0
	line := 1
	for {
		rec, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			require.NoError(t, err)
		}
		line++
		if isbnIdx >= len(rec) {
			continue
		}
		s := strings.TrimSpace(rec[isbnIdx])
		if s == "" {
			continue
		}
		title := ""
		if titleIdx >= 0 && titleIdx < len(rec) {
			title = strings.TrimSpace(rec[titleIdx])
		}
		checked++
		_, err = h.Hyphenate(ISBN(s))
		if err != nil {
			badCount++
			if len(bad) < 10 {
				bad = append(bad, fmt.Sprintf("line %d: %s (%s): %v", line, s, title, err))
			}
		}
	}
	if checked == 0 {
		t.Skip("no ISBN values found in export.csv ISBN column")
	}

	if badCount != 0 {
		require.Failf(t, "export.csv contains ISBNs that failed hyphenation", "bad=%d (showing up to 10)\n%v", badCount, bad)
	}
}
