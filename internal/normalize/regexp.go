package normalize

import (
	"bytes"
	"fmt"
	"iter"
	"maps"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// compileRE compiles a regular expression from a sequence of regular expression patterns.
// The patterns are joined with alternation (|) in the order they are provided.
// The flags string is used for regexp flags (e.g., "i" for case-insensitive).
// The flags string may be empty, in which case no flags are applied to the regexp.
func compileRE(flags string, prefix string, patterns iter.Seq[string], suffix string) (*regexp.Regexp, error) {
	var sb strings.Builder
	first := true
	for p := range patterns {
		if p == "" {
			continue
		}
		if first {
			sb.WriteString(prefix)
			sb.WriteString("(?")
			sb.WriteString(flags)
			sb.WriteString(":")
		} else {
			sb.WriteByte('|')
		}
		sb.WriteString(p)
		first = false
	}
	if first {
		// No non-empty patterns provided
		return nil, nil
	}
	sb.WriteByte(')')
	sb.WriteString(suffix)
	re, err := regexp.Compile(sb.String())
	if err != nil {
		return nil, fmt.Errorf("failed to compile regexp: %w", err)
	}
	return re, nil
}

// keepRE compiles a regular expression that matches any of the given keep phrases.
// The match is case-sensitive, full string.
func keepRE(keep []string) (*regexp.Regexp, error) {
	re, err := compileRE("", `^`, uniqNZ(slices.Values(keep)), `$`)
	if err != nil {
		return nil, fmt.Errorf("failed to compile keep patterns: %w", err)
	}
	return re, nil
}

// discardRE compiles a regular expression that matches any of the given discard phrases.
// The match is case-insensitive, full string.
func discardRE(discard []string) (*regexp.Regexp, error) {
	re, err := compileRE("i", `^`, uniqNZ(slices.Values(discard)), `$`)
	if err != nil {
		return nil, fmt.Errorf("failed to compile discard patterns: %w", err)
	}
	return re, nil
}

// basicRE compiles a regular expression that matches any of the given phrases.
// The match is case-insensitive, substring.
// It uses word boundaries for phrases that start/end with word characters.
func basicRE(kind string, phrases []string) (*regexp.Regexp, error) {
	re, err := compileRE("i", ``, boundaryLiterals(uniqNZ(slices.Values(phrases))), ``)
	if err != nil {
		return nil, fmt.Errorf("failed to compile %s patterns: %w", kind, err)
	}
	return re, nil
}

// canonicalRE compiles a regular expression and a corresponding canonical map.
//
// Every lowercase key in the returned canonical map is guaranteed to be
// matched by the returned regexp. This co-production is required by the
// extract() engine to prevent panics during lookup.
//
// The match is case-insensitive and restricted to word boundaries for
// word-based phrases.
func canonicalRE(kind string, m map[string]string) (re *regexp.Regexp, canonical map[string]string, err error) {
	canonical = make(map[string]string, len(m))
	for k, v := range m {
		if k == "" {
			continue
		}
		canonical[strings.ToLower(k)] = v
	}
	re, err = compileRE("i", ``, boundaryLiterals(maps.Keys(canonical)), ``)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to compile %s patterns: %w", kind, err)
	}
	return re, canonical, nil
}

// auditRE compiles a regular expression and a corresponding explanation map.
//
// Every lowercase key in the returned explanation map is guaranteed to be
// matched by the returned regexp. The map values are explanatory strings
// for audit reporting.
//
// The match is case-insensitive and restricted to word boundaries for
// word-based phrases.
func auditRE(kind string, m map[string]string) (re *regexp.Regexp, explanations map[string]string, err error) {
	explanations = make(map[string]string, len(m))
	for k, v := range m {
		if k == "" {
			continue
		}
		explanations[strings.ToLower(k)] = v
	}
	re, err = compileRE("i", ``, boundaryLiterals(maps.Keys(explanations)), ``)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to compile %s patterns: %w", kind, err)
	}
	return re, explanations, nil
}

// boundaryLiterals returns an iterator that yields each literal with word boundaries applied.
// If the literal starts with a word character, a leading \b is added.
// If the literal ends with a word character, a trailing \b is added.
// For other cases, it is assumed that the character alone is sufficient as a boundary.
func boundaryLiterals(literals iter.Seq[string]) iter.Seq[string] {
	return func(yield func(string) bool) {
		var b bytes.Buffer
		for p := range literals {
			if p == "" {
				continue
			}
			b.Grow(len(p)*2 + 4)
			if r, _ := utf8.DecodeRuneInString(p); unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
				b.WriteString(`\b`)
			}
			b.WriteString(regexp.QuoteMeta(p))
			if r, _ := utf8.DecodeLastRuneInString(p); unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
				b.WriteString(`\b`)
			}
			if !yield(b.String()) {
				return
			}
			b.Reset()
		}
	}
}

// quotemeta returns an iterator that yields each literal with regexp.QuoteMeta applied.
func quotemeta(literals iter.Seq[string]) iter.Seq[string] {
	return func(yield func(string) bool) {
		for pattern := range literals {
			if !yield(regexp.QuoteMeta(pattern)) {
				return
			}
		}
	}
}

// abbrevPattern generates a search pattern and a canonical replacement key for suffixes.
//
// To avoid misidentifying initials as abbreviations, this function only
// handles segments containing at least two letters before a period (e.g.,
// "Ph.D." is handled, but "M.D." or "A.B.C." are ignored).
//
// It returns:
// - key: the original string with a single space inserted wherever \s* was added.
// - pattern: the regex pattern with \s* injected.
// - ok: true if at least one \s* sequence was injected.
func abbrevPattern(s string) (key string, pattern string, ok bool) {
	// Fast path: if the configured suffix already contains spaces,
	// we do not need to auto-generate a space-collapsed version.
	if s == "" || strings.ContainsRune(s, ' ') {
		return "", "", false
	}

	var (
		keyB, patB strings.Builder
		cursor     = 0
		letters    = 0
		injected   = false
	)

	for i, r := range s {
		switch {
		case r == '.' && letters >= 2 && i < len(s)-1:
			rest := i + 1
			if !injected {
				keyB.Grow(len(s) + 5)
				patB.Grow(len(s)*2 + 10)
				injected = true
			}
			keyB.WriteString(s[cursor:rest])
			keyB.WriteByte(' ')
			patB.WriteString(regexp.QuoteMeta(s[cursor:rest]))
			patB.WriteString(`\s*`)
			cursor = rest
			letters = 0
		case r == '.':
			letters = 0
		case unicode.IsLetter(r):
			letters++
		default:
			letters = 0
		}
	}

	// If we never hit the injection case (e.g., "M.D."), return false.
	if !injected {
		return "", "", false
	}

	// Flush the remaining unwritten characters.
	if cursor < len(s) {
		keyB.WriteString(s[cursor:])
		patB.WriteString(regexp.QuoteMeta(s[cursor:]))
	}

	return keyB.String(), patB.String(), true
}

// spacePattern generates a regex pattern that matches the given string with optional spaces
// injected after every period.
//
// Unlike abbrevPattern, this function does not have a minimum letter requirement
// and is primarily used for identifying ambiguous abbreviations (like "M. D.")
// during auditing.
func spacePattern(s string) string {
	// Fast path: if the configured suffix already contains spaces,
	// we do not need to auto-generate a space-collapsed version.
	if s == "" || strings.ContainsRune(s, ' ') {
		return regexp.QuoteMeta(s)
	}

	var b strings.Builder
	b.Grow(len(s)*2 + 10)

	cursor := 0
	for i, r := range s {
		if i >= len(s)-1 {
			break
		}
		if r != '.' {
			continue
		}
		rest := i + 1
		b.WriteString(regexp.QuoteMeta(s[cursor:rest]))
		b.WriteString(`\s*`)
		cursor = rest
	}

	// Flush the remaining unwritten characters.
	if cursor < len(s) {
		b.WriteString(regexp.QuoteMeta(s[cursor:]))
	}

	return b.String()
}

// patternsByLen returns an iterator that yields the values from the patterns map, sorted by key length in descending order.
func patternsByLen(patterns iter.Seq2[string, string]) iter.Seq[string] {
	return func(yield func(string) bool) {
		type pair struct{ K, V string }
		items := []pair{}
		for k, v := range patterns {
			items = append(items, pair{K: k, V: v})
		}
		if len(items) == 0 {
			return
		}
		slices.SortStableFunc(items, func(a, b pair) int { return len(b.K) - len(a.K) })
		for _, item := range items {
			if !yield(item.V) {
				return
			}
		}
	}
}

// boundaryPatternsByLen returns an iterator that yields the values from the patterns map, sorted by key length in descending order.
func boundaryPatternsByLen(prefix string, patterns iter.Seq2[string, string], suffix string) iter.Seq[string] {
	return func(yield func(string) bool) {
		var b bytes.Buffer
		type pair struct{ K, V string }
		items := []pair{}
		for k, v := range patterns {
			if k == "" || v == "" {
				continue
			}
			b.Grow(len(prefix) + len(v) + len(suffix))
			if r, _ := utf8.DecodeRuneInString(k); unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
				b.WriteString(prefix)
			}
			b.WriteString(v)
			if r, _ := utf8.DecodeLastRuneInString(k); unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
				b.WriteString(suffix)
			}
			items = append(items, pair{K: k, V: b.String()})
			b.Reset()
		}
		if len(items) == 0 {
			return
		}
		slices.SortStableFunc(items, func(a, b pair) int { return len(b.K) - len(a.K) })
		for _, item := range items {
			if !yield(item.V) {
				return
			}
		}
	}
}
