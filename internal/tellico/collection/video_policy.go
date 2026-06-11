package collection

import (
	"context"
	"fmt"
	"slices"

	"github.com/DryHumour/readerware-to-tellico/internal/config"
	"github.com/DryHumour/readerware-to-tellico/internal/images"
)

var (
	_ Policy = (*VideoPolicy)(nil)

	videoTemplateNames = TemplateNames{
		Config: "video.config",
		Header: "video.header",
		Entry:  "video.entry",
		Footer: "video.footer",
	}
)

// VideoPolicy implements collection policy behavior for Readerware video exports.
type VideoPolicy struct {
	// info holds the collection configuration for this policy.
	info collectionInfo
}

// NewVideoPolicy creates a video policy.
func NewVideoPolicy(_ context.Context, _ *config.Config) *VideoPolicy {
	return &VideoPolicy{
		info: collectionInfo{
			kind:          KindVideo,
			templateNames: videoTemplateNames,
			blocklist:     make(map[string]bool),
		},
	}
}

func (p *VideoPolicy) Info() CollectionInfo {
	return &p.info
}

func (p *VideoPolicy) ConfigureHeaders(headers []string, imagesEnabled bool) error {
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

func (p *VideoPolicy) NewEntry(clean map[string]string, img images.Row) (Entry, error) {
	return newVideoEntry(p.Info(), clean, img)
}
