package collection

import (
	"testing"

	"github.com/DryHumour/readerware-to-tellico/internal/images"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestPolicy creates a BooksPolicy for testing.
func newTestPolicy(t *testing.T) *BooksPolicy {
	t.Helper()
	return NewBooksPolicy()
}

func TestNewBooksPolicy(t *testing.T) {

	t.Run("creates policy with default genre blocklist", func(t *testing.T) {
		policy := newTestPolicy(t)
		require.NotNil(t, policy)
		require.Equal(t, KindBooks, policy.Info().Kind())
	})

}

func TestBooksPolicy_ConfigureHeaders(t *testing.T) {

	t.Run("valid headers without images", func(t *testing.T) {
		policy := newTestPolicy(t)
		assert.NoError(t, policy.ConfigureHeaders([]string{"TITLE", "ROW#", "AUTHOR", "PUBLISHER"}, false))
	})

	t.Run("valid headers with ROW#", func(t *testing.T) {
		policy := newTestPolicy(t)
		assert.NoError(t, policy.ConfigureHeaders([]string{"TITLE", "ROW#", "AUTHOR"}, false))
	})

	t.Run("valid headers with ROWKEY and images", func(t *testing.T) {
		policy := newTestPolicy(t)
		assert.NoError(t, policy.ConfigureHeaders([]string{"TITLE", "ROWKEY", "AUTHOR"}, true))
	})

	t.Run("missing TITLE header", func(t *testing.T) {
		policy := newTestPolicy(t)
		err := policy.ConfigureHeaders([]string{"AUTHOR", "PUBLISHER"}, false)
		assert.ErrorContains(t, err, "missing required header: TITLE")
	})

	t.Run("missing ROWKEY when images enabled", func(t *testing.T) {
		policy := newTestPolicy(t)
		err := policy.ConfigureHeaders([]string{"TITLE", "AUTHOR"}, true)
		assert.ErrorContains(t, err, "missing required header: ROWKEY")
	})

	t.Run("missing both ROWKEY and ROW# when images disabled", func(t *testing.T) {
		policy := newTestPolicy(t)
		err := policy.ConfigureHeaders([]string{"TITLE", "AUTHOR"}, false)
		assert.ErrorContains(t, err, "missing required header: either ROWKEY or ROW#")
	})

	t.Run("duplicate header", func(t *testing.T) {
		policy := newTestPolicy(t)
		err := policy.ConfigureHeaders([]string{"TITLE", "AUTHOR", "TITLE"}, false)
		assert.ErrorContains(t, err, `duplicate header: "TITLE" appears more than once`)
	})
}

func TestBooksPolicy_TemplateNames(t *testing.T) {

	names := NewBooksPolicy().Info().TemplateNames()
	assert.Equal(t, "books.config", names.Config)
	assert.Equal(t, "books.header", names.Header)
	assert.Equal(t, "books.entry", names.Entry)
	assert.Equal(t, "books.footer", names.Footer)
}

func TestBooksPolicy_NewEntry(t *testing.T) {

	t.Run("builds entry data with all fields", func(t *testing.T) {

		policy := newTestPolicy(t)
		// Configure the policy with test data
		info := policy.Info().(*collectionInfo)
		info.columns = ColumnConfig{
			Names: map[string][]string{
				"TITLE":     {"TITLE"},
				"AUTHOR":    {"AUTHOR"},
				"PUBLISHER": {"PUBLISHER"},
				"CATEGORY":  {"CATEGORY"},
			},
		}
		clean := map[string]string{
			"TITLE":     "Test Book",
			"AUTHOR":    "John Doe",
			"PUBLISHER": "Test Publisher",
			"CATEGORY":  "Fiction",
		}

		entry, err := policy.NewEntry(clean, images.Row{})
		require.NoError(t, err)
		assert.IsType(t, &BooksEntry{}, entry)
		assert.Equal(t, "Test Book", entry.V("TITLE"))
		assert.Equal(t, "John Doe", entry.V("AUTHOR"))
	})

	t.Run("builds entry data with multiple authors", func(t *testing.T) {

		policy := NewBooksPolicy()
		// Configure the policy with test data
		info := policy.Info().(*collectionInfo)
		info.columns = ColumnConfig{
			Names: map[string][]string{
				"Authors": {"AUTHOR", "AUTHOR2"},
			},
		}
		clean := map[string]string{
			"AUTHOR":  "Author One",
			"AUTHOR2": "Author Two",
		}

		entry, err := policy.NewEntry(clean, images.Row{})
		require.NoError(t, err)
		assert.IsType(t, &BooksEntry{}, entry)
		assert.Equal(t, "Author One", entry.V("AUTHOR"))
		assert.Equal(t, "Author Two", entry.V("AUTHOR2"))
	})
}

func TestNew(t *testing.T) {

	t.Run("returns books policy for books kind", func(t *testing.T) {

		policy, err := New(KindBooks)
		require.NoError(t, err)
		require.NotNil(t, policy)
		require.Equal(t, KindBooks, policy.Info().Kind())
	})

	t.Run("returns music policy for music kind", func(t *testing.T) {

		policy, err := New(KindMusic)
		require.NoError(t, err)
		require.NotNil(t, policy)
		require.Equal(t, KindMusic, policy.Info().Kind())
	})

	t.Run("returns video policy for video kind", func(t *testing.T) {

		policy, err := New(KindVideo)
		require.NoError(t, err)
		require.NotNil(t, policy)
		require.Equal(t, KindVideo, policy.Info().Kind())
	})

	t.Run("error for unknown kind", func(t *testing.T) {

		_, err := New(Kind("unknown"))
		assert.ErrorIs(t, err, ErrUnknownKind("unknown"))
	})
}
