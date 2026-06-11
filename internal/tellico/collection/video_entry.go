package collection

import (
	"github.com/DryHumour/readerware-to-tellico/internal/images"
)

var (
	_ Entry = (*VideoEntry)(nil)
)

// VideoEntry is the data object passed to every video entry template.
// All fields hold plain, unescaped text. Template authors must apply
// the xml function at the point of XML output (e.g. {{ .V "TITLE" | xml }}
// is handled by e.g. V, but slice-returning methods return plain strings).
type VideoEntry struct {
	*basicEntry
}

func newVideoEntry(info CollectionInfo, clean map[string]string, images images.Row) (*VideoEntry, error) {
	basic, err := newBasicEntry(info, clean, images)
	if err != nil {
		return nil, err
	}
	return &VideoEntry{
		basicEntry: basic,
	}, nil
}

//	VideoEntry Methods
//
// Authors returns the plain-text list of authors in discovery order.
func (e *VideoEntry) Authors() []string { return e.agg.Roles()["Authors"] }

// Cast returns the plain-text list of cast members in discovery order.
func (e *VideoEntry) Cast() []string { return e.agg.Roles()["Cast"] }

// Composers returns the plain-text list of composers in discovery order.
func (e *VideoEntry) Composers() []string { return e.agg.Roles()["Composers"] }

// Directors returns the plain-text list of directors in discovery order.
func (e *VideoEntry) Directors() []string { return e.agg.Roles()["Directors"] }

// Editors returns the plain-text list of editors in discovery order.
func (e *VideoEntry) Editors() []string { return e.agg.Roles()["Editors"] }

// Photographers returns the plain-text list of photographers in discovery order.
func (e *VideoEntry) Photographers() []string { return e.agg.Roles()["Photographers"] }

// Producers returns the plain-text list of producers in discovery order.
func (e *VideoEntry) Producers() []string { return e.agg.Roles()["Producers"] }

// Screenwriters returns the plain-text list of screenwriters in discovery order.
func (e *VideoEntry) Screenwriters() []string { return e.agg.Roles()["Screenwriters"] }

// Writers returns the plain-text list of writers in discovery order.
func (e *VideoEntry) Writers() []string { return e.agg.Roles()["Writers"] }

// Genres returns the plain-text genre values derived from Readerware category paths.
func (e *VideoEntry) Genres() []string {
	return e.row.Genres()
}
