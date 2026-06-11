package isbn

import (
	"encoding/csv"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func FuzzHyphenator_Hyphenate(f *testing.F) {
	ranges, err := ParseISBNRangesJSON(isbnRangesJSON)
	if err != nil {
		f.Fatalf("ParseISBNRangesJSON: %v", err)
	}
	h, err := NewHyphenator(ranges)
	if err != nil {
		f.Fatalf("NewHyphenator: %v", err)
	}

	for _, seed := range []string{
		"9780306406157",
		"978-0-306-40615-7",
		"ISBN-13: 9780306406157",
		"isbn: 9780306406157",
		"0306406152",
		"9791032702284",
		"979-10-3270-228-4",
		"",
		"not an isbn",
		"9780306406158",
	} {
		f.Add(seed)
	}

	// Seed with a larger corpus of known-good ISBNs if available.
	// This keeps the fuzzing focused on hyphenation edge cases rather than spending
	// all its time on obviously invalid inputs.
	seedFromExportCSV(f, 250)

	f.Fuzz(func(t *testing.T, s string) {
		out, err := h.Hyphenate(ISBN(s))
		clean := ISBN(s).ISBN13()

		if s != "" && clean == "" {
			if err == nil {
				t.Fatalf("expected error for invalid ISBN input %q, got output %q", s, out)
			}
			if !errors.Is(err, ErrInvalidISBN) {
				t.Fatalf("expected ErrInvalidISBN for input %q, got %v", s, err)
			}
			return
		}

		if err != nil {
			if !(errors.Is(err, ErrRegistrationGroupNotFound) ||
				errors.Is(err, ErrPublisherRangeNotFound) ||
				errors.Is(err, ErrInvalidPublisherLength) ||
				errors.Is(err, ErrInvalidISBN) ||
				errors.Is(err, ErrMissingISBNRangeData)) {
				t.Fatalf("unexpected error category for input %q (clean %q): %v", s, clean, err)
			}
			return
		}

		// If we returned a hyphenated ISBN, it should parse back to the same ISBN-13.
		if got := ISBN(out).ISBN13(); got != clean {
			t.Fatalf("round-trip mismatch: input=%q clean=%q out=%q parsedOut=%q", s, clean, out, got)
		}
	})
}

func seedFromExportCSV(f *testing.F, maxSeeds int) {
	path := filepath.Join("..", "export.csv")
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	r := csv.NewReader(file)
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	header, err := r.Read()
	if err != nil {
		return
	}

	isbnIdx := -1
	for i, h := range header {
		if strings.EqualFold(strings.TrimSpace(h), "ISBN") {
			isbnIdx = i
			break
		}
	}
	if isbnIdx < 0 {
		return
	}

	seeded := 0
	for seeded < maxSeeds {
		rec, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return
		}
		if isbnIdx >= len(rec) {
			continue
		}
		v := strings.TrimSpace(rec[isbnIdx])
		if v == "" {
			continue
		}
		f.Add(v)
		seeded++
	}
}
