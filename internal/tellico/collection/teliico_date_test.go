package collection

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTellicoDate(t *testing.T) {
	t.Parallel()

	t.Run("empty string returns nil", func(t *testing.T) {
		t.Parallel()
		require.Nil(t, NewTellicoDate(""))
	})

	t.Run("valid ISO date format", func(t *testing.T) {
		t.Parallel()
		result := NewTellicoDate("2024-03-15")
		require.NotNil(t, result)
		require.Equal(t, 2024, result.YYYY)
		require.Equal(t, 3, result.MM)
		require.Equal(t, 15, result.DD)
		require.Empty(t, result.Literal)
	})

	t.Run("invalid date format uses literal", func(t *testing.T) {
		t.Parallel()
		result := NewTellicoDate("March 15, 2024")
		require.NotNil(t, result)
		require.Equal(t, "March 15, 2024", result.Literal)
	})
}
