package collection

import (
	"maps"
	"slices"
)

// ColumnConfig describes which CSV columns contain credits and markers
// for a specific collection type (Books, Music, Video).
type ColumnConfig struct {
	// Names maps a canonical role name to its ordered CSV column names.
	// Example: "Authors" ⟶ ["AUTHOR", "AUTHOR2", "AUTHOR3"]
	Names map[string][]string
	// Markers is the list of non-name columns to scan for entry-level
	// metadata markers (e.g. TITLE containing "(out of print)").
	Markers map[string]bool
	// Categories is the list of columns that contain category information.
	Categories map[string]bool
	// Headers is the list of CSV headers as read from the export file.
	Headers []string
	// columnRoleReverse is the reverse mapping (column → role), populated when
	// Columns.Names is configured and shared by all clones.
	columnRoleReverse map[string]string
}

func (c ColumnConfig) Clone() ColumnConfig {
	return ColumnConfig{
		Names:             maps.Clone(c.Names),
		Markers:           maps.Clone(c.Markers),
		Categories:        maps.Clone(c.Categories),
		Headers:           slices.Clone(c.Headers),
		columnRoleReverse: maps.Clone(c.columnRoleReverse),
	}
}

func (c ColumnConfig) ColumnRole(column string) string {
	rev := c.columnRoleReverse
	if rev == nil {
		// Unshared fallback for configs not built through Columns().
		rev = newColumnRoleReverse(c.Names)
	}
	if role, ok := rev[column]; ok {
		return role
	}
	return column
}

func newColumnRoleReverse(names map[string][]string) map[string]string {
	if names == nil {
		return nil
	}
	m := make(map[string]string)
	for role, columns := range names {
		for _, col := range columns {
			m[col] = role
		}
	}
	return m
}
