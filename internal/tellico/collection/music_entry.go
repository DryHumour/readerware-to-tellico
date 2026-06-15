package collection

import (
	"github.com/DryHumour/readerware-to-tellico/internal/images"
)

var (
	_ Entry = (*MusicEntry)(nil)
)

// MusicEntry is the data object passed to every music entry template.
// All fields hold plain, unescaped text. Template authors must apply
// the xml function at the point of XML output (e.g. {{ .V "TITLE" | xml }}
// is handled by e.g. V, but slice-returning methods return plain strings).
type MusicEntry struct {
	*basicEntry
}

func newMusicEntry(info CollectionInfo, clean map[string]string, images images.Row) (*MusicEntry, error) {
	basic, err := newBasicEntry(info, clean, images)
	if err != nil {
		return nil, err
	}
	return &MusicEntry{
		basicEntry: basic,
	}, nil
}

//
//  MusicEntry Methods
//

// Artists returns the plain-text list of primary artists in discovery order.
func (e *MusicEntry) Artists() []string { return e.agg.Roles()["Artists"] }

// Composers returns the plain-text list of composers in discovery order.
func (e *MusicEntry) Composers() []string { return e.agg.Roles()["Composers"] }

// Conductors returns the plain-text list of conductors in discovery order.
func (e *MusicEntry) Conductors() []string { return e.agg.Roles()["Conductors"] }

// Labels returns the plain-text list of labels in discovery order.
func (e *MusicEntry) Labels() []string { return e.row.L("PUBLISHER") } // FIXME(nschelle) placeholder for now; should sanitize and audit

// Orchestras returns the plain-text list of orchestras in discovery order.
func (e *MusicEntry) Orchestras() []string { return e.agg.Roles()["Orchestras"] }

// Soloists returns the plain-text list of soloists in discovery order.
func (e *MusicEntry) Soloists() []string { return e.agg.Roles()["Soloists"] }

// Genres returns the plain-text genre values derived from Readerware category paths.
func (e *MusicEntry) Genres() []string {
	return e.row.Genres()
}
