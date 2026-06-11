package normalize

import (
	"maps"
	"regexp"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegexpFactories(t *testing.T) {
	t.Parallel()

	t.Run("compileRE joins patterns with flags", func(t *testing.T) {
		t.Parallel()
		re, err := compileRE("i", ``, slices.Values([]string{"abc", "def"}), ``)
		require.NoError(t, err)
		assert.True(t, re.MatchString("ABC"))
		assert.True(t, re.MatchString("DEF"))
		assert.False(t, re.MatchString("GHI"))
		assert.Equal(t, "(?i:abc|def)", re.String())
	})

	t.Run("compileRE quotes phrases and applies smart boundaries", func(t *testing.T) {
		t.Parallel()
		re, err := compileRE("i", ``, boundaryLiterals(slices.Values([]string{"a.b", "c(d)", "na"})), ``)
		require.NoError(t, err)
		assert.True(t, re.MatchString("a.b"))
		assert.True(t, re.MatchString("C(D)"))
		assert.True(t, re.MatchString("NA"))
		// Ensure dots and parens are treated as literals
		assert.False(t, re.MatchString("axb"))
		assert.False(t, re.MatchString("cd"))
		// Ensure word boundaries protect "na"
		assert.False(t, re.MatchString("Dana"))
	})

	t.Run("boundaryLiterals applies \b correctly", func(t *testing.T) {
		t.Parallel()
		// "na" -> "\bna\b" (word chars)
		// "(ed.)" -> "\(ed\.\)" (non-word chars)
		// ".net" -> "\.net\b" (starts non-word, ends word)
		lit := []string{"na", "(ed.)", ".net"}
		it := boundaryLiterals(slices.Values(lit))
		results := slices.Collect(it)

		assert.Contains(t, results, `\bna\b`)
		assert.Contains(t, results, `\(ed\.\)`)
		assert.Contains(t, results, `\.net\b`)
	})

	t.Run("boundaryPatternsByLen sorts and applies \b", func(t *testing.T) {
		t.Parallel()
		m := map[string]string{
			"Jr.": `Jr\.`,
			"Sr.": `Sr\.`,
			"a":   `a`,
		}
		it := boundaryPatternsByLen(`\b`, maps.All(m), `\b`)
		results := slices.Collect(it)

		// Should be sorted by length of key (Jr./Sr. before a)
		assert.Len(t, results, 3)
		assert.Contains(t, results, `\bJr\.`)
		assert.Contains(t, results, `\bSr\.`)
		assert.Contains(t, results, `\ba\b`)
	})

	t.Run("canonicalRE maintains symmetry", func(t *testing.T) {
		t.Parallel()
		m := map[string]string{
			"Ph.D.":    "PHD",
			"(signed)": "SIGNED",
		}
		re, canonical, err := canonicalRE("test", m)
		require.NoError(t, err)

		// Keys must be lowercased in the map
		assert.Contains(t, canonical, "ph.d.")
		assert.Contains(t, canonical, "(signed)")
		assert.Equal(t, "PHD", canonical["ph.d."])

		// Regexp must match the keys regardless of case
		assert.True(t, re.MatchString("PH.D."))
		assert.True(t, re.MatchString("(SIGNED)"))
	})
}

func TestAbbrevPattern(t *testing.T) {
	t.Parallel()

	t.Run("handles Ph.D. (2+ letters per segment)", func(t *testing.T) {
		t.Parallel()
		key, pattern, ok := abbrevPattern("Ph.D.")
		assert.True(t, ok)
		assert.Equal(t, "Ph. D.", key)
		assert.Equal(t, `Ph\.\s*D\.`, pattern)
	})

	t.Run("ignores M.D. (single letter segment)", func(t *testing.T) {
		t.Parallel()
		_, _, ok := abbrevPattern("M.D.")
		assert.False(t, ok)
	})

	t.Run("ignores single letter initials", func(t *testing.T) {
		t.Parallel()
		_, _, ok := abbrevPattern("A.B.C.")
		assert.False(t, ok)
	})

	t.Run("handles long abbreviations", func(t *testing.T) {
		t.Parallel()
		key, pattern, ok := abbrevPattern("Prof.Dr.")
		assert.True(t, ok)
		assert.Equal(t, "Prof. Dr.", key)
		assert.Equal(t, `Prof\.\s*Dr\.`, pattern)
	})

	t.Run("ignores strings with existing spaces", func(t *testing.T) {
		t.Parallel()
		_, _, ok := abbrevPattern("Ph. D.")
		assert.False(t, ok)
	})
}

func TestSpacePattern(t *testing.T) {
	t.Parallel()

	t.Run("injects space after every period including single letters", func(t *testing.T) {
		t.Parallel()
		pat := spacePattern("M.D.")
		re := regexp.MustCompile(pat)
		assert.True(t, re.MatchString("M.D."))
		assert.True(t, re.MatchString("M. D."))
		assert.Equal(t, `M\.\s*D\.`, pat)
	})
}

func TestKeepRE(t *testing.T) {
	t.Parallel()

	t.Run("compiles case-sensitive full string match", func(t *testing.T) {
		t.Parallel()
		re, err := keepRE([]string{"Prince", "King"})
		require.NoError(t, err)

		assert.True(t, re.MatchString("Prince"))
		assert.True(t, re.MatchString("King"))
		assert.False(t, re.MatchString("prince"))         // case-sensitive
		assert.False(t, re.MatchString("PrinceCharming")) // full string only
	})

	t.Run("handles empty input", func(t *testing.T) {
		t.Parallel()
		re, err := keepRE([]string{})
		require.NoError(t, err)
		assert.Nil(t, re)
	})
}

func TestDiscardRE(t *testing.T) {
	t.Parallel()

	t.Run("compiles case-insensitive full string match", func(t *testing.T) {
		t.Parallel()
		re, err := discardRE([]string{"n/a", "N/A"})
		require.NoError(t, err)

		assert.True(t, re.MatchString("n/a"))
		assert.True(t, re.MatchString("N/A"))
		assert.True(t, re.MatchString("N/A"))        // case-insensitive
		assert.False(t, re.MatchString("n/a value")) // full string only
	})

	t.Run("handles empty input", func(t *testing.T) {
		t.Parallel()
		re, err := discardRE([]string{})
		require.NoError(t, err)
		assert.Nil(t, re)
	})
}

func TestBasicRE(t *testing.T) {
	t.Parallel()

	t.Run("compiles case-insensitive substring match with boundaries", func(t *testing.T) {
		t.Parallel()
		re, err := basicRE("test", []string{"edition", "ed."})
		require.NoError(t, err)

		assert.True(t, re.MatchString("First Edition"))
		assert.True(t, re.MatchString("First edition")) // case-insensitive
		assert.True(t, re.MatchString("First ed."))     // substring
		assert.False(t, re.MatchString("editioned"))    // word boundary
	})

	t.Run("handles empty input", func(t *testing.T) {
		t.Parallel()
		re, err := basicRE("test", []string{})
		require.NoError(t, err)
		assert.Nil(t, re)
	})
}

func TestAuditRE(t *testing.T) {
	t.Parallel()

	t.Run("compiles audit patterns with explanations", func(t *testing.T) {
		t.Parallel()
		m := map[string]string{
			"et al.":  "Multiple authors",
			"unknown": "Unknown author",
		}
		re, explanations, err := auditRE("test", m)
		require.NoError(t, err)

		assert.True(t, re.MatchString("Smith et al."))
		assert.Equal(t, "Multiple authors", explanations["et al."])
		assert.Equal(t, "Unknown author", explanations["unknown"])
	})

	t.Run("lowercases keys in explanations map", func(t *testing.T) {
		t.Parallel()
		m := map[string]string{
			"Et Al.": "Multiple authors",
		}
		_, explanations, err := auditRE("test", m)
		require.NoError(t, err)

		assert.Contains(t, explanations, "et al.")
		assert.Equal(t, "Multiple authors", explanations["et al."])
	})

	t.Run("handles empty input", func(t *testing.T) {
		t.Parallel()
		re, explanations, err := auditRE("test", map[string]string{})
		require.NoError(t, err)
		assert.Nil(t, re)
		assert.Empty(t, explanations)
	})
}
