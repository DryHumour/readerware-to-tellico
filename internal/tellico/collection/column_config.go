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
	// columnRoleReverse is a lazy cache for the reverse mapping (column → role).
	columnRoleReverse map[string]string
}

func (c ColumnConfig) Clone() ColumnConfig {
	return ColumnConfig{
		Names:             maps.Clone(c.Names),
		Markers:           maps.Clone(c.Markers),
		Categories:        maps.Clone(c.Categories),
		Headers:           slices.Clone(c.Headers),
		columnRoleReverse: nil, // Reverse mapping is lazy, don't clone
	}
}

func (c ColumnConfig) ColumnRole(column string) string {
	// Build reverse mapping lazily on first call
	if c.columnRoleReverse == nil {
		c.columnRoleReverse = make(map[string]string)
		for role, columns := range c.Names {
			for _, col := range columns {
				c.columnRoleReverse[col] = role
			}
		}
	}
	if role, ok := c.columnRoleReverse[column]; ok {
		return role
	}
	return column
}
