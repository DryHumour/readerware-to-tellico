package normalize

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommonEngines(t *testing.T) {
	t.Parallel()

	t.Run("scrub removes matches and re-squeezes", func(t *testing.T) {
		t.Parallel()
		re := regexp.MustCompile(`(?i:edition|ed\.)`)

		// Note: Input is already squeezed per precondition
		r := Result{Value: "First Edition"}
		result := scrub(r, re)
		assert.Equal(t, "First", result.Value)

		r2 := Result{Value: "Something Edition Else"}
		result2 := scrub(r2, re)
		assert.Equal(t, "Something Else", result2.Value)
	})

	t.Run("filter marks Literal or Discarded", func(t *testing.T) {
		t.Parallel()
		keep := regexp.MustCompile(`^Prince$`)
		discard := regexp.MustCompile(`(?i:^n/a$)`)

		// Keep case
		r := Result{Value: "Prince"}
		resKeep := filter(r, keep, discard)
		assert.True(t, resKeep.Literal)
		assert.False(t, resKeep.Discarded)

		// Discard case
		r2 := Result{Value: "N/A"}
		resDiscard := filter(r2, keep, discard)
		assert.True(t, resDiscard.Discarded)
		assert.Empty(t, resDiscard.Value)

		// Passthrough case
		r3 := Result{Value: "John Doe"}
		resPass := filter(r3, keep, discard)
		assert.False(t, resPass.Literal)
		assert.False(t, resPass.Discarded)
		assert.Equal(t, "John Doe", resPass.Value)
	})

	t.Run("extract pulls matches using co-produced parameters", func(t *testing.T) {
		t.Parallel()
		m := map[string]string{
			"(ed.)": "Editors",
		}
		re, canonical, err := canonicalRE("test", m)
		require.NoError(t, err)

		// Input is squeezed per precondition
		remaining, matches := extract("Adams, Douglas (ed.)", re, canonical)
		assert.Equal(t, "Adams, Douglas", remaining)
		assert.Equal(t, []string{"Editors"}, matches)
	})

	t.Run("audit identifies patterns and flags Result", func(t *testing.T) {
		t.Parallel()
		re := regexp.MustCompile(`(?i:et al\.)`)

		r := Result{Value: "Smith et al."}
		result := audit(r, re, nil)
		assert.True(t, result.RequiresAudit)
		assert.Contains(t, result.AuditReasons, `[audit]: "et al."`)
	})

	t.Run("auditReason truncates and quotes", func(t *testing.T) {
		t.Parallel()
		reason := AuditReason("[test]", "Long value that exceeds twenty-five runes")
		assert.Contains(t, reason, `[test]: "Long value that exceeds t…"`)

		reasonShort := AuditReason("[test]", "short")
		assert.Equal(t, `[test]: "short"`, reasonShort)
	})
}
