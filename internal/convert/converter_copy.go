package convert

import (
	"context"
	"iter"

	"github.com/DryHumour/readerware-to-tellico/internal/convert/copier/parallel"
	"github.com/DryHumour/readerware-to-tellico/internal/convert/copier/simple"
)

// copyAllImages returns an iterator that yields reports for each copied image file.
// It copies all image files from the source filesystems into the TC archive.
func (c *Converter) copyAllImages(ctx context.Context) iter.Seq2[Report, error] {
	return func(yield func(Report, error) bool) {
		logger := c.cfg.Logger
		var copier Copier
		if c.cfg.Concurrency > 0 {
			copier = parallel.New(logger, c.tcf, c.cfg.Concurrency)
		} else {
			copier = simple.New(logger, c.tcf)
		}
		for report, err := range copier.CopyAll(ctx, c.images.AllUsed()) {
			if !yield(report, err) {
				return
			}
		}
	}
}
