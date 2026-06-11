package collection

import (
	"testing"

	"github.com/DryHumour/readerware-to-tellico/internal/images"
	"github.com/DryHumour/readerware-to-tellico/internal/normalize"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBooksEntry_Authors(t *testing.T) {
	t.Parallel()

	t.Run("returns all authors from AUTHOR columns", func(t *testing.T) {
		t.Parallel()

		policy := newTestPolicy(t)
		info := policy.Info().(*collectionInfo)
		info.columns = ColumnConfig{
			Names: map[string][]string{
				"Authors": {"AUTHOR", "AUTHOR2", "AUTHOR3"},
			},
		}
		clean := map[string]string{
			"AUTHOR":  "First Author",
			"AUTHOR2": "Second Author",
			"AUTHOR3": "Third Author",
		}

		entry, err := policy.NewEntry(clean, images.Row{})
		require.NoError(t, err)
		booksEntry := entry.(*BooksEntry)

		// Run aggregation to process the name columns
		booksEntry.Aggregate()

		authors := booksEntry.Authors()
		assert.Equal(t, []string{"First Author", "Second Author", "Third Author"}, authors)
	})

	t.Run("includes manually credited authors", func(t *testing.T) {
		t.Parallel()

		policy := newTestPolicy(t)
		info := policy.Info().(*collectionInfo)
		info.columns = ColumnConfig{
			Names: map[string][]string{
				"Authors": {"AUTHOR"},
			},
		}
		clean := map[string]string{
			"AUTHOR": "Column Author",
		}

		entry, err := policy.NewEntry(clean, images.Row{})
		require.NoError(t, err)
		booksEntry := entry.(*BooksEntry)

		// Run aggregation to process the name columns
		booksEntry.Aggregate()

		// Add a manual credit
		booksEntry.AddCredit("Authors", "Manual Author")

		authors := booksEntry.Authors()
		assert.Equal(t, []string{"Column Author", "Manual Author"}, authors)
	})

	t.Run("deduplicates authors within the same role", func(t *testing.T) {
		t.Parallel()

		policy := newTestPolicy(t)
		info := policy.Info().(*collectionInfo)
		info.columns = ColumnConfig{
			Names: map[string][]string{
				"Authors": {"AUTHOR", "AUTHOR2"},
			},
		}
		clean := map[string]string{
			"AUTHOR":  "Duplicate Author",
			"AUTHOR2": "Duplicate Author",
		}

		entry, err := policy.NewEntry(clean, images.Row{})
		require.NoError(t, err)
		booksEntry := entry.(*BooksEntry)

		// Run aggregation to process the name columns
		booksEntry.Aggregate()

		authors := booksEntry.Authors()
		assert.Equal(t, []string{"Duplicate Author"}, authors, "should deduplicate authors")
	})

	t.Run("preserves discovery order across columns", func(t *testing.T) {
		t.Parallel()

		policy := newTestPolicy(t)
		info := policy.Info().(*collectionInfo)
		info.columns = ColumnConfig{
			Names: map[string][]string{
				"Authors": {"AUTHOR", "AUTHOR2", "AUTHOR3"},
			},
		}
		clean := map[string]string{
			"AUTHOR":  "Author One",
			"AUTHOR2": "Author Two",
			"AUTHOR3": "Author Three",
		}

		entry, err := policy.NewEntry(clean, images.Row{})
		require.NoError(t, err)
		booksEntry := entry.(*BooksEntry)

		// Run aggregation to process the name columns
		booksEntry.Aggregate()

		authors := booksEntry.Authors()
		assert.Equal(t, []string{"Author One", "Author Two", "Author Three"}, authors)
	})

	t.Run("handles empty AUTHOR columns", func(t *testing.T) {
		t.Parallel()

		policy := newTestPolicy(t)
		info := policy.Info().(*collectionInfo)
		info.columns = ColumnConfig{
			Names: map[string][]string{
				"Authors": {"AUTHOR", "AUTHOR2", "AUTHOR3"},
			},
		}
		clean := map[string]string{
			"AUTHOR":  "Only Author",
			"AUTHOR2": "",
			"AUTHOR3": "",
		}

		entry, err := policy.NewEntry(clean, images.Row{})
		require.NoError(t, err)
		booksEntry := entry.(*BooksEntry)

		// Run aggregation to process the name columns
		booksEntry.Aggregate()

		authors := booksEntry.Authors()
		assert.Equal(t, []string{"Only Author"}, authors)
	})

	t.Run("handles semicolon-separated authors in single column", func(t *testing.T) {
		t.Parallel()

		policy := newTestPolicy(t)
		info := policy.Info().(*collectionInfo)
		info.columns = ColumnConfig{
			Names: map[string][]string{
				"Authors": {"AUTHOR"},
			},
		}
		clean := map[string]string{
			"AUTHOR": "First Author; Second Author; Third Author",
		}

		entry, err := policy.NewEntry(clean, images.Row{})
		require.NoError(t, err)
		booksEntry := entry.(*BooksEntry)

		// Run aggregation to process the name columns
		booksEntry.Aggregate()

		authors := booksEntry.Authors()
		assert.Equal(t, []string{"First Author", "Second Author", "Third Author"}, authors)
	})

	t.Run("combines column authors and manual credits without duplicates", func(t *testing.T) {
		t.Parallel()

		policy := newTestPolicy(t)
		info := policy.Info().(*collectionInfo)
		info.columns = ColumnConfig{
			Names: map[string][]string{
				"Authors": {"AUTHOR"},
			},
		}
		clean := map[string]string{
			"AUTHOR": "Column Author",
		}

		entry, err := policy.NewEntry(clean, images.Row{})
		require.NoError(t, err)
		booksEntry := entry.(*BooksEntry)

		// Run aggregation to process the name columns
		booksEntry.Aggregate()

		// Add manual credits including a duplicate
		booksEntry.AddCredit("Authors", "Manual Author 1")
		booksEntry.AddCredit("Authors", "Column Author") // duplicate
		booksEntry.AddCredit("Authors", "Manual Author 2")

		authors := booksEntry.Authors()
		assert.Equal(t, []string{"Column Author", "Manual Author 1", "Manual Author 2"}, authors)
	})
}

func TestBooksEntry_Editors(t *testing.T) {
	t.Parallel()

	t.Run("returns all editors from EDITOR columns", func(t *testing.T) {
		t.Parallel()

		policy := newTestPolicy(t)
		info := policy.Info().(*collectionInfo)
		info.columns = ColumnConfig{
			Names: map[string][]string{
				"Editors": {"EDITOR", "EDITOR2"},
			},
		}
		clean := map[string]string{
			"EDITOR":  "First Editor",
			"EDITOR2": "Second Editor",
		}

		entry, err := policy.NewEntry(clean, images.Row{})
		require.NoError(t, err)
		booksEntry := entry.(*BooksEntry)

		// Run aggregation to process the name columns
		booksEntry.Aggregate()

		editors := booksEntry.Editors()
		assert.Equal(t, []string{"First Editor", "Second Editor"}, editors)
	})

	t.Run("includes editors from cross-role annotations in AUTHOR columns", func(t *testing.T) {
		t.Parallel()

		policy := newTestPolicy(t)
		info := policy.Info().(*collectionInfo)
		info.columns = ColumnConfig{
			Names: map[string][]string{
				"Authors": {"AUTHOR", "AUTHOR2"},
			},
		}
		info.names = normalize.NamesConfig{
			Role: map[string]string{
				"(ed.)": "Editors",
			},
		}
		clean := map[string]string{
			"AUTHOR":  "Primary Author",
			"AUTHOR2": "Bailey, Cyril (ed.)",
		}

		entry, err := policy.NewEntry(clean, images.Row{})
		require.NoError(t, err)
		booksEntry := entry.(*BooksEntry)

		// Run aggregation to process the name columns
		booksEntry.Aggregate()

		editors := booksEntry.Editors()
		assert.Equal(t, []string{"Cyril Bailey"}, editors)
	})

	t.Run("combines EDITOR columns and cross-role annotations", func(t *testing.T) {
		t.Parallel()

		policy := newTestPolicy(t)
		info := policy.Info().(*collectionInfo)
		info.columns = ColumnConfig{
			Names: map[string][]string{
				"Authors": {"AUTHOR", "AUTHOR2"},
				"Editors": {"EDITOR"},
			},
		}
		info.names = normalize.NamesConfig{
			Role: map[string]string{
				"(ed.)": "Editors",
			},
		}
		clean := map[string]string{
			"AUTHOR":  "Primary Author",
			"AUTHOR2": "Bailey, Cyril (ed.)",
			"EDITOR":  "Smith, John (ed.)",
		}

		entry, err := policy.NewEntry(clean, images.Row{})
		require.NoError(t, err)
		booksEntry := entry.(*BooksEntry)

		// Run aggregation to process the name columns
		booksEntry.Aggregate()

		editors := booksEntry.Editors()
		assert.Equal(t, []string{"John Smith", "Cyril Bailey"}, editors)
	})

	t.Run("includes manually credited editors", func(t *testing.T) {
		t.Parallel()

		policy := newTestPolicy(t)
		info := policy.Info().(*collectionInfo)
		info.columns = ColumnConfig{
			Names: map[string][]string{
				"Editors": {"EDITOR"},
			},
		}
		clean := map[string]string{
			"EDITOR": "Column Editor",
		}

		entry, err := policy.NewEntry(clean, images.Row{})
		require.NoError(t, err)
		booksEntry := entry.(*BooksEntry)

		// Run aggregation to process the name columns
		booksEntry.Aggregate()

		// Add a manual credit
		booksEntry.AddCredit("Editors", "Manual Editor")

		editors := booksEntry.Editors()
		assert.Equal(t, []string{"Column Editor", "Manual Editor"}, editors)
	})

	t.Run("deduplicates editors within the same role", func(t *testing.T) {
		t.Parallel()

		policy := newTestPolicy(t)
		info := policy.Info().(*collectionInfo)
		info.columns = ColumnConfig{
			Names: map[string][]string{
				"Editors": {"EDITOR", "EDITOR2"},
			},
		}
		clean := map[string]string{
			"EDITOR":  "Duplicate Editor",
			"EDITOR2": "Duplicate Editor",
		}

		entry, err := policy.NewEntry(clean, images.Row{})
		require.NoError(t, err)
		booksEntry := entry.(*BooksEntry)

		// Run aggregation to process the name columns
		booksEntry.Aggregate()

		editors := booksEntry.Editors()
		assert.Equal(t, []string{"Duplicate Editor"}, editors, "should deduplicate editors")
	})

	t.Run("preserves discovery order across columns", func(t *testing.T) {
		t.Parallel()

		policy := newTestPolicy(t)
		info := policy.Info().(*collectionInfo)
		info.columns = ColumnConfig{
			Names: map[string][]string{
				"Editors": {"EDITOR", "EDITOR2", "EDITOR3"},
			},
		}
		clean := map[string]string{
			"EDITOR":  "Editor One",
			"EDITOR2": "Editor Two",
			"EDITOR3": "Editor Three",
		}

		entry, err := policy.NewEntry(clean, images.Row{})
		require.NoError(t, err)
		booksEntry := entry.(*BooksEntry)

		// Run aggregation to process the name columns
		booksEntry.Aggregate()

		editors := booksEntry.Editors()
		assert.Equal(t, []string{"Editor One", "Editor Two", "Editor Three"}, editors)
	})

	t.Run("handles empty EDITOR columns", func(t *testing.T) {
		t.Parallel()

		policy := newTestPolicy(t)
		info := policy.Info().(*collectionInfo)
		info.columns = ColumnConfig{
			Names: map[string][]string{
				"Editors": {"EDITOR", "EDITOR2"},
			},
		}
		clean := map[string]string{
			"EDITOR":  "Only Editor",
			"EDITOR2": "",
		}

		entry, err := policy.NewEntry(clean, images.Row{})
		require.NoError(t, err)
		booksEntry := entry.(*BooksEntry)

		// Run aggregation to process the name columns
		booksEntry.Aggregate()

		editors := booksEntry.Editors()
		assert.Equal(t, []string{"Only Editor"}, editors)
	})

	t.Run("handles semicolon-separated editors in single column", func(t *testing.T) {
		t.Parallel()

		policy := newTestPolicy(t)
		info := policy.Info().(*collectionInfo)
		info.columns = ColumnConfig{
			Names: map[string][]string{
				"Editors": {"EDITOR"},
			},
		}
		clean := map[string]string{
			"EDITOR": "First Editor; Second Editor; Third Editor",
		}

		entry, err := policy.NewEntry(clean, images.Row{})
		require.NoError(t, err)
		booksEntry := entry.(*BooksEntry)

		// Run aggregation to process the name columns
		booksEntry.Aggregate()

		editors := booksEntry.Editors()
		assert.Equal(t, []string{"First Editor", "Second Editor", "Third Editor"}, editors)
	})

	t.Run("combines column editors and manual credits without duplicates", func(t *testing.T) {
		t.Parallel()

		policy := newTestPolicy(t)
		info := policy.Info().(*collectionInfo)
		info.columns = ColumnConfig{
			Names: map[string][]string{
				"Editors": {"EDITOR"},
			},
		}
		clean := map[string]string{
			"EDITOR": "Column Editor",
		}

		entry, err := policy.NewEntry(clean, images.Row{})
		require.NoError(t, err)
		booksEntry := entry.(*BooksEntry)

		// Run aggregation to process the name columns
		booksEntry.Aggregate()

		// Add manual credits including a duplicate
		booksEntry.AddCredit("Editors", "Manual Editor 1")
		booksEntry.AddCredit("Editors", "Column Editor") // duplicate
		booksEntry.AddCredit("Editors", "Manual Editor 2")

		editors := booksEntry.Editors()
		assert.Equal(t, []string{"Column Editor", "Manual Editor 1", "Manual Editor 2"}, editors)
	})
}
