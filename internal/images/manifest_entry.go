package images

import (
	"fmt"
	"os"

	"github.com/DryHumour/readerware-to-tellico/internal/config"
)

type ManifestEntry struct {
	Path     string
	Position int
	IsLarge  bool
	ID       string
	Format   string
	IsGIF    bool
	Info     os.FileInfo
	IsUsed   bool
}

func (m *ManifestEntry) Slot() config.Slot {
	if m.IsLarge {
		return config.SlotLarge1 + config.Slot(m.Position-1)
	}
	return config.Slot1 + config.Slot(m.Position-1)
}

func (m *ManifestEntry) Comment() string {
	desc := "image"
	if m.IsLarge {
		desc = "large image"
	}
	return fmt.Sprintf("Readerware %s %d", desc, m.Position)
}

func (m *ManifestEntry) Use() *ManifestEntry {
	m.IsUsed = true
	return m
}
