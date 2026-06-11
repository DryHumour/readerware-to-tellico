package normalize

import (
	"testing"
)

// FuzzProcessMarkers targets the entire Markers pipeline.
func FuzzProcessMarkers(f *testing.F) {
	f.Add("Title (signed) [noise]")
	f.Add("KeepMe")
	f.Add("DiscardMe")

	config := MarkersConfig{
		Keep:    []string{"KeepMe"},
		Discard: []string{"DiscardMe"},
		Scrub:   []string{"noise"},
		Marker:  map[string]string{"(signed)": "Signed"},
		Audit:   map[string]string{"audit-me": "[audit]"},
	}
	m, err := NewMarkers(config)
	if err != nil {
		f.Fatalf("failed to create Markers: %v", err)
	}

	f.Fuzz(func(t *testing.T, input string) {
		_ = m.Process(input)
	})
}
