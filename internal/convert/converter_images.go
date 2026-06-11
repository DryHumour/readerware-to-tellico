package convert

import (
	"context"

	"github.com/DryHumour/readerware-to-tellico/internal/images"
)

// rowImages returns the images.Row for the given row key from the image index.
func (c *Converter) rowImages(rowKey string) images.Row {
	row, _ := c.images.Row(rowKey)
	return row
}

// buildFooterData constructs the FooterData structure needed for rendering
// the XML footer template, which includes the image manifest.
func (c *Converter) buildFooterData(ctx context.Context) FooterData {
	return FooterData{
		Images: func(yield func(ImageData) bool) {
			c.allImageElements(ctx, yield)
		},
	}
}

// allImageElements yields ImageData for all used images in the index.
// It respects context cancellation and stops if the yield function returns false.
func (c *Converter) allImageElements(ctx context.Context, yield func(ImageData) bool) {
	for entry := range c.images.AllUsed() {
		if err := context.Cause(ctx); err != nil {
			return
		}
		if !yield(ImageData{ID: entry.ID, Format: entry.Format}) {
			return
		}
	}
}
