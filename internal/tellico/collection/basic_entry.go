package collection

import (
	"github.com/DryHumour/readerware-to-tellico/internal/images"
	"github.com/DryHumour/readerware-to-tellico/internal/normalize"
)

var (
	_ Entry = (*basicEntry)(nil)
)

type basicEntry struct {
	normalize Normalize
	row       *rowData
	agg       *aggregation
}

func newBasicEntry(info CollectionInfo, clean map[string]string, images images.Row) (*basicEntry, error) {
	normalize, err := info.Normalize()
	if err != nil {
		return nil, err
	}
	return &basicEntry{
		normalize: normalize,
		row:       newRowData(info, clean, images),
		agg:       newAggregation(),
	}, nil
}

//
//  Accessor Interface Methods
//

// Columns returns the column configuration for this collection.
func (e *basicEntry) Columns() ColumnConfig {
	return e.row.Columns()
}

// Clean returns the plain-text cleaned column values for this entry.
func (e *basicEntry) Clean() map[string]string {
	return e.row.Clean()
}

// Images returns the cover image data for this entry.
func (e *basicEntry) Images() images.Row {
	return e.row.Images()
}

// V returns the plain-text cleaned value for the given column.
// Apply xml in the template at the point of XML output: {{ .V "COL" | xml }}
func (e *basicEntry) V(col string) string {
	return e.row.V(col)
}

// L returns the plain-text split values for a semicolon or slash separated column.
// Apply xml in the template at the point of XML output: {{ range .L "COL" }}{{ . | xml }}{{ end }}
func (e *basicEntry) L(col string) []string {
	return e.row.L(col)
}

// D returns a TellicoDate value for the given column.
func (e *basicEntry) D(col string) *TellicoDate {
	return e.row.D(col)
}

// Is returns true if the given column is "true".
func (e *basicEntry) Is(col string) bool {
	return e.row.Is(col)
}

// Categories returns the plain-text Readerware category values for this entry.
func (e *basicEntry) Categories() []string {
	return e.row.Categories()
}

//
//  Aggregator Interface Methods
//

// Aggregate performs the full aggregation using the built-in logic.
func (e *basicEntry) Aggregate() (_ Nothing) {
	e.AggregateNames()
	e.AggregateMarkers()
	return
}

//
//  AggregatorHelper Interface Methods
//

// AggregateNames processes name columns from meta.Clean, populating
// meta.Credits, meta.Markers, and audit fields using the built-in logic.
func (e *basicEntry) AggregateNames() (_ Nothing) {
	AggregateNames(e, e.normalize.Names)
	return
}

// AggregateMarkers processes marker columns from meta.Clean, populating
// meta.Markers and audit fields using the built-in logic.
func (e *basicEntry) AggregateMarkers() (_ Nothing) {
	AggregateMarkers(e, e.normalize.Markers)
	return
}

// AddCredit adds a name to the specified role's credit list.
// De-duplicates within the role.
func (e *basicEntry) AddCredit(role, name string) Nothing {
	return e.agg.AddCredit(role, name)
}

// AddMarker adds a marker to the entry.
func (e *basicEntry) AddMarker(marker string) Nothing {
	return e.agg.AddMarker(marker)
}

// AddCreditResult adds the normalized value from a Result to the specified role's credit list,
// along with any markers, and sets audit if required.
func (e *basicEntry) AddCreditResult(role string, r normalize.Result) Nothing {
	return e.agg.AddCreditResult(role, r)
}

// AddMarkerResult adds markers from a Result to the entry and sets audit if required.
func (e *basicEntry) AddMarkerResult(r normalize.Result) Nothing {
	return e.agg.AddMarkerResult(r)
}

//
//  Aggregation Interface Methods
//

// Roles returns a map of role names to their associated names.
func (e *basicEntry) Roles() map[string][]string {
	return e.agg.Roles()
}

// Credits returns a plain-text list of names for the given role, with the
// abbreviation appended to each name when provided.
// For example, Credits("Revisers", "[rev.]") might return ["G. P. Goold [rev.]."].
// Use this for roles not already exposed as named methods.
// Apply xml in the template at the point of XML output.
func (e *basicEntry) Credits(role, abbrev string) []string {
	return e.agg.Credits(role, abbrev)
}

// Markers returns a slice of marker strings.
func (e *basicEntry) Markers() []string {
	return e.agg.Markers()
}

// HasMarker reports whether the given entry-level metadata flag was detected
// anywhere in the row (e.g. in the title or edition text).
// Example: {{ if .HasMarker "<signed>" }}
func (e *basicEntry) HasMarker(marker string) bool {
	return e.agg.HasMarker(marker)
}

//
//  Auditor Interface Methods
//

// RequiresAudit reports whether this entry was flagged for human review.
func (e *basicEntry) RequiresAudit() bool {
	return e.agg.RequiresAudit()
}

// AuditReasons returns the list of reasons why this entry needs human review.
func (e *basicEntry) AuditReasons() []string {
	return e.agg.AuditReasons()
}

// AddAudit flags the entry for audit with an optional reason.
func (e *basicEntry) AddAudit(reasons ...string) Nothing {
	return e.agg.AddAudit(reasons...)
}

// AddAuditResult adds audit information from a normalize.Result.
func (e *basicEntry) AddAuditResult(r normalize.Result) Nothing {
	return e.agg.AddAuditResult(r)
}
