package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDirectories_Get(t *testing.T) {
	t.Parallel()

	t.Run("returns First directory for Slot1", func(t *testing.T) {
		t.Parallel()

		dirs := Directories{
			First:  "/path/first",
			Second: "/path/second",
		}

		result := dirs.Get(Slot1)
		require.Equal(t, "/path/first", result)
	})

	t.Run("returns Second directory for Slot2", func(t *testing.T) {
		t.Parallel()

		dirs := Directories{
			First:  "/path/first",
			Second: "/path/second",
		}

		result := dirs.Get(Slot2)
		require.Equal(t, "/path/second", result)
	})

	t.Run("returns Third directory for Slot3", func(t *testing.T) {
		t.Parallel()

		dirs := Directories{
			Third: "/path/third",
		}

		result := dirs.Get(Slot3)
		require.Equal(t, "/path/third", result)
	})

	t.Run("returns Fourth directory for Slot4", func(t *testing.T) {
		t.Parallel()

		dirs := Directories{
			Fourth: "/path/fourth",
		}

		result := dirs.Get(Slot4)
		require.Equal(t, "/path/fourth", result)
	})

	t.Run("returns FirstLarge directory for SlotLarge1", func(t *testing.T) {
		t.Parallel()

		dirs := Directories{
			FirstLarge: "/path/large1",
		}

		result := dirs.Get(SlotLarge1)
		require.Equal(t, "/path/large1", result)
	})

	t.Run("returns SecondLarge directory for SlotLarge2", func(t *testing.T) {
		t.Parallel()

		dirs := Directories{
			SecondLarge: "/path/large2",
		}

		result := dirs.Get(SlotLarge2)
		require.Equal(t, "/path/large2", result)
	})

	t.Run("returns empty string for invalid slot", func(t *testing.T) {
		t.Parallel()

		dirs := Directories{
			First: "/path/first",
		}

		result := dirs.Get(Slot(99))
		require.Equal(t, "", result)
	})

	t.Run("returns empty string for empty directory", func(t *testing.T) {
		t.Parallel()

		dirs := Directories{}

		result := dirs.Get(Slot1)
		require.Equal(t, "", result)
	})
}

func TestSlot_String(t *testing.T) {
	t.Parallel()

	cases := []struct {
		slot     Slot
		expected string
	}{
		{Slot1, "first"},
		{Slot2, "second"},
		{Slot3, "third"},
		{Slot4, "fourth"},
		{SlotLarge1, "first large"},
		{SlotLarge2, "second large"},
		{SlotLarge3, "third large"},
		{SlotLarge4, "fourth large"},
		{Slot(99), ""},
	}

	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, tc.slot.String())
		})
	}
}

func TestSlot_Position(t *testing.T) {
	t.Parallel()

	t.Run("returns 1-based position for regular slots", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, 1, Slot1.Position())
		require.Equal(t, 2, Slot2.Position())
		require.Equal(t, 3, Slot3.Position())
		require.Equal(t, 4, Slot4.Position())
	})

	t.Run("returns 1-based position for large slots", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, 1, SlotLarge1.Position())
		require.Equal(t, 2, SlotLarge2.Position())
		require.Equal(t, 3, SlotLarge3.Position())
		require.Equal(t, 4, SlotLarge4.Position())
	})
}

func TestSlot_IsLarge(t *testing.T) {
	t.Parallel()

	t.Run("returns false for regular slots", func(t *testing.T) {
		t.Parallel()

		require.False(t, Slot1.IsLarge())
		require.False(t, Slot2.IsLarge())
		require.False(t, Slot3.IsLarge())
		require.False(t, Slot4.IsLarge())
	})

	t.Run("returns true for large slots", func(t *testing.T) {
		t.Parallel()

		require.True(t, SlotLarge1.IsLarge())
		require.True(t, SlotLarge2.IsLarge())
		require.True(t, SlotLarge3.IsLarge())
		require.True(t, SlotLarge4.IsLarge())
	})
}

func TestSlot_Invert(t *testing.T) {
	t.Parallel()

	t.Run("converts regular slot to large slot", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, SlotLarge1, Slot1.Invert())
		require.Equal(t, SlotLarge2, Slot2.Invert())
		require.Equal(t, SlotLarge3, Slot3.Invert())
		require.Equal(t, SlotLarge4, Slot4.Invert())
	})

	t.Run("converts large slot to regular slot", func(t *testing.T) {
		t.Parallel()

		require.Equal(t, Slot1, SlotLarge1.Invert())
		require.Equal(t, Slot2, SlotLarge2.Invert())
		require.Equal(t, Slot3, SlotLarge3.Invert())
		require.Equal(t, Slot4, SlotLarge4.Invert())
	})
}
