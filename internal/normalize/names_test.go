package normalize

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNames(t *testing.T) {
	t.Parallel()

	config := NamesConfig{
		Keep:       []string{"Prince"},
		Discard:    []string{"n/a"},
		Scrub:      []string{"DELETE_ME"},
		Role:       map[string]string{"(ed.)": "Editors"},
		Suffixes:   []string{"Jr.", "Sr."},
		Honorifics: []string{"Sir", "Dr."},
		Corporate:  []string{"Orchestra"},
	}

	n, err := NewNames(config)
	require.NoError(t, err)

	t.Run("Scrub removes noise and canonicalizes suffixes", func(t *testing.T) {
		t.Parallel()
		// Precondition: input is squeezed
		r := n.Begin("John Doe jr.")
		result := n.Scrub(r)
		assert.Equal(t, "John Doe Jr.", result.Value)

		r2 := n.Begin("Name DELETE_ME") // Use explicit noise
		result2 := n.Scrub(r2)
		assert.Equal(t, "Name", result2.Value)
	})

	t.Run("Extract pulls roles", func(t *testing.T) {
		t.Parallel()
		// Precondition: input is squeezed and scrubbed
		r := n.Begin("Adams, Douglas (ed.)")
		result := n.Extract(r)
		assert.Equal(t, "Adams, Douglas", result.Value)
		assert.Contains(t, result.Roles, "Editors")
	})

	t.Run("NaturalOrder flips names", func(t *testing.T) {
		t.Parallel()
		// Precondition: input is squeezed and extracted
		r := n.Begin("Adams, Douglas")
		result := n.NaturalOrder(r)
		assert.Equal(t, "Douglas Adams", result.Value)
	})

	t.Run("NaturalOrder handles 3-part names", func(t *testing.T) {
		t.Parallel()
		// Suffix case
		r := n.Begin("Doe, John, Jr.")
		resSfx := n.NaturalOrder(r)
		assert.Equal(t, "John Doe Jr.", resSfx.Value)

		// Honorific case
		r2 := n.Begin("Redgrave, Michael, Sir")
		resHon := n.NaturalOrder(r2)
		assert.Equal(t, "Sir Michael Redgrave", resHon.Value)
	})

	t.Run("NaturalOrder corporate short-circuit", func(t *testing.T) {
		t.Parallel()
		r := n.Begin("London Orchestra, The")
		result := n.NaturalOrder(r)
		assert.Equal(t, "London Orchestra, The", result.Value)
		assert.True(t, result.RequiresAudit)
	})

	t.Run("Process executes full pipeline", func(t *testing.T) {
		t.Parallel()
		// No preconditions for Process
		result := n.Process(" Adams, Douglas (ed.) ")
		assert.Equal(t, "Douglas Adams", result.Value)
		assert.Contains(t, result.Roles, "Editors")
	})

	t.Run("Filter marks keep/discard", func(t *testing.T) {
		t.Parallel()
		// Keep
		rKeep := n.Filter(n.Begin("Prince"))
		assert.True(t, rKeep.Literal)

		// Discard
		rDiscard := n.Filter(n.Begin("n/a"))
		assert.True(t, rDiscard.Discarded)

		// Passthrough
		rPass := n.Filter(n.Begin("John Doe"))
		assert.False(t, rPass.Literal)
		assert.False(t, rPass.Discarded)
		assert.Equal(t, "John Doe", rPass.Value)
	})

	t.Run("Audit flags names-specific issues", func(t *testing.T) {
		t.Parallel()

		t.Run("ambiguous honorifics in middle of name", func(t *testing.T) {
			t.Parallel()
			config := NamesConfig{
				Honorifics: []string{"Dr."},
			}
			n2, err := NewNames(config)
			require.NoError(t, err)

			r := n2.Begin("John Dr. Smith")
			result := n2.Audit(r)
			assert.True(t, result.RequiresAudit)
			assert.Contains(t, result.AuditReasons[0], "[ambiguous]")
		})

		t.Run("collaboration signifier", func(t *testing.T) {
			t.Parallel()
			config := NamesConfig{
				Collaboration: []string{"featuring"},
			}
			n2, err := NewNames(config)
			require.NoError(t, err)

			r := n2.Begin("Artist featuring Guest")
			result := n2.Audit(r)
			assert.True(t, result.RequiresAudit)
			assert.Contains(t, result.AuditReasons[0], "[collaboration]")
		})

		t.Run("parentheses in name", func(t *testing.T) {
			t.Parallel()
			r := n.Begin("Name (with parens)")
			result := n.Audit(r)
			assert.True(t, result.RequiresAudit)
			assert.Contains(t, result.AuditReasons[0], "[parentheses]")
		})

		t.Run("trailing punctuation not in config", func(t *testing.T) {
			t.Parallel()
			r := n.Begin("Name!")
			result := n.Audit(r)
			assert.True(t, result.RequiresAudit)
			assert.Contains(t, result.AuditReasons[0], "[punctuation]")
		})
	})

	t.Run("canonicalSuffix returns canonical form", func(t *testing.T) {
		t.Parallel()
		// Test that suffix is canonicalized to configured form
		assert.Equal(t, "Jr.", n.canonicalSuffix("jr."))
		assert.Equal(t, "Jr.", n.canonicalSuffix("JR."))
		assert.Equal(t, "Jr.", n.canonicalSuffix("Jr."))
		assert.Equal(t, "Sr.", n.canonicalSuffix("sr"))
		assert.Equal(t, "unknown", n.canonicalSuffix("unknown")) // no match, returns as-is
	})
}
