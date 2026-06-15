package parallel

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/DryHumour/readerware-to-tellico/internal/tellico/tcfile"
)

// writer writes image data from payloads to the Tellico collection.
// It receives payloads from readers and writes them sequentially to the zip file.
type writer struct {
	tcf     *tcfile.TCFile
	resultC chan<- result
	writerC <-chan *payload
	logger  *slog.Logger
}

func newWriter(logger *slog.Logger, tcf *tcfile.TCFile, resultC chan<- result, writerC <-chan *payload) *writer {
	return &writer{
		tcf:     tcf,
		resultC: resultC,
		writerC: writerC,
		logger:  logger,
	}
}

// Run processes payloads from the writer channel until it is closed or the context is cancelled.
// For each payload, it writes the image data to the Tellico collection.
// Errors are reported to the result channel.
func (w *writer) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case p, ok := <-w.writerC:
			if !ok {
				return
			}
			select {
			case w.resultC <- progressResult(fmt.Sprintf("writing id=%q", p.ID)):
			case <-ctx.Done():
				return
			}

			wr, err := w.tcf.Image(p.Image)
			if err != nil {
				releasePayload(p)
				select {
				case w.resultC <- fatalResult("failed to get image writer", err):
				case <-ctx.Done():
				}
				return
			}
			_, err = wr.Write(p.Data)
			if err != nil {
				releasePayload(p)
				select {
				case w.resultC <- fatalResult("failed to write image data", err):
				case <-ctx.Done():
				}
				return
			}

			releasePayload(p)
		}
	}
}
