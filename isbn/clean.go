package isbn

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// Clean compacts whitespace, formatting, and hyphens into single spaces,
// drops invisible control characters, and uppercases the input string.
//
// This function is designed for high-performance pipelines. It utilizes a
// zero-allocation "fast path" to return the input string immediately if
// no normalization is required.
//
// Formatting Rules:
//   - Trims all leading/trailing whitespace, punctuation, and control characters.
//   - Squashes consecutive whitespace/hyphens into a single space (' ').
//   - Discards all Unicode category Cf (format) characters.
//   - Uppercases all applicable characters.
//
// If the returned string is fewer than 10 characters, note that it cannot
// possibly represent a valid ISBN.
func Clean(s string) string {
	s = strings.TrimFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.Is(unicode.Cf, r) || unicode.Is(unicode.P, r)
	})
	if len(s) < 10 {
		return s // too short to be a valid ISBN, so no point in cleaning
	}
	firstBad := -1
	for i, r := range s {
		if isHyphen(r) || unicode.IsSpace(r) || unicode.IsLower(r) || unicode.Is(unicode.Cf, r) {
			firstBad = i
			break
		}
	}
	if firstBad == -1 {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	b.WriteString(s[:firstBad])
	inWS := false
	for _, r := range s[firstBad:] {
		switch {
		case r >= '0' && r <= '9', r == 'X', r == 'x':
			// Pass through (handled below)
		case isHyphen(r) || unicode.IsSpace(r):
			if !inWS {
				b.WriteByte(' ')
				inWS = true
			}
			continue
		case unicode.Is(unicode.Cf, r):
			continue // completely discard
		}
		b.WriteRune(unicode.ToUpper(r))
		inWS = false
	}
	return b.String()
}

// ParseTagged attempts to extract a valid ISBN from a string that may contain
// metadata tags, prefixes, or labels (e.g., "ISBN10: 0306406152" or "Notes: 978...").
//
// This function implements a multi-tiered heuristic lexer:
//
//  1. Strong Tags (e.g., "ISBN10", "ISBN-13"): These are treated as definitive
//     metadata. If the parsed payload length does not match the tag (e.g., an
//     "ISBN10" tag followed by 13 digits), the function returns an error
//     wrapping ErrInvalidISBN10 or ErrInvalidISBN13.
//
//  2. Weak Tags (e.g., "10:", "13:"): These are only treated as tags if
//     followed by a colon. This prevents ambiguity with unrelated numeric
//     data (e.g., "10 copies").
//
//  3. Opaque Label Stripping: If no tags are matched, the function attempts
//     to peel off generic "ISBN" prefixes or split on colons to isolate a
//     potential ISBN payload.
//
// If a tag is successfully matched but the subsequent payload is structurally
// invalid or mathematically incorrect, the function returns the parsed ISBN
// (if possible) joined with ErrGuessedKind and the underlying validation error.
func ParseTagged(candidate string) (ISBN, error) {
	// clean/normalise the candidate (squash spaces/hyphens to a single space, uppercase)
	cleaned := Clean(candidate)
	// try to pull off a tag
	const (
		tagISBN10  = "ISBN10"
		tagISBN13  = "ISBN13"
		tagISBNx10 = "ISBN 10"
		tagISBNx13 = "ISBN 13"
		tagWeak10  = "10:"
		tagWeak13  = "13:"
		tagWeak10x = "10 :"
		tagWeak13x = "13 :"
		tagISBN    = "ISBN"
	)
	for {
		if len(cleaned) < 10 {
			return ISBN{}, fmt.Errorf("%w: too short to be a valid ISBN: %q", ErrInvalidISBN, cleaned)
		}
		switch {
		case strings.HasPrefix(cleaned, tagISBN10):
			body := strings.TrimLeft(cleaned[len(tagISBN10):], " :")
			if parsed, err := Parse10(body); !parsed.IsZero() {
				return parsed, err
			}
		case strings.HasPrefix(cleaned, tagISBN13):
			body := strings.TrimLeft(cleaned[len(tagISBN13):], " :")
			if parsed, err := Parse13(body); !parsed.IsZero() {
				return parsed, err
			}
		case strings.HasPrefix(cleaned, tagISBNx10):
			body := strings.TrimLeft(cleaned[len(tagISBNx10):], " :")
			if parsed, err := Parse10(body); !parsed.IsZero() {
				return parsed, err
			}
		case strings.HasPrefix(cleaned, tagISBNx13):
			body := strings.TrimLeft(cleaned[len(tagISBNx13):], " :")
			if parsed, err := Parse13(body); !parsed.IsZero() {
				return parsed, err
			}
		case strings.HasPrefix(cleaned, tagWeak10), strings.HasPrefix(cleaned, tagWeak10x):
			idx := strings.IndexByte(cleaned, ':') // (cannot fail)
			cleaned = strings.TrimLeft(cleaned[idx+1:], " :")
			if parsed, err := Parse10(cleaned); !parsed.IsZero() {
				return parsed, err
			}
			continue
		case strings.HasPrefix(cleaned, tagWeak13), strings.HasPrefix(cleaned, tagWeak13x):
			idx := strings.IndexByte(cleaned, ':') // (cannot fail)
			cleaned = strings.TrimLeft(cleaned[idx+1:], " :")
			if parsed, err := Parse13(cleaned); !parsed.IsZero() {
				return parsed, err
			}
			continue
		}
		// throw out any "ISBN" prefix, if present
		s, ok := strings.CutPrefix(cleaned, tagISBN)
		if !ok {
			break
		}
		cleaned = strings.TrimLeft(s, " :")
	}
	// make an attempt assuming no tag remains
	if parsed, err := New(cleaned); err == nil || errors.Is(err, ErrInvalidCheckDigit) {
		return parsed, err
	}
	// make a final attempt, splitting on colon
	// E.g. candidate was "ISBN (Paperback) : 978-0-306-40615-7"
	if idx := strings.IndexByte(cleaned, ':'); idx >= 0 {
		if parsed, err := New(cleaned[idx+1:]); err == nil || errors.Is(err, ErrInvalidCheckDigit) {
			return parsed, err
		}
	}
	// give up
	return ISBN{}, fmt.Errorf("%w: %q", ErrInvalidISBN, candidate)
}

// Parse10 parses an ISBN-10, returning an error if the input is not a valid ISBN-10.
func Parse10(s string) (ISBN, error) {
	result, err := New(s)
	if !result.Is10() && !errors.Is(err, ErrInvalidISBN10) {
		return result, errors.Join(ErrInvalidISBN10, err)
	}
	return result, err
}

// Parse13 parses an ISBN-13, returning an error if the input is not a valid ISBN-13.
func Parse13(s string) (ISBN, error) {
	result, err := New(s)
	if !result.Is13() && !errors.Is(err, ErrInvalidISBN13) {
		return result, errors.Join(ErrInvalidISBN13, err)
	}
	return result, err
}
