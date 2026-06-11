package collection

import (
	"testing"

	"github.com/DryHumour/readerware-to-tellico/internal/images"
)

// FuzzAggregateNames verifies that AggregateNames never panics regardless of
// the cell value content. The column config and normalizer are fixed; only the
// value stored in the AUTHOR column is fuzzed.
func FuzzAggregateNames(f *testing.F) {
	seeds := []string{
		"Adams, Douglas",
		"Adams, Douglas (ed.)",
		"Smith, John; Doe, Jane (trans.)",
		"Adams, Douglas ; Jones, Bob (ed.)",
		"",
		"   ",
		"Non-ASCII: 世界",
		"(ed.)",
		";;; (illus.) ;;;",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	info := &collectionInfo{columns: testNameColumns}
	names := newTestNames(f)

	f.Fuzz(func(t *testing.T, cellValue string) {
		entry, err := newBasicEntry(info, map[string]string{"AUTHOR": cellValue}, images.Row{})
		if err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}
		AggregateNames(entry, names)
	})
}

// FuzzGenres verifies that BooksEntry.Genres never panics regardless of the
// category path value. The column config and blocklist are fixed.
func FuzzGenres(f *testing.F) {
	seeds := []string{
		"Fiction:Science Fiction",
		"Fiction:Authors, A-Z:Adams",
		"F:Fiction",
		"",
		"   ",
		"Non-ASCII: 世界",
		"A:B:C:D:E:F",
		":::",
		"Fiction|Fantasy>Epic",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	info := &collectionInfo{columns: testNameColumns}

	f.Fuzz(func(t *testing.T, path string) {
		entry, err := newBooksEntry(info, map[string]string{"CATEGORY1": path}, images.Row{})
		if err != nil {
			t.Fatalf("Failed to create entry: %v", err)
		}
		_ = entry.Genres()
	})
}
