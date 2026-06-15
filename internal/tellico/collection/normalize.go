package collection

import "github.com/DryHumour/readerware-to-tellico/internal/normalize"

// Normalize holds the normalizers for names and markers.
type Normalize struct {
	// Names holds the name normalizer.
	Names *normalize.Names
	// Markers holds the marker normalizer.
	Markers *normalize.Markers
}
