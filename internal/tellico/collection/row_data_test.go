package collection

import (
	"testing"

	"github.com/DryHumour/readerware-to-tellico/internal/images"
	"github.com/stretchr/testify/assert"
)

func TestRowData_Categories(t *testing.T) {

	info := &collectionInfo{
		columns: ColumnConfig{
			Categories: map[string]bool{
				"CATEGORY1": true,
				"CATEGORY2": true,
				"CATEGORY3": true,
			},
		},
	}

	t.Run("returns values from configured category columns", func(t *testing.T) {

		clean := map[string]string{
			"CATEGORY1": "Fiction:Science Fiction",
			"CATEGORY2": "Fiction:Authors, A-Z:Adams",
		}
		row := newRowData(info, clean, images.Row{})

		result := row.Categories()
		assert.ElementsMatch(t, []string{"Fiction:Science Fiction", "Fiction:Authors, A-Z:Adams"}, result)
	})

	t.Run("filters out empty values", func(t *testing.T) {

		clean := map[string]string{
			"CATEGORY1": "Fiction",
			"CATEGORY2": "",
			"CATEGORY3": "Non-Fiction",
		}
		row := newRowData(info, clean, images.Row{})

		result := row.Categories()
		assert.ElementsMatch(t, []string{"Fiction", "Non-Fiction"}, result)
	})

	t.Run("returns empty when no categories", func(t *testing.T) {

		clean := map[string]string{
			"CATEGORY1": "",
			"CATEGORY2": "",
		}
		row := newRowData(info, clean, images.Row{})

		result := row.Categories()
		assert.Empty(t, result)
	})
}

func TestRowData_Genres(t *testing.T) {

	info := &collectionInfo{
		columns: ColumnConfig{
			Categories: map[string]bool{
				"CATEGORY1": true,
			},
		},
		blocklist: map[string]bool{
			"Blocklisted": true,
		},
	}

	t.Run("parses colon-separated path", func(t *testing.T) {

		clean := map[string]string{
			"CATEGORY1": "Fiction:Science Fiction",
		}
		row := newRowData(info, clean, images.Row{})

		result := row.Genres()
		assert.ElementsMatch(t, []string{"Fiction", "Science Fiction"}, result)
	})

	t.Run("parses pipe-separated path", func(t *testing.T) {

		clean := map[string]string{
			"CATEGORY1": "Fiction|Science Fiction",
		}
		row := newRowData(info, clean, images.Row{})

		result := row.Genres()
		assert.ElementsMatch(t, []string{"Fiction", "Science Fiction"}, result)
	})

	t.Run("parses greater-than-separated path", func(t *testing.T) {

		clean := map[string]string{
			"CATEGORY1": "Fiction>Science Fiction",
		}
		row := newRowData(info, clean, images.Row{})

		result := row.Genres()
		assert.ElementsMatch(t, []string{"Fiction", "Science Fiction"}, result)
	})

	t.Run("removes single-character values", func(t *testing.T) {

		clean := map[string]string{
			"CATEGORY1": "Fiction:A:B:C",
		}
		row := newRowData(info, clean, images.Row{})

		result := row.Genres()
		assert.ElementsMatch(t, []string{"Fiction"}, result)
	})

	t.Run("removes blocklisted values", func(t *testing.T) {

		clean := map[string]string{
			"CATEGORY1": "Fiction:Blocklisted:Science Fiction",
		}
		row := newRowData(info, clean, images.Row{})

		result := row.Genres()
		assert.ElementsMatch(t, []string{"Fiction", "Science Fiction"}, result)
	})

	t.Run("stops at path navigation nodes", func(t *testing.T) {

		clean := map[string]string{
			"CATEGORY1": "Fiction:Authors, A-Z:Adams",
		}
		row := newRowData(info, clean, images.Row{})

		result := row.Genres()
		// "Authors, A-Z" matches the cut regex, so "Adams" should be excluded
		assert.ElementsMatch(t, []string{"Fiction"}, result)
	})

	t.Run("deduplicates across paths", func(t *testing.T) {

		info := &collectionInfo{
			columns: ColumnConfig{
				Categories: map[string]bool{
					"CATEGORY1": true,
					"CATEGORY2": true,
				},
			},
		}
		clean := map[string]string{
			"CATEGORY1": "Fiction:Science Fiction",
			"CATEGORY2": "Fiction:Science Fiction",
		}
		row := newRowData(info, clean, images.Row{})

		result := row.Genres()
		assert.ElementsMatch(t, []string{"Fiction", "Science Fiction"}, result)
	})

	t.Run("handles empty path", func(t *testing.T) {

		clean := map[string]string{
			"CATEGORY1": "",
		}
		row := newRowData(info, clean, images.Row{})

		result := row.Genres()
		assert.Empty(t, result)
	})

	t.Run("handles mixed separators", func(t *testing.T) {

		clean := map[string]string{
			"CATEGORY1": "Fiction:Science Fiction|Fantasy",
		}
		row := newRowData(info, clean, images.Row{})

		result := row.Genres()
		assert.ElementsMatch(t, []string{"Fiction", "Science Fiction", "Fantasy"}, result)
	})
}

func TestColumnConfig_ColumnRole(t *testing.T) {

	config := ColumnConfig{
		Names: map[string][]string{
			"Authors":    {"AUTHOR"},
			"Editors":    {"EDITOR"},
			"Artists":    {"ARTIST"},
			"Publishers": {"PUBLISHER"},
		},
	}

	t.Run("returns configured role for column", func(t *testing.T) {
		assert.Equal(t, "Authors", config.ColumnRole("AUTHOR"))
		assert.Equal(t, "Editors", config.ColumnRole("EDITOR"))
	})

	t.Run("returns column name when not configured", func(t *testing.T) {
		assert.Equal(t, "TITLE", config.ColumnRole("TITLE"))
		assert.Equal(t, "ISBN", config.ColumnRole("ISBN"))
	})
}
