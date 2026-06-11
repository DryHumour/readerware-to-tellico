package collection

import (
	"fmt"
	"maps"
	"slices"

	"github.com/DryHumour/readerware-to-tellico/internal/normalize"
)

var (
	_ Aggregation      = (*aggregation)(nil)
	_ Auditor          = (*aggregation)(nil)
	_ AggregatorHelper = (*aggregation)(nil)
)

// aggregation implements Aggregation, Auditor, and AggregatorHelper.
type aggregation struct {
	credits map[string][]string
	markers map[string]bool
	auditor auditor
}

func newAggregation() *aggregation {
	return &aggregation{
		credits: make(map[string][]string),
		markers: make(map[string]bool),
	}
}

// Roles returns a map of role names to their associated names.
func (a *aggregation) Roles() map[string][]string {
	return a.credits
}

// Credits returns a plain-text list of names for the given role, with the
// abbreviation appended to each name when provided.
// For example, Credits("Revisers", "[rev.]") might return ["G. P. Goold [rev.]."].
// Use this for roles not already exposed as named methods.
// Apply xml in the template at the point of XML output.
func (a *aggregation) Credits(role, abbrev string) []string {
	names := a.credits[role]
	if len(names) == 0 {
		return nil
	}
	credits := make([]string, len(names))
	for i, n := range names {
		if abbrev != "" {
			credits[i] = fmt.Sprintf("%s %s", n, abbrev)
		} else {
			credits[i] = n
		}
	}
	return credits
}

// Markers returns a slice of marker strings.
func (a *aggregation) Markers() []string {
	return slices.Collect(maps.Keys(a.markers))
}

// HasMarker reports whether the given entry-level metadata flag was detected
// anywhere in the row (e.g. in the title or edition text).
// Example: {{ if .HasMarker "<signed>" }}
func (a *aggregation) HasMarker(marker string) bool {
	return a.markers[marker]
}

// AddCredit adds a name to the specified role's credit list.
// De-duplicates within the role.
func (a *aggregation) AddCredit(role, name string) (_ Nothing) {
	if a.credits[role] == nil {
		a.credits[role] = []string{}
	}
	for _, existing := range a.credits[role] {
		if existing == name {
			return
		}
	}
	a.credits[role] = append(a.credits[role], name)
	return
}

// AddMarker adds a marker to the entry.
func (a *aggregation) AddMarker(marker string) (_ Nothing) {
	a.markers[marker] = true
	return
}

// AddCreditResult adds the normalized value from a Result to the specified role's credit list,
// along with any markers, and sets audit if required.
func (a *aggregation) AddCreditResult(role string, r normalize.Result) (_ Nothing) {
	if r.Value == "" {
		return
	}
	a.AddCredit(role, r.Value)
	for _, marker := range r.Markers {
		a.AddMarker(marker)
	}
	a.auditor.AddAuditResult(r)
	return
}

// AddMarkerResult adds markers from a Result to the entry and sets audit if required.
func (a *aggregation) AddMarkerResult(r normalize.Result) (_ Nothing) {
	for _, marker := range r.Markers {
		a.AddMarker(marker)
	}
	a.auditor.AddAuditResult(r)
	return
}

// RequiresAudit reports whether this entry was flagged for human review.
func (a *aggregation) RequiresAudit() bool {
	return a.auditor.RequiresAudit()
}

// AuditReasons returns the list of reasons why this entry needs human review.
func (a *aggregation) AuditReasons() []string {
	return a.auditor.AuditReasons()
}

// AddAudit flags the entry for audit with optional reasons.
func (a *aggregation) AddAudit(reasons ...string) Nothing {
	return a.auditor.AddAudit(reasons...)
}

// AddAuditResult adds audit information from a normalize.Result.
func (a *aggregation) AddAuditResult(r normalize.Result) Nothing {
	return a.auditor.AddAuditResult(r)
}
