package collection

import (
	"github.com/DryHumour/readerware-to-tellico/internal/images"
	"github.com/DryHumour/readerware-to-tellico/internal/normalize"
)

// Entry represents a single entry in the collection.
// Entry is the interface that entry templates receive.
type Entry interface {
	Accessor
	Aggregator
	AggregatorHelper
	Aggregation
	Auditor
}

// Accessor provides access to entry data for templates.
type Accessor interface {
	// Columns returns the column configuration for this collection.
	Columns() ColumnConfig
	// Clean returns the plain-text cleaned column values for this entry.
	Clean() map[string]string
	// Images returns the image data for this entry.
	Images() images.Row
	// Categories returns the plain-text Readerware category values for this entry.
	Categories() []string
	// V returns the plain-text cleaned value for the given column.
	// Apply xml in the template at the point of XML output: {{ .V "COL" | xml }}
	V(col string) string
	// L returns the plain-text split values for a semicolon or slash separated column.
	// Apply xml in the template at the point of XML output: {{ range .L "COL" }}{{ . | xml }}{{ end }}
	L(col string) []string
	// D returns a TellicoDate value for the given column.
	D(col string) *TellicoDate
	// Is returns true if the given column is, excatly, "true".
	Is(col string) bool
}

// Aggregator performs default aggregation of entry data.
type Aggregator interface {
	// Aggregate performs the full aggregation using the built-in logic.
	Aggregate() Nothing
	// AggregateNames processes name columns from meta.Clean, populating
	// meta.Credits, meta.Markers, and audit fields using the built-in logic.
	AggregateNames() Nothing
	// AggregateMarkers processes marker columns from meta.Clean, populating
	// meta.Markers and audit fields using the built-in logic.
	AggregateMarkers() Nothing
}

// AggregatorHelper provides helper methods for user templates to implement aggregation.
type AggregatorHelper interface {
	NameAggregatorHelper
	MarkerAggregatorHelper
}

// NameAggregatorHelper provides helper methods for user name aggregation templates.
type NameAggregatorHelper interface {
	// AddCredit adds a name to the specified role's credit list.
	// De-duplicates within the role.
	AddCredit(role, name string) Nothing
	// AddCreditResult adds the normalized value from a Result to the specified role's credit list,
	// along with any markers, and sets audit if required.
	AddCreditResult(role string, r normalize.Result) Nothing
}

// MarkerAggregatorHelper provides helper methods for user marker aggregation templates.
type MarkerAggregatorHelper interface {
	// AddMarker adds a marker to the entry.
	AddMarker(marker string) Nothing
	// AddMarkerResult adds markers from a Result to the entry and sets audit if required.
	AddMarkerResult(r normalize.Result) Nothing
}

// Aggregation provides access to aggregated data for templates.
type Aggregation interface {
	// Roles returns a map of role names to their associated names.
	Roles() map[string][]string
	// Credits returns a plain-text list of names for the given role, with the
	// abbreviation appended to each name when provided.
	// For example, Credits("Revisers", "[rev.]") might return ["G. P. Goold [rev.]"].
	Credits(role, abbrev string) []string
	// Markers returns a slice of marker strings.
	Markers() []string
	// HasMarker reports whether the given entry-level metadata flag was detected
	// anywhere in the row (e.g. in the title or edition text).
	// Example: {{ if .HasMarker "<signed>" }}
	HasMarker(marker string) bool
}

type Auditor interface {
	// RequiresAudit reports whether this entry was flagged for human review.
	RequiresAudit() bool
	// AuditReasons returns the list of reasons why this entry needs human review.
	AuditReasons() []string
	// AddAudit flags the entry for audit with optional reasons.
	AddAudit(reasons ...string) Nothing
	// AddAuditResult adds audit information from a normalize.Result.
	AddAuditResult(r normalize.Result) Nothing
}

type Nothing string
