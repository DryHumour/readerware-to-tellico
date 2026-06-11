package collection

import (
	"github.com/DryHumour/readerware-to-tellico/internal/normalize"
	"github.com/DryHumour/readerware-to-tellico/internal/strutil"
)

type nameAggregatorHelper interface {
	AggregatorHelper
	Accessor
	Aggregation
	Auditor
}

type markerAggregatorHelper interface {
	MarkerAggregatorHelper
	Accessor
	Aggregation
	Auditor
}

// AggregateNames processes name columns from meta.Clean, populating
// meta.Credits, meta.Markers, and audit fields.
//
// The algorithm runs in two passes to preserve ordering:
//   - Pass 1: names whose annotation matches (or is absent from) the column's default
//     role are appended directly to meta.Credits in column order.
//   - Pass 2: names whose annotation places them in a different role (cross-role) are
//     appended after all primary names, preserving discovery order.
func AggregateNames(agg nameAggregatorHelper, names *normalize.Names) {
	if names == nil {
		return
	}

	var (
		columns = agg.Columns()
		clean   = agg.Clean()
	)

	type crossRoleEntry struct {
		role string
		name string
	}

	var crossRoleQueue []crossRoleEntry

	// Iterate over roles first, then columns in order
	for defaultRole, columnList := range columns.Names {
		for _, col := range columnList {
			val := clean[col]
			if val == "" {
				continue
			}

			for _, part := range strutil.SplitList(val) {
				res := names.Process(part)
				if res.Value == "" {
					continue
				}

				if len(res.Roles) == 0 {
					if defaultRole != "" {
						agg.AddCredit(defaultRole, res.Value)
					}
				} else {
					for _, r := range res.Roles {
						if r == defaultRole {
							agg.AddCredit(r, res.Value)
						} else {
							crossRoleQueue = append(crossRoleQueue, crossRoleEntry{r, res.Value})
						}
					}
				}

				for _, marker := range res.Markers {
					agg.AddMarker(marker)
				}

				agg.AddAuditResult(res)
			}
		}
	}

	for _, cr := range crossRoleQueue {
		agg.AddCredit(cr.role, cr.name)
	}
}

// AggregateMarkers processes marker columns from meta.Clean, populating
// meta.Markers and audit fields.
// This handles entry-level metadata flags embedded in non-name columns (e.g. TITLE,
// EDITION) such as "(out of print)" or "(signed)".
func AggregateMarkers(agg markerAggregatorHelper, markers *normalize.Markers) {
	if markers == nil {
		return
	}
	var (
		columns = agg.Columns()
		clean   = agg.Clean()
	)
	for col := range columns.Markers {
		val := clean[col]
		if val == "" {
			continue
		}
		res := markers.Process(val)
		for _, marker := range res.Markers {
			agg.AddMarker(marker)
		}
		agg.AddAuditResult(res)
	}
}
