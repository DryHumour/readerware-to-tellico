// Package strutil provides string utility functions for text processing,
// formatting, and escaping used across the readerware-to-tellico codebase.
package strutil

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/jaytaylor/html2text"
)

var (
	// dimensionsSepRE is a regular expression to match dimension separators.
	dimensionsSepRE = regexp.MustCompile(`\s*x\s*`)
	// ratingNumberRE is a regular expression to match rating numbers (e.g., "4.5" or "5").
	ratingNumberRE = regexp.MustCompile(`\d+(?:\.\d+)?`)
	// htmlDetectorRE looks for valid tag structures or HTML entities.
	// Branch 1: < followed by a letter or slash, then anything until >
	// Branch 2: & followed by letters/numbers/#, ending in ;
	htmlDetectorRE = regexp.MustCompile(`(?i)<[a-z/][^>]*>|&[#a-z0-9]+;`)
)

// Checkbox returns "true" if the string is truthy, empty otherwise.
// Truthy values: "true", "t", "yes", "y", "1" (case-insensitive).
func Checkbox(s string) string {
	switch strings.ToLower(Squeeze(s)) {
	case "true", "t", "yes", "y", "1":
		return "true"
	default:
		return ""
	}
}

// ContainsHTML checks if a string contains HTML tags or entities.
func ContainsHTML(s string) bool {
	return htmlDetectorRE.MatchString(s)
}

// Dimensions replaces "x" separators with the Unicode multiplication sign "×"
// for proper dimension formatting (e.g., "6 x 9" → "6 × 9").
func Dimensions(s string) string {
	return dimensionsSepRE.ReplaceAllLiteralString(s, "×")
}

// ExtractAndSqueeze replaces regex matches, collapses whitespace, and returns the matches.
func ExtractAndSqueeze(s string, re *regexp.Regexp) (string, []string) {
	return engineRegexSqueeze(s, re, false, true)
}

// ExtractAndSqueezePreserveNewlines replaces matches, preserves newlines, and returns matches.
func ExtractAndSqueezePreserveNewlines(s string, re *regexp.Regexp) (string, []string) {
	return engineRegexSqueeze(s, re, true, true)
}

// engineRegexSqueeze is the master streaming engine for all regex-mutated strings.
func engineRegexSqueeze(s string, re *regexp.Regexp, preserveNL, extract bool) (string, []string) {
	loc := re.FindStringIndex(s)
	if loc == nil {
		// Fast path: fallback to the pure string engine
		return squeeze(s, preserveNL), nil
	}

	var (
		b       strings.Builder
		matches []string
	)

	b.Grow(len(s))
	if extract {
		matches = make([]string, 0, 4) // Only allocate the slice if we need it
	}

	const (
		stateLeadingWS = iota
		stateText
		stateWS
	)

	st := stateLeadingWS
	nlCount := 0

	// processRune acts as our stream processor
	processRune := func(r rune) {
		switch st {
		case stateLeadingWS:
			if !unicode.IsSpace(r) {
				b.WriteRune(r)
				st = stateText
			}

		case stateText:
			if unicode.IsSpace(r) {
				st = stateWS
				if r == '\n' {
					nlCount = 1
				} else {
					nlCount = 0
				}
			} else {
				b.WriteRune(r)
			}

		case stateWS:
			if unicode.IsSpace(r) {
				if r == '\n' {
					nlCount++
				}
			} else {
				if preserveNL && nlCount > 0 {
					for j := 0; j < nlCount; j++ {
						b.WriteByte('\n')
					}
				} else {
					b.WriteByte(' ')
				}
				b.WriteRune(r)
				st = stateText
			}
		}
	}

	offset := 0
	for {
		if loc == nil {
			// Process remaining text after the final match
			for _, r := range s[offset:] {
				processRune(r)
			}
			break
		}

		start, end := offset+loc[0], offset+loc[1]

		// Process text BEFORE the match
		for _, r := range s[offset:start] {
			processRune(r)
		}

		// CONDITIONAL EXTRACTION: Only clone and append if requested!
		if extract {
			matches = append(matches, strings.Clone(s[start:end]))
		}

		// The regex match itself acts as a space
		processRune(' ')

		// Advance offset and look for the next match
		offset = end
		loc = re.FindStringIndex(s[offset:])
	}

	return b.String(), matches
}

// HTMLToText converts HTML to plain text using html2text.
// Falls back to squeezing whitespace if conversion fails.
func HTMLToText(s string) string {
	if ContainsHTML(s) {
		if out, err := html2text.FromString(s); err == nil {
			return out
		}
	}
	return SqueezePreserveNewlines(s)
}

// JoinParts joins non-empty strings with a single space.
// This differs from a simple strings.Join because it skips empty strings.
func JoinParts(parts ...string) string {
	var b strings.Builder
	for _, p := range parts {
		if p != "" {
			if b.Len() > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(p)
		}
	}
	return b.String()
}

// Keywords splits a string by commas, semicolons, or slashes and squeezes each part,
// then rejoins with semicolons for Tellico keyword format.
func Keywords(s string) string {
	var result []string
	for s := range strings.FieldsFuncSeq(s, func(r rune) bool { return r == ',' || r == ';' || r == '/' }) {
		result = append(result, Squeeze(s))
	}
	return strings.Join(result, ";")
}

// Paragraphs converts newlines to <br/> tags.
// This is useful for Tellico paragraph (type="2") fields.
func Paragraphs(s string) string {
	return strings.ReplaceAll(s, "\n", "<br/>")
}

// Price returns the price if it is not "0.00", empty otherwise.
// Used to suppress zero-valued prices in the output.
func Price(s string) string {
	s = Squeeze(s)
	if s == "0.00" {
		return ""
	}
	return s
}

// Rating1to5 returns the rating as an integer between 1 and 5, empty otherwise.
// Extracts the first numeric sequence and rounds to nearest integer.
// Values outside [1, 5] range are rejected.
func Rating1to5(s string) string {
	m := ratingNumberRE.FindString(s)
	if m == "" {
		return ""
	}
	f, err := strconv.ParseFloat(m, 64)
	if err != nil {
		return ""
	}
	if f < 1 || f > 5 {
		return ""
	}
	r := int(math.Floor(f + 0.5))
	if r < 1 || r > 5 {
		return ""
	}
	return strconv.Itoa(r)
}

// ReplaceAndSqueeze replaces regex matches with a space and collapses whitespace.
func ReplaceAndSqueeze(s string, re *regexp.Regexp) string {
	cleaned, _ := engineRegexSqueeze(s, re, false, false)
	return cleaned
}

// ReplaceAndSqueezePreserveNewlines does the same, but preserves interior newlines.
func ReplaceAndSqueezePreserveNewlines(s string, re *regexp.Regexp) string {
	cleaned, _ := engineRegexSqueeze(s, re, true, false)
	return cleaned
}

// SplitList splits a Readerware multi-value field on semicolons and slashes,
// squeezes whitespace from each element, and returns the non-empty results as a slice.
// The string "N/A" is kept together and not split.
// Commas are intentionally not treated as separators because they are significant
// in "Last, First" name formatting.
func SplitList(s string) []string {
	var results []string
	start := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case 'N', 'n':
			// keep "N/A" unsplit
			if i < len(s)-2 && s[i+1] == '/' && (s[i+2] == 'A' || s[i+2] == 'a') {
				i += 2
			}
		case ';', '/':
			if squeezed := Squeeze(s[start:i]); squeezed != "" {
				results = append(results, squeezed)
			}
			start = i + 1
		}
	}
	// Grab any trailing field
	if start < len(s) {
		if squeezed := Squeeze(s[start:]); squeezed != "" {
			results = append(results, squeezed)
		}
	}
	return results
}

// Squeeze trims leading and trailing whitespace and collapses all interior
// runs of whitespace into a single space.
func Squeeze(s string) string {
	return squeeze(s, false)
}

// SqueezePreserveNewlines trims leading and trailing whitespace. It collapses
// horizontal whitespace, but preserves the exact number of interior newlines
// found within any whitespace run.
func SqueezePreserveNewlines(s string) string {
	return squeeze(s, true)
}

// squeeze is the underlying state machine. If preserveNL is true, a run of
// whitespace containing newlines will be collapsed into those newlines rather
// than a single space.
func squeeze(s string, preserveNL bool) string {
	const (
		stateLeadingWS = iota
		stateScanText
		stateSPC
		stateTrailingWS
		stateText
		stateWS
	)
	var (
		st      = stateLeadingWS
		pfx     int // byte offset of start of text
		sfx     int // byte offset of start of ws suffix
		nlCount int // tracks newlines within the current whitespace run
		b       strings.Builder
	)

	for i, r := range s {
		switch st {
		case stateLeadingWS:
			switch {
			case unicode.IsSpace(r):
				// skip ws pfx
			default:
				// text: scan
				st = stateScanText
				pfx = i
			}

		case stateScanText:
			switch {
			case r == ' ':
				// single ASCII SPC (so far)
				st = stateSPC
				sfx = i
				nlCount = 0
			case unicode.IsSpace(r):
				// could be start of trailing ws
				st = stateTrailingWS
				sfx = i
				if r == '\n' {
					nlCount = 1
				} else {
					nlCount = 0
				}
			default:
				// text: continue scan
			}

		case stateSPC:
			switch {
			case unicode.IsSpace(r):
				// multiple ws: could be trailing ws
				st = stateTrailingWS
				if r == '\n' {
					nlCount = 1
				} else {
					nlCount = 0
				}
			default:
				// single ASCII SPC, text: scan
				st = stateScanText
				sfx = 0
			}

		case stateTrailingWS:
			switch {
			case unicode.IsSpace(r):
				// still trailing ws: continue
				if r == '\n' {
					nlCount++
				}
			default:
				// not trailing ws: must squeeze
				st = stateText
				b.Grow(len(s) - pfx)
				b.WriteString(s[pfx:sfx])

				if preserveNL && nlCount > 0 {
					for j := 0; j < nlCount; j++ {
						b.WriteByte('\n')
					}
				} else {
					b.WriteByte(' ')
				}

				b.WriteRune(r)
			}

		case stateText:
			switch {
			case unicode.IsSpace(r):
				// ws: skip runs
				st = stateWS
				if r == '\n' {
					nlCount = 1
				} else {
					nlCount = 0
				}
			default:
				// text: copy
				b.WriteRune(r)
			}

		case stateWS:
			switch {
			case unicode.IsSpace(r):
				// ws run: skip
				if r == '\n' {
					nlCount++
				}
			default:
				// text: copy
				st = stateText

				if preserveNL && nlCount > 0 {
					for j := 0; j < nlCount; j++ {
						b.WriteByte('\n')
					}
				} else {
					b.WriteByte(' ')
				}

				b.WriteRune(r)
			}
		}
	}

	switch {
	case st == stateLeadingWS:
		// string was empty or all WS
		return ""
	case b.Len() > 0:
		// had to squeeze
		return b.String()
	case sfx > pfx:
		// trim of both leading and trailing ws
		return s[pfx:sfx]
	case pfx > 0:
		// trim of leading ws (no trailing)
		return s[pfx:]
	}
	// no changes needed
	return s
}

// ToStringSlice converts an any value (expected []string or Sprig list []any) to []string.
func ToStringSlice(v any) ([]string, error) {
	switch val := v.(type) {
	case []string:
		return val, nil
	case []any:
		result := make([]string, 0, len(val))
		for _, item := range val {
			result = append(result, fmt.Sprintf("%v", item))
		}
		return result, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("expected list of strings, got %T", v)
	}
}

// ToStringStringMap converts an any value (expected map[string]string or Sprig dict map[string]any) to map[string]string.
func ToStringStringMap(v any) (map[string]string, error) {
	switch val := v.(type) {
	case map[string]string:
		return val, nil
	case map[string]any:
		result := make(map[string]string, len(val))
		for k, item := range val {
			result[k] = fmt.Sprintf("%v", item)
		}
		return result, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("expected dict of strings, got %T", v)
	}
}

// ToStringStringSliceMap converts an any value (expected map[string][]string or Sprig dict map[string]any) to map[string][]string.
// Handles both direct map[string][]string types and Sprig dict values where each value is a slice ([]string or []any).
func ToStringStringSliceMap(v any) (map[string][]string, error) {
	switch val := v.(type) {
	case map[string][]string:
		return val, nil
	case map[string]any:
		result := make(map[string][]string, len(val))
		for k, item := range val {
			slice, err := ToStringSlice(item)
			if err != nil {
				return nil, fmt.Errorf("key %q: %w", k, err)
			}
			result[k] = slice
		}
		return result, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("expected dict of string slices, got %T", v)
	}
}

// XMLEscape escapes XML special characters (<, >, &, ", ') to their entity equivalents.
// This is the canonical XML escaping location for the codebase.
//
// Note that xml.EscapeText also escapes newlines.  One would have to arrange
// to use either html.EscapeString or else xml.Encoder.EncodeToken(xml.CharData(p))
// to avoid that (or the full xml.Marshal mechanism).  Since Tellico renders
// paragraphs (type="2") fields using a Qt QTestDocument, that's all moot since
// we are forced to code <br/> into the text anyway.
func XMLEscape(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s)) // cannot fail (since bytes.Buffer)
	return b.String()
}
