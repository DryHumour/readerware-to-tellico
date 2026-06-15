package normalize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResult(t *testing.T) {
	t.Parallel()

	t.Run("Keep marks as literal", func(t *testing.T) {
		t.Parallel()
		r := Result{Value: "Original"}
		r2 := r.Keep()

		assert.True(t, r2.Literal)
		assert.Equal(t, "Original", r2.Value)
	})

	t.Run("Discard marks as discarded and clears value", func(t *testing.T) {
		t.Parallel()
		r := Result{Value: "Original"}
		r2 := r.Discard()

		assert.True(t, r2.Discarded)
		assert.Empty(t, r2.Value)
	})

	t.Run("Update modifies value but preserves other fields", func(t *testing.T) {
		t.Parallel()
		r := Result{
			Value:         "Original",
			Roles:         []string{"Role1"},
			Markers:       []string{"Marker1"},
			RequiresAudit: true,
			AuditReasons:  []string{"Reason1"},
		}
		r2 := r.Update("New Value")

		assert.Equal(t, "New Value", r2.Value)
		assert.Equal(t, r.Roles, r2.Roles)
		assert.Equal(t, r.Markers, r2.Markers)
		assert.Equal(t, r.RequiresAudit, r2.RequiresAudit)
		assert.Equal(t, r.AuditReasons, r2.AuditReasons)
	})

	t.Run("AddAudit appends reasons and flags audit", func(t *testing.T) {
		t.Parallel()
		r := Result{
			Value:        "Value",
			AuditReasons: []string{"Reason1"},
		}
		r2 := r.AddAudit("Reason2")

		assert.True(t, r2.RequiresAudit)
		assert.Equal(t, []string{"Reason1", "Reason2"}, r2.AuditReasons)

		// Verify that the original Result remains unchanged (immutability check)
		assert.False(t, r.RequiresAudit)
		assert.Equal(t, []string{"Reason1"}, r.AuditReasons)
	})

	t.Run("isImmutable logic", func(t *testing.T) {
		t.Parallel()
		assert.False(t, Result{Value: "v"}.isImmutable())
		assert.True(t, Result{Value: "v", Literal: true}.isImmutable())
		assert.True(t, Result{Value: "", Discarded: true}.isImmutable())
		assert.True(t, Result{Value: "v", Literal: true, Discarded: true}.isImmutable())
	})
}
