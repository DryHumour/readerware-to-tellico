package normalize

import (
	"fmt"
	"iter"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/DryHumour/readerware-to-tellico/internal/strutil"
)

const (
	// AnonymousMarker is a pre-defined marker used to indicate an anonymous value.
	AnonymousMarker = "<anonymous>"
	// UnknownMarker is a pre-defined marker used to indicate an unknown value.
	UnknownMarker = "<unknown>"
)

// must panics if err is non-nil, otherwise returns v.
// This is a convenience helper for initializing global variables that must not fail.
func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// chain combines multiple iter.Seq[T] into a single iter.Seq[T].
func chain[I iter.Seq[T], T any](iters ...I) I {
	return func(yield func(T) bool) {
		for _, it := range iters {
			for v := range it {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// chain2 combines multiple iter.Seq2[K,V] into a single iter.Seq2[K,V].
func chain2[I iter.Seq2[K, V], K, V any](iters ...I) I {
	return func(yield func(K, V) bool) {
		for _, it := range iters {
			for k, v := range it {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}

// uniq returns a sequence that yields unique values from the input sequence.
// Order is preserved based on first occurrence.
func uniq[T comparable](s iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		seen := make(map[T]struct{})
		for v := range s {
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				if !yield(v) {
					return
				}
			}
		}
	}
}

// uniqNZ returns a sequence that yields unique non-zero values from the input sequence.
// Order is preserved based on first occurrence.
func uniqNZ[T comparable](s iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		var zero T
		seen := map[T]struct{}{zero: struct{}{}}
		for v := range s {
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				if !yield(v) {
					return
				}
			}
		}
	}
}

// scrub removes matches of the provided regexp from the result value.
//
// The input Result.Value has no preconditions; it need not even be UTF-8.
//
// If a match is removed, the remaining parts of the string are joined
// with a single space and re-squeezed to ensure no double spaces are
// introduced by the removal.
func scrub(r Result, re *regexp.Regexp) Result {
	if re != nil {
		r.Value = strutil.ReplaceAndSqueeze(r.Value, re)
	} else {
		r.Value = strutil.Squeeze(r.Value)
	}
	return r
}

// filter evaluates the Result against the provided keep and discard regexps.
//
// The input Result.Value is expected to be sanitized as if by Scrub.
// In particular, it should have had its whitespace trimmed and squeezed as if
// by strutil.Squeeze.
//
// If keepRE matches, Result.Keep is returned
// If discardRE matches, the Result.Discard is returned.
func filter(r Result, keepRE, discardRE *regexp.Regexp) Result {
	if keepRE != nil && keepRE.MatchString(r.Value) {
		return r.Keep()
	}
	if discardRE != nil && discardRE.MatchString(r.Value) {
		return r.Discard()
	}
	return r
}

// extract extracts matches from the given string using the provided regexp and canonical map.
// It returns the remaining string after extracting matches and the list of extracted matches.
//
// The input string is expected to be sanitized as if by Scrub.
// In particular, it should have had its whitespace trimmed and squeezed as if
// by strutil.Squeeze.
//
// # Invariants
//
// The re and canonical map MUST be co-produced (e.g., by canonicalRE) such that
// every possible match returned by the regexp is guaranteed to exist as a
// lowercase key in the canonical map. Failure to maintain this symmetry will
// result in undefined behaviour (panic).
func extract(s string, re *regexp.Regexp, canonical map[string]string) (remaining string, matches []string) {
	if re == nil {
		return s, nil
	}

	locs := re.FindAllStringIndex(s, -1)
	if len(locs) == 0 {
		return s, nil
	}

	b := make([]byte, 0, len(s))
	cursor := 0

	for i := 0; i < len(s); i++ {
		// 1. Are we at the start of a match?
		if cursor < len(locs) && i == locs[cursor][0] {
			match := s[locs[cursor][0]:locs[cursor][1]]

			found, ok := canonical[strings.ToLower(match)]
			if !ok {
				panic(fmt.Sprintf("canonical map missing entry for: %+q", match)) // FIXME(nschelle) do something more elegant in the production code
			}
			matches = append(matches, found)

			// Fast-forward 'i' to the end of the match.
			// (-1 because the loop's i++ will advance it the rest of the way)
			i = locs[cursor][1] - 1
			cursor++
			continue
		}

		c := s[i]

		// 2. The Spacing State Machine (UTF-8 safe)
		switch c {
		case ',', '.', ';', ':', '-', ')', ']', '\'':
			// If we hit punctuation and the previous char was a space,
			// "backspace" the space by overwriting it with the punctuation.
			// Thus "Smith, John (ed.), Jr." ⟶ "Smith, John, Jr."
			if len(b) > 0 && b[len(b)-1] == ' ' {
				b[len(b)-1] = c
				continue
			}
		case ' ':
			// Ignore leading spaces or consecutive spaces
			if len(b) == 0 || b[len(b)-1] == ' ' {
				continue
			}
		}

		// Write the valid byte (for all non-skipped characters)
		b = append(b, c)
	}

	// 3. Cleanup: Drop any trailing space left at the absolute end of the string
	if len(b) > 0 && b[len(b)-1] == ' ' {
		b = b[:len(b)-1]
	}

	// string(b) safely allocates exactly once for the return value
	return string(b), matches
}

// audit performs various audit checks on the result.
//
// The input Result.Value is expected to be sanitized as if by Scrub.
// In particular, it should have had its whitespace trimmed and squeezed as if
// by strutil.Squeeze.
//
// The explanations map maps lowercase audit patterns to explanatory strings.
func audit(r Result, re *regexp.Regexp, explanations map[string]string) Result {
	// 1. Check audit patterns.

	if re != nil {
		locs := re.FindAllStringIndex(r.Value, -1)
		if len(locs) > 0 {
			r.RequiresAudit = true
			for _, loc := range locs {
				matched := r.Value[loc[0]:loc[1]]
				reason := "[audit]"
				if expl, ok := explanations[strings.ToLower(matched)]; ok {
					reason = expl
				}
				r.AuditReasons = append(r.AuditReasons, AuditReason(reason, matched))
			}
		}
	}

	// 2. Check for unusual runes.

	if err := strutil.AssessRunes(r.Value); err != nil {
		r.RequiresAudit = true
		r.AuditReasons = append(r.AuditReasons, AuditReason(fmt.Sprintf("[unusual] Contains unusual characters: %s", err.Error()), r.Value))
	}

	// 3. Check for results that don't start with a capital letter or number.

	if len(r.Value) != 0 {
		firstRune, _ := utf8.DecodeRuneInString(r.Value)
		if !unicode.IsUpper(firstRune) && !unicode.IsDigit(firstRune) {
			r.RequiresAudit = true
			r.AuditReasons = append(r.AuditReasons, AuditReason("[capital] Does not start with an uppercase letter or digit", r.Value))
		}
	}

	// 4. Check for results containing certain somewhat unlikely symbols.

	if strings.ContainsAny(r.Value, "@#$%^*=<>~`") {
		r.RequiresAudit = true
		r.AuditReasons = append(r.AuditReasons, AuditReason("[symbols] Contains unusual symbols", r.Value))
	}

	// 5. Check for <anonymous> and <unknown> markers.

	if slices.Contains(r.Markers, AnonymousMarker) {
		r.RequiresAudit = true
		r.AuditReasons = append(r.AuditReasons, AuditReason("[anonymous] Marker for "+AnonymousMarker+" present", r.Value))
	}
	if slices.Contains(r.Markers, UnknownMarker) {
		r.RequiresAudit = true
		r.AuditReasons = append(r.AuditReasons, AuditReason("[unknown] Marker for "+UnknownMarker+" present", r.Value))
	}

	return r
}

// AuditReason formats an audit reason with the associated value.
// Values longer than 25 runes are truncated with an ellipsis.
func AuditReason(reason, value string) string {
	const maxRunes = 25
	var nRunes int
	for i := range value {
		if nRunes >= maxRunes {
			u := strconv.QuoteToASCII(value[:i])
			return reason + ": " + u[:len(u)-1] + `…"`
		}
		nRunes++
	}
	return reason + ": " + strconv.QuoteToASCII(value)
}

// AuditReasonWithMatch formats an audit reason with the associated value and matched substring.
// Values longer than 25 runes are truncated with an ellipsis.
func AuditReasonWithMatch(reason, value string, match []int) string {
	return AuditReason(fmt.Sprintf("%s: %+q", reason, value[match[0]:match[1]]), value)
}
