package isbn

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
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

	// DefaultHyphenator is the default hyphenator used for ISBN hyphenation.
	//
	// It is safe for concurrent use.
	DefaultHyphenator defaultHyphenatorStore

	// isbnTagRegexp matches an optional leading actualKind prefix on the input, such as
	// "isbn: ", "ISBN-10:", or "isbn 13 :". The actualKind is captured in a named
	// group "actualKind".
	isbnTagRegexp = regexp.MustCompile(`(?i)^\s*isbn(?:(?:-|\s*)(?P<actualKind>\d+))?\s*(?::\s+)?`)
	// isbnSepRegexp matches any sequence of hyphens or whitespace characters,
	// which are used as separators in ISBNs and should be removed for validation
	// and normalization.
	isbnSepRegexp = regexp.MustCompile(`[-\s]+`)
	// isbn10Regexp matches a string that consists of exactly 9 digits followed
	// by either a digit or 'X' (case-insensitive), which is the format of a valid
	// ISBN-10 after removing separators.
	isbn10Regexp = regexp.MustCompile(`^[0-9]{9}[0-9Xx]$`)
	// isbn13Regexp matches a string that consists of exactly 13 digits, which
	// is the format of a valid ISBN-13 after removing separators.
	isbn13Regexp = regexp.MustCompile(`^[0-9]{13}$`)
)

func init() {
	DefaultHyphenator.Set(defaultHyphenator())
}

// ISBNKind represents the kind of an ISBN, either 10 or 13.
type ISBNKind string

const (
	// ISBNKind10 represents an ISBN-10.
	ISBNKind10 ISBNKind = "10"
	// ISBNKind13 represents an ISBN-13.
	ISBNKind13 ISBNKind = "13"
	// ISBNKindAny represents any ISBN kind.
	ISBNKindAny ISBNKind = ""
)

// Match checks if the ISBNKind matches the expected kind.
// Returns nil if the kinds match or if either is ISBNKindAny.
// Otherwise, returns an error with details about the mismatch.
func (k ISBNKind) Match(expected ISBNKind) error {
	if k == ISBNKindAny || expected == ISBNKindAny || k == expected {
		return nil
	}
	errs := []error{fmt.Errorf("expected ISBN kind %s, have %s: %w", expected, k, ErrInvalidKind)}
	switch expected {
	case ISBNKind10:
		errs = append(errs, ErrInvalidISBN10)
	case ISBNKind13:
		errs = append(errs, ErrInvalidISBN13)
	}
	return errors.Join(errs...)
}

// ISBN represents an International Standard Book Number with helper methods for
// validation and normalization.
type ISBN string

// New constructs a sanitised ISBN value from the provided string.
// It also accepts an optional leading actualKind prefix on the input, such as
// "isbn: ", "ISBN-10:", or "isbn 13 :". The actualKind is removed before
// sanitisation.
// It returns the zero value if the input is not a valid ISBN.
func New(value string) ISBN {
	if _, isbn, err := ParseISBN(ISBNKindAny, value); err == nil {
		return isbn
	}
	return ""
}

// ISBN10 constructs a sanitised ISBN-10 value with only digits and an optional
// final X.
// It returns the zero value if the input is not a valid ISBN-10.
func ISBN10(value string) ISBN {
	if _, isbn, err := ParseISBN(ISBNKind10, value); err == nil {
		return isbn
	}
	return ""
}

// ISBN13 constructs a sanitised ISBN-13 value with only digits.
// It returns the zero value if the input is not a valid ISBN-13.
func ISBN13(value string) ISBN {
	if _, isbn, err := ParseISBN(ISBNKind13, value); err == nil {
		return isbn
	}
	return ""
}

// ParseISBN parses an ISBN value with an optional leading tag.
// It returns the actual ISBN kind (10 or 13), the sanitised ISBN value, and an
// error if the ISBN is invalid or the kind does not match the expected kind.
// If the ISBN check digit is incorrect but is otherwise valid, it will still return
// the sanitised ISBN value with an ErrInvalidCheckDigit error.
func ParseISBN(expected ISBNKind, value string) (kind ISBNKind, isbn ISBN, err error) {
	rest := value
	match := isbnTagRegexp.FindStringSubmatch(rest)
	if len(match) != 0 {
		kind = ISBNKind(match[1])
		if err := kind.Match(expected); err != nil {
			return kind, "", err
		}
		rest = rest[len(match[0]):]
	}
	s := isbnSepRegexp.ReplaceAllString(rest, "")
	switch kind {
	case ISBNKind10:
		isbn = ISBN(strings.ToUpper(s))
		return kind, isbn, isbn.validateISBN10()
	case ISBNKind13:
		isbn = ISBN(s)
		return kind, isbn, isbn.validateISBN13()
	case ISBNKindAny:
		if isbn10Regexp.MatchString(s) {
			isbn = ISBN(strings.ToUpper(s))
			if err := expected.Match(ISBNKind10); err != nil {
				return ISBNKind10, isbn, err
			}
			return ISBNKind10, isbn, isbn.validateISBN10()
		}
		if isbn13Regexp.MatchString(s) {
			isbn = ISBN(s)
			if err := expected.Match(ISBNKind13); err != nil {
				return ISBNKind13, isbn, err
			}
			return ISBNKind13, isbn, isbn.validateISBN13()
		}
		return kind, ISBN(s), fmt.Errorf("%q: %w", s, ErrInvalidISBN)
	default:
		return kind, ISBN(s), fmt.Errorf("%q: %w", value, ErrInvalidKind)
	}
}

// String returns the ISBN value sanitised to only digits and an optional X.
// If the underlying value has an ISBN tag prefix, that tag is removed.
// If the underlying value is not a valid ISBN, the raw ISBN is returned.
func (i ISBN) String() string {
	switch _, isbn, err := ParseISBN(ISBNKindAny, string(i)); {
	case err == nil:
		return string(isbn)
	case errors.Is(err, ErrInvalidCheckDigit):
		return string(isbn)
	default:
		return string(i)
	}
}

// Hyphenate returns the ISBN in hyphenated form.
// If hyphenation fails, the original ISBN is returned.
func (i ISBN) Hyphenate() string {
	if s, err := DefaultHyphenator.Hyphenate(i); err == nil || errors.Is(err, ErrInvalidCheckDigit) {
		return s
	}
	return string(i)
}

// Canonical returns the ISBN in canonical hyphenated form.
// If the ISBN is invalid or hyphenation fails, the empty string is returned.
func (i ISBN) Canonical() string {
	if s, err := DefaultHyphenator.Hyphenate(i); err == nil {
		return s
	}
	return ""
}

// IsISBN reports whether the ISBN is a valid ISBN-10 or ISBN-13.
func (i ISBN) IsISBN() bool {
	_, _, err := ParseISBN(ISBNKindAny, string(i))
	return err == nil
}

// IsISBN10 reports whether the ISBN is a valid ISBN-10.
// If a actualKind is present, it must be "ISBN-10:".
func (i ISBN) IsISBN10() bool {
	_, _, err := ParseISBN(ISBNKind10, string(i))
	return err == nil
}

// IsISBN13 reports whether the ISBN is a valid ISBN-13.
// If a actualKind is present, it must be "ISBN-13:".
func (i ISBN) IsISBN13() bool {
	_, _, err := ParseISBN(ISBNKind13, string(i))
	return err == nil
}

// ISBN10 returns the ISBN value converted to ISBN-10 when possible.
// If the ISBN is already a valid ISBN-10, it is returned sanitised.
// If the ISBN is a valid ISBN-13 with a 978 prefix, the ISBN-10 equivalent
// is returned. Otherwise the empty string is returned.
func (i ISBN) ISBN10() ISBN {
	kind, isbn, err := ParseISBN(ISBNKindAny, string(i))
	if err != nil {
		return ""
	}
	if kind == ISBNKind10 {
		return isbn
	}
	if !strings.HasPrefix(string(isbn), ISBN10Prefix) {
		return ""
	}
	// convert ISBN-13 to ISBN-10
	core := string(i[len(ISBN10Prefix) : len(ISBN10Prefix)+9])
	sum := 0
	for idx := 0; idx < 9; idx++ {
		digit := int(core[idx] - '0')
		sum += digit * (10 - idx)
	}
	checkDigit := (11 - (sum % 11)) % 11
	if checkDigit == 10 {
		return ISBN(core + "X")
	}
	return ISBN(core + strconv.Itoa(checkDigit))
}

// ISBN13 returns the ISBN value converted to ISBN-13 when possible.
// If the ISBN is already a valid ISBN-13, it is returned sanitised.
// If the ISBN is a valid ISBN-10, the sanitised ISBN-13 equivalent is returned.
// Otherwise the empty string is returned.
func (i ISBN) ISBN13() ISBN {
	kind, isbn, err := ParseISBN(ISBNKindAny, string(i))
	if err != nil {
		return ""
	}
	if kind == ISBNKind13 {
		return isbn
	}
	// convert ISBN-10 to ISBN-13
	core := string(ISBN10Prefix + isbn[:9])
	sum := 0
	for idx := 0; idx < 12; idx++ {
		digit := int(core[idx] - '0')
		if idx%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}
	checkDigit := (10 - (sum % 10)) % 10
	return ISBN(core + strconv.Itoa(checkDigit))
}

// validateISBN10 checks if the ISBN is a valid ISBN-10.
// It assumes the ISBN has already been sanitised to only digits and an optional X.
func (i ISBN) validateISBN10() error {
	if len(i) != 10 {
		return fmt.Errorf("%q: %w", string(i), ErrInvalidISBN10)
	}
	sum := 0
	for idx := 0; idx < 9; idx++ {
		digit := int(i[idx] - '0')
		if digit < 0 || digit > 9 {
			return fmt.Errorf("%s: %w", string(i), ErrInvalidISBN10)
		}
		sum += digit * (10 - idx)
	}
	var checkValue int
	if last := i[9]; last == 'X' || last == 'x' {
		checkValue = 10
	} else if last >= '0' && last <= '9' {
		checkValue = int(last - '0')
	} else {
		return fmt.Errorf("%s: %w", string(i), ErrInvalidISBN10)
	}
	if (sum+checkValue)%11 != 0 {
		return errors.Join(
			fmt.Errorf("%s: %w", string(i), ErrInvalidISBN10),
			fmt.Errorf("%s: %w", string(i), ErrInvalidCheckDigit))
	}
	return nil
}

// validateISBN13 checks if the ISBN is a valid ISBN-13.
// It assumes the ISBN has already been sanitised to only digits.
func (i ISBN) validateISBN13() error {
	if len(i) != 13 {
		return fmt.Errorf("%q: %w", string(i), ErrInvalidISBN13)
	}
	sum := 0
	for idx := 0; idx < 12; idx++ {
		digit := int(i[idx] - '0')
		if digit < 0 || digit > 9 {
			return fmt.Errorf("%s: %w", string(i), ErrInvalidISBN13)
		}
		if idx%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}
	checkDigit := int(i[12] - '0')
	if (10-(sum%10))%10 != checkDigit {
		return errors.Join(
			fmt.Errorf("%s: %w", string(i), ErrInvalidISBN13),
			fmt.Errorf("%s: %w", string(i), ErrInvalidCheckDigit))
	}
	return nil
}

type defaultHyphenatorStore struct {
	p atomic.Pointer[Hyphenator]
}

func (s *defaultHyphenatorStore) Get() *Hyphenator {
	return s.p.Load()
}

func (s *defaultHyphenatorStore) Set(h *Hyphenator) {
	s.p.Store(h)
}

func (s *defaultHyphenatorStore) Hyphenate(isbn ISBN) (string, error) {
	h := s.p.Load()
	if h == nil {
		return "", ErrMissingISBNRangeData
	}
	return h.Hyphenate(isbn)
}

func defaultHyphenator() *Hyphenator {
	ranges, err := ParseISBNRangesJSON(isbnRangesJSON)
	if err != nil {
		panic(err)
	}
	h, err := NewHyphenator(ranges)
	if err != nil {
		panic(err)
	}
	return h
}
