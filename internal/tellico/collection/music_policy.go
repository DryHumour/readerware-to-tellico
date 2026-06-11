package collection

import (
	"context"
	"fmt"
	"slices"

	"github.com/DryHumour/readerware-to-tellico/internal/config"
	"github.com/DryHumour/readerware-to-tellico/internal/images"
)

var (
	_ Policy = (*MusicPolicy)(nil)

	musicTemplateNames = TemplateNames{
		Config: "music.config",
		Header: "music.header",
		Entry:  "music.entry",
		Footer: "music.footer",
	}
)

// MusicPolicy implements collection policy behavior for Readerware music exports.
type MusicPolicy struct {
	// info holds the collection configuration for this policy.
	info collectionInfo
}

// NewMusicPolicy creates a music policy.
func NewMusicPolicy(_ context.Context, _ *config.Config) *MusicPolicy {
	return &MusicPolicy{
		info: collectionInfo{
			kind:          KindMusic,
			templateNames: musicTemplateNames,
			blocklist:     make(map[string]bool),
		},
	}
}

func (p *MusicPolicy) Info() CollectionInfo {
	return &p.info
}

func (p *MusicPolicy) ConfigureHeaders(headers []string, imagesEnabled bool) error {
	headerSet := make(map[string]bool, len(headers))
	for _, h := range headers {
		if headerSet[h] {
			return fmt.Errorf("duplicate header: %q appears more than once", h)
		}
		headerSet[h] = true
	}
	if !headerSet["TITLE"] {
		return fmt.Errorf("missing required header: TITLE")
	}
	if imagesEnabled {
		if !headerSet["ROWKEY"] {
			return fmt.Errorf("missing required header: ROWKEY is required when image directories are specified")
		}
	}
	if !imagesEnabled && !headerSet["ROWKEY"] && !headerSet["ROW#"] {
		return fmt.Errorf("missing required header: either ROWKEY or ROW# must be present")
	}
	p.info.columns.Headers = slices.Clone(headers)
	return nil
}

func (p *MusicPolicy) NewEntry(clean map[string]string, img images.Row) (Entry, error) {
	return newMusicEntry(p.Info(), clean, img)
}
