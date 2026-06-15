package collection

import (
	"github.com/DryHumour/readerware-to-tellico/internal/images"
)

var (
	_ Entry = (*BooksEntry)(nil)
)

// BooksEntry is the data object passed to every books entry template.
// All fields hold plain, unescaped text. Template authors must apply
// the xml function at the point of XML output (e.g. {{ .V "TITLE" | xml }}
// is handled by e.g. V, but slice-returning methods return plain strings).
type BooksEntry struct {
	*basicEntry
}

func newBooksEntry(info CollectionInfo, clean map[string]string, images images.Row) (*BooksEntry, error) {
	basic, err := newBasicEntry(info, clean, images)
	if err != nil {
		return nil, err
	}
	return &BooksEntry{
		basicEntry: basic,
	}, nil
}

//
//  BookEntry Methods
//

// Authors returns the plain-text list of primary authors in discovery order.
func (e *BooksEntry) Authors() []string { return e.agg.Roles()["Authors"] }

// Editors returns the plain-text list of editors in discovery order.
func (e *BooksEntry) Editors() []string { return e.agg.Roles()["Editors"] }

// Translators returns the plain-text list of translators in discovery order.
func (e *BooksEntry) Translators() []string { return e.agg.Roles()["Translators"] }

// Illustrators returns the plain-text list of illustrators in discovery order.
func (e *BooksEntry) Illustrators() []string { return e.agg.Roles()["Illustrators"] }

// Genres returns the plain-text genre values derived from Readerware category paths.
func (e *BooksEntry) Genres() []string {
	return e.row.Genres()
}
