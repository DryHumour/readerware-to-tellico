package isbn

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
)

const (
	// isbn13BodyDigits is the number of body (non-check) digits in an ISBN-13.
	isbn13BodyDigits = 12
)

var (
	ErrMissingISBNRangeData      = errors.New("missing ISBN range data")
	ErrRegistrationGroupNotFound = fmt.Errorf("registration group not found: %w", ErrInvalidISBN)
	ErrPublisherRangeNotFound    = fmt.Errorf("invalid or unassigned publisher range: %w", ErrInvalidISBN)
	ErrInvalidPublisherLength    = fmt.Errorf("invalid publisher length: %w", ErrInvalidISBN)

	// defaultHyphenator is a lazy once-value that returns the default hyphenator.
	defaultHyphenator = sync.OnceValue(newDefaultHyphenator)
)

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// Hyphenator is an optimized, read-only representation of ISBN range data for fast hyphenation.
// Construct it once via NewHyphenator and reuse it for many ISBNs.
//
// It assumes the underlying ISBNRanges data is valid per the official ISBN range message.
// If the data is malformed, hyphenation may return an error.
type Hyphenator struct {
	// rules is sorted lexicographically by (prefix, minBound), both ascending.
	// Registration group prefixes are prefix-free, so a single binary search
	// against a clean ISBN-13 locates the unique covering rule.
	rules []compiledRule
}

// compiledRule is a single hyphenation rule keyed by the digits-only EAN+group
// prefix. The bounds are clamped at compile time so that
// len(prefix)+len(minBound) <= isbn13BodyDigits, allowing the digits following
// the prefix to be compared lexicographically without padding.  (The check
// digit position is excluded: it carries no range information, and for ISBN-10
// inputs it may hold a non-digit 'X'.)
type compiledRule struct {
	prefix   string // EAN+group digits, no hyphen, e.g. "97899932"
	group    string // group digits only, e.g. "99932"
	minBound string // clamped lower bound, e.g. "10000"
	maxBound string // clamped upper bound, same width as minBound
	pubLen   int    // publisher element length; 0 marks an unassigned range
}

// NewHyphenator compiles ISBN range data into an optimized lookup structure.
// It assumes the underlying ISBNRanges data is valid per the official ISBN range message.
// If the data is malformed, hyphenation may return an error.
func NewHyphenator(data ISBNRanges) (*Hyphenator, error) {
	h := &Hyphenator{}

	for _, g := range data.ISBNRangeMessage.RegistrationGroups.Group {
		parts := strings.SplitN(g.Prefix, "-", 2)
		if len(parts) != 2 {
			continue
		}
		ean, grp := parts[0], parts[1]
		if ean == "" || grp == "" {
			continue
		}
		prefix := ean + grp
		if len(prefix) >= isbn13BodyDigits || !isDigits(prefix) {
			continue
		}

		for _, r := range g.Rules.Rule {
			if rule, ok := compileRule(prefix, grp, r); ok {
				h.rules = append(h.rules, rule)
			}
		}
	}

	if len(h.rules) == 0 {
		return nil, fmt.Errorf("%w", ErrMissingISBNRangeData)
	}

	slices.SortFunc(h.rules, func(a, b compiledRule) int {
		return cmp.Or(cmp.Compare(a.prefix, b.prefix), cmp.Compare(a.minBound, b.minBound))
	})

	return h, nil
}

// DefaultHyphenator returns a hyphenator which uses built-in range data.
//
// This function panics if the built-in ranges cannot be parsed or compiled
// (which should never happen in practice).
func DefaultHyphenator() *Hyphenator {
	return defaultHyphenator()
}

// newDefaultHyphenator creates a hyphenator from the built-in range data.
//
// This function panics if the built-in ranges cannot be parsed or compiled
// (which should never happen in practice).
func newDefaultHyphenator() *Hyphenator {
	ranges := must(ParseISBNRangesJSON(isbnRangesJSON))
	return must(NewHyphenator(ranges))
}

// LoadHyphenator fetches ISBN range data (falling back on built-in data) and
// creates a new Hyphenator.
func LoadHyphenator(ctx context.Context, client HTTPClient) (*Hyphenator, error) {
	isbnRanges, err := FetchISBNRanges(ctx, client)
	if err != nil {
		return nil, err
	}
	return NewHyphenator(isbnRanges)
}

// compileRule converts a raw range rule into a compiledRule, clamping the
// bounds to the number of data digits actually available after the prefix in
// an ISBN-13, excluding the trailing check digit.
// It reports false for malformed or unmatchable rules.
func compileRule(prefix, group string, r RangeRule) (compiledRule, bool) {
	bounds := strings.SplitN(r.Range, "-", 2)
	if len(bounds) != 2 {
		return compiledRule{}, false
	}
	loS, hiS := bounds[0], bounds[1]
	if loS == "" || len(loS) != len(hiS) || !isDigits(loS) || !isDigits(hiS) || loS > hiS {
		return compiledRule{}, false
	}
	pubLen, err := strconv.Atoi(r.Length)
	if err != nil || pubLen < 0 {
		return compiledRule{}, false
	}

	if width := isbn13BodyDigits - len(prefix); width < len(loS) {
		hiS = hiS[:width]
		var ok bool
		if loS, ok = clampLowerBound(loS, width); !ok || loS > hiS {
			return compiledRule{}, false
		}
	}

	return compiledRule{prefix: prefix, group: group, minBound: loS, maxBound: hiS, pubLen: pubLen}, true
}

// isDigits reports whether s is non-empty and consists solely of ASCII digits.
func isDigits(s string) bool {
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return s != ""
}

// clampLowerBound truncates lo to width digits, rounding up when the truncated
// tail is non-zero. It reports false if rounding up overflows.
func clampLowerBound(lo string, width int) (string, bool) {
	head, tail := lo[:width], lo[width:]
	if strings.Trim(tail, "0") == "" {
		return head, true
	}
	return incrementDigits(head)
}

// incrementDigits returns the decimal digit string s incremented by one,
// keeping the same width. It reports false on overflow (all nines).
func incrementDigits(s string) (string, bool) {
	b := []byte(s)
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] != '9' {
			b[i]++
			return string(b), true
		}
		b[i] = '0'
	}
	return "", false
}

// cmpRule orders a compiled rule relative to a clean 12-digit ISBN body for
// binary search: it compares the rule prefix against the corresponding leading
// digits of the ISBN, then compares the digits following the prefix against the
// rule bounds. Registration group prefixes are prefix-free, so slicing the ISBN
// to the rule's prefix length yields a consistent total order.
func cmpRule(rule compiledRule, body string) int {
	if n := cmp.Compare(rule.prefix, body[:len(rule.prefix)]); n != 0 {
		return n
	}
	chunk := body[len(rule.prefix) : len(rule.prefix)+len(rule.minBound)]
	if chunk < rule.minBound {
		return 1
	}
	if chunk > rule.maxBound {
		return -1
	}
	return 0
}

// findRule locates the hyphenation rule covering the given clean ISBN body digits.
func (h *Hyphenator) findRule(body string) (compiledRule, bool) {
	if len(body) != isbn13BodyDigits {
		return compiledRule{}, false
	}
	if idx, found := slices.BinarySearchFunc(h.rules, body, cmpRule); found {
		return h.rules[idx], true
	}
	return compiledRule{}, false
}

// Hyphenate applies ISBN hyphenation rules to the given ISBN.
//
// It assumes the underlying ISBNRanges data is valid per the official ISBN
// range message.  If the range data is malformed, hyphenation may return an
// error.
func (h *Hyphenator) Hyphenate(isbn ISBN) (string, error) {
	if h == nil || len(h.rules) == 0 {
		return "", ErrMissingISBNRangeData
	}

	body, checkDigit := isbn.Split()
	if body == "" {
		return "", fmt.Errorf("%w: %s", ErrInvalidISBN, isbn)
	}
	if isbn.Is10() {
		body = ISBN10Prefix + body
	}

	rule, found := h.findRule(body)
	if !found {
		return "", fmt.Errorf("%w: %s", ErrRegistrationGroupNotFound, isbn)
	}
	if rule.pubLen == 0 {
		return "", fmt.Errorf("%w: %s", ErrPublisherRangeNotFound, isbn)
	}

	publisherPart := body[len(rule.prefix):]
	if rule.pubLen >= len(publisherPart) {
		return "", fmt.Errorf("%w: %s: length %d", ErrInvalidPublisherLength, isbn, rule.pubLen)
	}

	publisher := publisherPart[:rule.pubLen]
	publication := publisherPart[rule.pubLen:]

	if isbn.Is10() {
		return rule.group + "-" + publisher + "-" + publication + "-" + checkDigit, nil
	}
	ean := rule.prefix[:len(rule.prefix)-len(rule.group)]
	return ean + "-" + rule.group + "-" + publisher + "-" + publication + "-" + checkDigit, nil
}
