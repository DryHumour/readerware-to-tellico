package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTailBuffer(t *testing.T) {
	t.Parallel()

	t.Run("retains everything under max", func(t *testing.T) {
		t.Parallel()

		buf := newTailBuffer(16)
		n, err := buf.Write([]byte("hello"))
		require.NoError(t, err)
		assert.Equal(t, 5, n)
		assert.Equal(t, "hello", buf.String())
	})

	t.Run("keeps only the tail over max", func(t *testing.T) {
		t.Parallel()

		buf := newTailBuffer(4)
		n, err := buf.Write([]byte("abcdefgh"))
		require.NoError(t, err)
		assert.Equal(t, 8, n)
		assert.Equal(t, "efgh", buf.String())
	})

	t.Run("accumulates across writes", func(t *testing.T) {
		t.Parallel()

		buf := newTailBuffer(6)
		_, err := buf.Write([]byte("abc"))
		require.NoError(t, err)
		_, err = buf.Write([]byte("defgh"))
		require.NoError(t, err)
		assert.Equal(t, "cdefgh", buf.String())
	})

	t.Run("write larger than max keeps tail of that write", func(t *testing.T) {
		t.Parallel()

		buf := newTailBuffer(3)
		_, err := buf.Write([]byte("0123456789"))
		require.NoError(t, err)
		assert.Equal(t, "789", buf.String())
	})
}
