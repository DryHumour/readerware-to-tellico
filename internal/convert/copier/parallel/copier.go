package parallel

import (
	"context"
	"iter"
	"log/slog"
	"sync"

	"github.com/DryHumour/readerware-to-tellico/internal/images"
	"github.com/DryHumour/readerware-to-tellico/internal/tellico/tcfile"
)

// parallelCopier manages parallel copying of image files from a manifest into a Tellico collection.
// It uses a pool of reader goroutines to read files concurrently and a single writer
// goroutine to write them sequentially (since the zip format cannot be written to in parallel).
type parallelCopier struct {
	tcf         *tcfile.TCFile
	concurrency int
	logger      *slog.Logger
}

// New creates a new copier for writing images to the given Tellico file.
func New(logger *slog.Logger, tcf *tcfile.TCFile, concurrency int) parallelCopier {
	if concurrency < 1 {
		concurrency = 1
	}
	return parallelCopier{
		tcf:         tcf,
		concurrency: concurrency,
		logger:      logger,
	}
}

// CopyAll copies all image files from the manifest entries into the Tellico collection.
// It uses parallel readers to read files concurrently and a single writer to write them.
// The returned iterator yields reports and errors as they occur.
// Context cancellation will stop all processing and close the iterator.
func (p parallelCopier) CopyAll(ctx context.Context, entries iter.Seq[*images.ManifestEntry]) iter.Seq2[Report, error] {
	return func(yield func(Report, error) bool) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		readerC := make(chan *images.ManifestEntry)
		writerC := make(chan *payload)
		resultC := make(chan result, 2*(p.concurrency+1)) // enough for each reader and the writer to report progress and final result

		// start the feeder (iterator to reader channel)
		go func() {
			defer close(readerC) // begin readers shutdown
			defer p.logger.DebugContext(ctx, "feeder done")
			for entry := range entries {
				select {
				case readerC <- entry:
				case <-ctx.Done():
					return
				}
			}
		}()

		// start the readers group
		var rdG sync.WaitGroup
		rdG.Add(p.concurrency)
		for id := range p.concurrency {
			go func() {
				defer rdG.Done()
				logger := p.logger.With("reader", id+1)
				defer logger.DebugContext(ctx, "reader done")
				newReader(logger, resultC, writerC, readerC).Run(ctx)
			}()
		}

		// start the (sole) writer (zip cannot be written to in parallel)
		var wrG sync.WaitGroup
		wrG.Add(1)
		go func() {
			defer wrG.Done()
			logger := p.logger.With("writer", 1)
			defer logger.DebugContext(ctx, "writer done")
			newWriter(logger, p.tcf, resultC, writerC).Run(ctx)
		}()

		// start the coordinator
		go func() {
			rdG.Wait()
			close(writerC) // begin writer shutdown
			wrG.Wait()
			close(resultC) // begin yield loop shutdown
		}()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			select {
			case result, ok := <-resultC:
				if !ok {
					return
				}
				if !yield(result.Report, result.Err) || result.Err != nil {
					cancel()
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}
}
