package normalize

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkers(t *testing.T) {
	t.Parallel()

	config := MarkersConfig{
		Keep:    []string{"KeepMe"},
		Discard: []string{"DiscardMe"},
		Scrub:   []string{"noise"},
		Marker:  map[string]string{"(signed)": "<signed>"},
		Audit:   map[string]string{"audit-me": "[audit]"},
	}

	m, err := NewMarkers(config)
	require.NoError(t, err)

	t.Run("Scrub removes noise and canonicalizes suffixes", func(t *testing.T) {
		t.Parallel()
		// Precondition: input is squeezed
		r := m.Begin("John Doe jr.")
		result := m.Scrub(r)
		assert.Equal(t, "John Doe jr.", result.Value)

		r2 := m.Begin("Name noise") // Use word-based noise for now
		result2 := m.Scrub(r2)
		assert.Equal(t, "Name", result2.Value)
	})

	t.Run("Filter marks keep/discard", func(t *testing.T) {
		t.Parallel()
		// Keep
		rKeep := m.Filter(m.Begin("KeepMe"))
		assert.True(t, rKeep.Literal)

		// Discard
		rDiscard := m.Filter(m.Begin("DiscardMe"))
		assert.True(t, rDiscard.Discarded)
	})

	t.Run("Extract pulls markers", func(t *testing.T) {
		t.Parallel()
		// Precondition: input is squeezed and scrubbed
		r := m.Begin("Book Title (signed)")
		result := m.Extract(r)
		assert.Equal(t, "Book Title", result.Value)
		assert.Contains(t, result.Markers, "<signed>")
	})

	t.Run("Audit flags issues", func(t *testing.T) {
		t.Parallel()
		// Precondition: input is squeezed and extracted
		r := m.Begin("Value audit-me")
		result := m.Audit(r)
		assert.True(t, result.RequiresAudit)
	})

	t.Run("Process executes full pipeline", func(t *testing.T) {
		t.Parallel()
		// No preconditions for Process
		result := m.Process(" noise (signed) audit-me ")
		assert.Equal(t, "audit-me", result.Value)
		assert.Contains(t, result.Markers, "<signed>")
	})
}
