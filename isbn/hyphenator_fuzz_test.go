package isbn

import (
	"bufio"
	"errors"
	"os"
	"strings"
	"testing"
)

func FuzzHyphenator_Hyphenate(f *testing.F) {
	h := DefaultHyphenator()

	file, err := os.Open("testdata/isbn.txt")
	if err != nil {
		f.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		i, err := New(line)
		if err != nil {
			continue
		}
		f.Add(i.String())
	}
	if err := scanner.Err(); err != nil {
		f.Fatal(err)
	}

	f.Fuzz(func(t *testing.T, candidate string) {
		parsed, _ := New(candidate)
		if parsed.IsZero() {
			return
		}

		hyphenated, err := h.Hyphenate(parsed)
		switch {
		case err == nil:
		case errors.Is(err, ErrRegistrationGroupNotFound):
			return
		case errors.Is(err, ErrPublisherRangeNotFound):
			return
		case errors.Is(err, ErrInvalidPublisherLength):
			return
		default:
			t.Fatalf("Hyphenator failed unexpectedly on %q: %v", parsed.String(), err)
		}

		stripped := strings.ReplaceAll(hyphenated, "-", "")
		if stripped != parsed.String() {
			t.Fatalf("Data corruption! Original: %q, Hyphenated: %q, Stripped: %q", parsed.String(), hyphenated, stripped)
		}

		if parsed.Is10() && len(hyphenated) != 13 {
			t.Fatalf("Invalid ISBN-10 hyphenation bounds: %q (len %d)", hyphenated, len(hyphenated))
		}

		if parsed.Is13() && len(hyphenated) != 17 {
			t.Fatalf("Invalid ISBN-13 hyphenation bounds: %q (len %d)", hyphenated, len(hyphenated))
		}
	})
}
