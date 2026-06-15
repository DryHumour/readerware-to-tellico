package normalize

import (
	"maps"
	"slices"
)

type MarkersConfig struct {
	// Keep is a list of inputs to keep as-is.
	Keep []string
	// Discard is a list of inputs to silently discard (replace with a space)
	Discard []string
	// Scrub is a list of phrases to remove (replace with a space).
	Scrub []string
	// Marker is a map of marker text to canonical markers (extract and replace with a space).
	Marker map[string]string
	// Audit is a map of phrases to explanatory strings for audit reports.
	// The key is the phrase to match, the value is the explanation shown in audit output.
	Audit map[string]string
}

// Clone returns a shallow copy of the configuration.
// Since the slices and maps are of strings, this also amounts to a deep copy.
func (c MarkersConfig) Clone() MarkersConfig {
	return MarkersConfig{
		Keep:    slices.Clone(c.Keep),
		Discard: slices.Clone(c.Discard),
		Scrub:   slices.Clone(c.Scrub),
		Marker:  maps.Clone(c.Marker),
		Audit:   maps.Clone(c.Audit),
	}
}
