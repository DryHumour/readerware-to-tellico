package normalize

import (
	"maps"
	"slices"
)

type NamesConfig struct {
	// Keep is a list of names to keep as-is.
	Keep []string
	// Discard is a list of names to silently discard (both after scrubbing, and also after flipping)
	Discard []string
	// Scrub is a list of phrases to remove from names.
	Scrub []string
	// Role is a map of role patterns to canonical roles.
	Role map[string]string
	// Markers is a map of markers to canonical markers.
	Marker map[string]string
	// Suffixes is a list of name suffixes e.g. "Ph.D.".
	Suffixes []string
	// Honorifics is a list of honorifics e.g. "the right hon.".
	Honorifics []string
	// Corporate is a list of corporate signifiers e.g. "ltd.".
	Corporate []string
	// Collaboration is a list of collaboration signifiers e.g. "featuring".
	Collaboration []string
	// Audit is a map of phrases to explanatory strings for audit reports.
	// The key is the phrase to match, the value is the explanation shown in audit output.
	Audit map[string]string
}

// Clone returns a shallow copy of the configuration.
// Since the slices and maps are of strings, this also amounts to a deep copy.
func (c NamesConfig) Clone() NamesConfig {
	return NamesConfig{
		Keep:          slices.Clone(c.Keep),
		Discard:       slices.Clone(c.Discard),
		Scrub:         slices.Clone(c.Scrub),
		Role:          maps.Clone(c.Role),
		Marker:        maps.Clone(c.Marker),
		Suffixes:      slices.Clone(c.Suffixes),
		Honorifics:    slices.Clone(c.Honorifics),
		Corporate:     slices.Clone(c.Corporate),
		Collaboration: slices.Clone(c.Collaboration),
		Audit:         maps.Clone(c.Audit),
	}
}
