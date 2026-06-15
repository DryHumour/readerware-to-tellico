package isbn

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	// ISBN10Prefix is the prefix used in ISBN-13 for the ISBN-10 equivalent.
	ISBN10Prefix = "978"
)

var (
	ErrInvalidISBN       = errors.New("invalid ISBN")
	ErrInvalidCheckDigit = fmt.Errorf("invalid ISBN check digit: %w", ErrInvalidISBN)
	ErrInvalidKind       = fmt.Errorf("invalid ISBN kind: %w", ErrInvalidISBN)
	ErrInvalidISBN10     = fmt.Errorf("invalid ISBN-10: %w", ErrInvalidKind)
	ErrInvalidISBN13     = fmt.Errorf("invalid ISBN-13: %w", ErrInvalidKind)
	ErrInvalidConversion = fmt.Errorf("invalid ISBN-13 for conversion to ISBN-10: %w", ErrInvalidISBN)
)

// ISBN represents an International Standard Book Number.
// It is guaranteed to be either an ISBN-10 or ISBN-13 or empty.
// The string will consist of only digits and the check character; no spaces or
// hyphens will be included.
// Note that the check character for an ISBN-10 may be 'X' (representing 10),
// per the ISBN specification.
type ISBN struct {
	s string
}

// New parses a string and returns the ISBN it represents.
//
// This function operates as a strict, high-performance gatekeeper. It does
// NOT parse human metadata or prefix tags (e.g., "ISBN:", "ISBN-13: ", "10 ").
//
// # Formatting and Whitespace
//
// While strictly rejecting text metadata, the parser is highly permissive of
// standard typographic formatting.  The following characters are silently
// ignored anywhere in the string:
//   - Standard whitespace (spaces, tabs, form feeds, carriage returns, line feeds)
//   - Unicode spaces (e.g., non-breaking spaces)
//   - Standard hyphens and typographic dashes (en-dash, em-dash, figure dash, etc.)
//   - Invisible formatting and bidirectional control characters (Unicode category Cf)
//
// The actual ISBN payload may consist only of ASCII digits and, for ISBN-10,
// the check character 'X' or 'x'.
//
// # Validation and Errors
//
// If the string contains invalid characters, or if the resulting payload is
// not exactly 10 or 13 characters long, an empty ISBN and an error wrapping
// ErrInvalidISBN are returned.
//
// If a 13-character payload contains an 'X', an error wrapping ErrInvalidISBN13
// is returned.
//
// # Structural Soundness
//
// If the ISBN is structurally sound (it has the correct length and valid characters)
// but the mathematical check digit is incorrect, the parser returns the populated
// ISBN struct and an error wrapping ErrInvalidCheckDigit. The caller may choose to
// ignore the error and save the structurally valid data.
func New(candidate string) (ISBN, error) {
	var (
		buf    [13]byte // accumulated ISBN digits/X
		length int      // number of ISBN digits gathered in buf
	)

	i := 0
	n := len(candidate)

	if n < 10 {
		return ISBN{}, fmt.Errorf("%w: too short to be valid: %q", ErrInvalidISBN, candidate)
	}

	for i < n {
		switch b := candidate[i]; {
		case b >= '0' && b <= '9':
			// ASCII digits 0-9.
			if length == len(buf) {
				return ISBN{}, fmt.Errorf("%w: too long to be valid: %q", ErrInvalidISBN, candidate)
			}
			buf[length] = b
			length++
			i++
		case b == 'X' || b == 'x':
			// Only valid in ISBN-10, at end.
			if length != 9 {
				return ISBN{}, fmt.Errorf("%w: X only valid in ISBN-10 at end: %q", ErrInvalidISBN, candidate)
			}
			buf[length] = 'X'
			length++
			i++
		case b == '-' || b == ' ' || b == '\t' || b == '\v' || b == '\f' || b == '\r' || b == '\n':
			// Ignore hyphens and spaces
			i++
		case b < utf8.RuneSelf:
			// No other ASCII runes are valid.
			return ISBN{}, fmt.Errorf("%w: invalid character: %q", ErrInvalidISBN, b)
		default:
			// Ignore hyphens, spaces, and format/bidi controls.
			r, width := utf8.DecodeRuneInString(candidate[i:])
			if !isHyphen(r) && !unicode.IsSpace(r) && !unicode.Is(unicode.Cf, r) {
				return ISBN{}, fmt.Errorf("%w: invalid character: %q", ErrInvalidISBN, r)
			}
			i += width
		}
	}

	var parsed ISBN

	switch length {
	case 10:
		// 'X' can only exist at buf[9], which is perfectly valid (last char).
		parsed = ISBN{s: string(buf[:10])}
	case 13:
		// ISBN-13 cannot contain 'X' (which can only be at buf[9]).
		if buf[9] == 'X' {
			return ISBN{}, fmt.Errorf("%w: ISBN-13 cannot contain X", ErrInvalidISBN13)
		}
		parsed = ISBN{s: string(buf[:13])}
	default:
		return ISBN{}, fmt.Errorf("%w: invalid body length: %d: %q", ErrInvalidISBN, length, candidate)
	}

	if !parsed.IsValid() {
		return parsed, fmt.Errorf("%w: %s", ErrInvalidCheckDigit, parsed.s)
	}

	return parsed, nil
}

// String returns the string representation of the ISBN.
// String implements the fmt.Stringer interface.
func (i ISBN) String() string {
	return i.s
}

// Len returns the length of the ISBN, either 10 or 13.
func (i ISBN) Len() int {
	return len(i.s)
}

// IsZero reports whether the ISBN is the zero value (is empty).
func (i ISBN) IsZero() bool {
	return i.s == ""
}

// Is10 reports whether the ISBN is structurally an ISBN-10 (10 characters).
func (i ISBN) Is10() bool {
	return len(i.s) == 10
}

// Is13 reports whether the ISBN is structurally an ISBN-13 (13 characters).
func (i ISBN) Is13() bool {
	return len(i.s) == 13
}

// Split splits the ISBN into body and check digit.
// The zero value returns empty strings.
func (i ISBN) Split() (body, checkDigit string) {
	switch len(i.s) {
	case 10:
		return string(i.s[:9]), string(i.s[9:])
	case 13:
		return string(i.s[:12]), string(i.s[12:])
	default:
		return "", ""
	}
}

// Body returns the body of the ISBN (the first 9 characters for ISBN-10, or
// the first 12 characters for ISBN-13) i.e. all but the check digit.
// The zero value returns an empty string.
func (i ISBN) Body() string {
	switch len(i.s) {
	case 10:
		return string(i.s[:9])
	case 13:
		return string(i.s[:12])
	default:
		return ""
	}
}

// CheckDigit returns the ASCII check digit of the ISBN.
// Note that an ISBN-10 may return 'X' as the check "digit" (meaning 10, per
// the ISBN-10 specification).
// The zero value returns 0 (ASCII NUL).
func (i ISBN) CheckDigit() byte {
	switch len(i.s) {
	case 10:
		return i.s[9]
	case 13:
		return i.s[12]
	default:
		return 0
	}
}

// ExpectedCheckDigit returns the expected ASCII check digit for the ISBN.
// Note that an ISBN-10 may return 'X' as the check "digit" (meaning 10, per
// the ISBN-10 specification).
// The zero value returns 0 (ASCII NUL).
func (i ISBN) ExpectedCheckDigit() byte {
	switch len(i.s) {
	case 10:
		return mod11(string(i.s[:9]))
	case 13:
		return mod10(string(i.s[:12]))
	default:
		return 0
	}
}

// IsValid reports whether the ISBN is a valid ISBN-10 or ISBN-13.
// The zero value returns false.
func (i ISBN) IsValid() bool {
	switch len(i.s) {
	case 10:
		body, check := string(i.s[:9]), i.s[9]
		return mod11(body) == check
	case 13:
		body, check := string(i.s[:12]), i.s[12]
		return mod10(body) == check
	default:
		return false
	}
}

// To10 converts the ISBN to ISBN-10 if possible.
// An ISBN-10 is returned as-is.
// An ISBN-13 with the 978 prefix is converted to the equivalent ISBN-10.  The
// conversion will occur even if the original check digit is incorrect.
// An ISBN-13 that cannot be converted to ISBN-10 returns an error.
// The zero value returns an error wrapping ErrInvalidISBN.
func (i ISBN) To10() (ISBN, error) {
	switch len(i.s) {
	case 10:
		return i, nil
	case 13:
		if !strings.HasPrefix(string(i.s), ISBN10Prefix) {
			return ISBN{}, ErrInvalidConversion
		}
		body := string(i.s[len(ISBN10Prefix):12])
		return appendCheck(body, mod11(body)), nil
	default:
		return ISBN{}, ErrInvalidISBN
	}
}

// To13 converts the ISBN to ISBN-13.
// An ISBN-13 is returned as-is.
// An ISBN-10 is converted to the equivalent ISBN-13.
// The zero value returns the zero value.
func (i ISBN) To13() ISBN {
	switch len(i.s) {
	case 10:
		body := ISBN10Prefix + string(i.s[:9])
		return appendCheck(body, mod10(body))
	case 13:
		return i
	default:
		return ISBN{}
	}
}

// AppendText implements encoding.TextAppender.
func (i ISBN) AppendText(b []byte) ([]byte, error) {
	return append(b, i.s...), nil
}

// MarshalText implements encoding.TextMarshaler.
func (i ISBN) MarshalText() ([]byte, error) {
	return []byte(i.s), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (i *ISBN) UnmarshalText(text []byte) error {
	parsed, err := New(string(text))
	if err != nil {
		return err
	}
	*i = parsed
	return nil
}

// mod10 calculates the ISBN-13 check digit for the given body.
func mod10(body string) byte {
	sum := 0
	for idx := range 12 {
		digit := int(body[idx] - '0')
		sum += digit * (1 + (idx&1)*2) // branchless alternating weights 1, 3, 1, 3, ....
	}
	check := byte((10 - (sum % 10)) % 10)
	return '0' + check
}

// mod11 calculates the ISBN-10 check digit for the given body.
func mod11(body string) byte {
	sum := 0
	for idx := range 9 {
		digit := int(body[idx] - '0')
		sum += digit * (10 - idx)
	}
	check := byte((11 - (sum % 11)) % 11)
	if check == 10 {
		return 'X'
	}
	return '0' + check
}

// appendCheck appends the given check digit to the body and returns the resulting ISBN.
func appendCheck(body string, check byte) ISBN {
	return ISBN{s: body + string([]byte{check})} // (avoid lint whinging about string(check)...)
}

// isHyphen reports whether r is a hyphen-like character.
func isHyphen(r rune) bool {
	switch r {
	case '-',
		'\u2010', // Typographic hyphen
		'\u2011', // Non-breaking hyphen
		'\u2012', // Figure dash
		'\u2013', // En dash
		'\u2014', // Em dash
		'\u2212': // Math minus sign
		return true
	default:
		return false
	}
}
