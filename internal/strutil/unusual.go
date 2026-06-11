package strutil

import (
	"errors"
	"fmt"
	"unicode"
)

var (
	ErrUnusualRune = errors.New("unusual character")
)

// IsUnusualString returns true if the string contains any unusual runes for Latinate text.
// This is used to detect potential encoding issues or non-textual content.
func IsUnusualString(s string) bool {
	return AssessRunes(s) != nil
}

// AssessRunes returns true if the string contains any unusual runes for Latinate text.
// This is used to detect potential encoding issues or non-textual content.
func AssessRunes(s string) error {
	count := 1
	for _, r := range s {
		if err := AssessRune(r); err != nil {
			return fmt.Errorf("character %d: %w", count, err)
		}
		count++
	}
	return nil
}

// IsUnusualRune returns true if the rune is considered unusual for normal Latinate text.
// Letters outside common Unicode blocks are flagged as unusual.
func IsUnusualRune(r rune) bool {
	return AssessRune(r) != nil
}

// AssessRune returns true if the rune is considered unusual for normal Latinate text.
// Letters outside common Unicode blocks are flagged as unusual.
func AssessRune(r rune) error {
	switch {
	case r == '\u0000', r == unicode.ReplacementChar:
		return fmt.Errorf("%w %q", ErrUnusualRune, r)
	case unicode.IsSpace(r): // category Z
		return nil
	}
	if err := AssessUnicodeBlock(r); err != nil { // unusual unicode blocks
		return err
	}
	switch {
	case unicode.IsLetter(r): // category L
		return nil
	case unicode.IsNumber(r): // category N
		return nil
	case unicode.IsPunct(r): // category P
		return nil
	case unicode.IsSymbol(r): // category S
		return nil
	case unicode.In(r, unicode.Latin): // Latin script
		return nil
	default:
		return fmt.Errorf("%w %+q", ErrUnusualRune, r)
	}
}

// IsUnusualUnicodeBlock returns true if the rune is outside common Latinate text Unicode blocks.
// Allows Basic Latin, Latin-1, Latin Extended, General Punctuation, and common presentation forms.
func IsUnusualUnicodeBlock(r rune) bool {
	return AssessUnicodeBlock(r) != nil
}

// AssessUnicodeBlock returns true if the rune is outside common Latinate text Unicode blocks.
// Allows Basic Latin, Latin-1, Latin Extended, General Punctuation, and common presentation forms.
func AssessUnicodeBlock(r rune) error {
	switch {
	case r <= 0x024F: // Basic Latin, Latin-1 Supplement, Latin Extended-A, Latin Extended-B
		return nil
	case r >= 0x2000 && r <= 0x206F: // General Punctuation
		return nil
	case r >= 0x2070 && r <= 0x209F: // Superscripts and Subscripts
		return nil
	case r >= 0x2150 && r <= 0x218F: // Number Forms
		return nil
	case r >= 0xFB00 && r <= 0xFB4F: // Alphabetic Presentation Forms
		return nil
	case r >= 0xFE50 && r <= 0xFE6F: // Small Form Variants
		return nil
	case r >= 0xFF00 && r <= 0xFFEF: // Halfwidth and Fullwidth Forms
		return nil
	default:
		return fmt.Errorf("%w %+q", ErrUnusualRune, r)
	}
}

// IsArtifact returns true if the string appears to be structural (e.g. HTML or
// CSS) or other non-textual fragments common in scraped Readerware data.
func IsArtifact(s string) bool {
	var (
		first     rune
		colons    int
		semis     int
		hashes    int
		wordCount int
		inWord    bool
	)

	for _, r := range s {
		// 1. Literal braces are almost never valid in metadata fields
		if r == '{' || r == '}' {
			return true
		}

		if unicode.IsSpace(r) {
			inWord = false
			continue
		}

		if first == 0 {
			first = r
		}

		if !inWord {
			wordCount++
			inWord = true
		}

		switch r {
		case ':':
			colons++
		case ';':
			semis++
		case '#':
			hashes++
		}
	}

	// 2. High density of CSS-like structural markers.
	// Common in: "color: #333; font-family: ...;"
	if colons >= 2 && semis >= 1 {
		return true
	}

	// 3. ID/Hex color markers in short strings
	// Common in: "#productDescription" or "#000"
	if hashes >= 1 && wordCount <= 3 && colons == 0 {
		// This might be a bit aggressive, but titles rarely start with '#'
		// unless it's a "Number" marker which usually has spaces or other context.
		// For now, focus on the CSS selector case.
		if first == '#' {
			return true
		}
	}

	return false
}
