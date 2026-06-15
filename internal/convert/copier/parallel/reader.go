package parallel

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/DryHumour/readerware-to-tellico/internal/convert/copier"
	"github.com/DryHumour/readerware-to-tellico/internal/images"
	"github.com/DryHumour/readerware-to-tellico/internal/tellico/tcfile"
)

// reader reads image files from the manifest and sends them to the writer.
// It validates the image format and reports errors to the result channel.
type reader struct {
	resultC chan<- result
	writerC chan<- *payload
	readerC <-chan *images.ManifestEntry
	logger  *slog.Logger
}

func newReader(logger *slog.Logger, resultC chan<- result, writerC chan<- *payload, readerC <-chan *images.ManifestEntry) *reader {
	return &reader{
		resultC: resultC,
		writerC: writerC,
		readerC: readerC,
		logger:  logger,
	}
}

// report sends a result to the result channel, respecting context cancellation.
// It returns an error if the context is cancelled, allowing the caller to bail out.
func (r *reader) report(ctx context.Context, res result) error {
	select {
	case r.resultC <- res:
		return nil
	case <-ctx.Done():
		return context.Cause(ctx)
	}
}

// Run processes entries from the reader channel until it is closed or the context is cancelled.
// For each entry, it creates an image header, reads the file, and sends the payload to the writer.
// Errors are reported to the result channel.
func (r *reader) Run(ctx context.Context) {
	for {
		// give priority to cancellation
		select {
		case <-ctx.Done():
			return
		default:
		}
		// check for work
		select {
		case entry, ok := <-r.readerC:
			if !ok {
				return
			}

			// issue a progress report
			if err := r.report(ctx, progressResult(fmt.Sprintf("reading %s", entry.Path))); err != nil {
				return
			}

			// create the image header
			img, err := tcfile.NewImage(entry.ID, entry.Info, entry.Comment())
			if err != nil {
				_ = r.report(ctx, fatalImageResult("failed to create image header", entry, err))
				return
			}

			// set up a payload (ideally recycled from the pool)
			p := acquirePayload()
			p.ID = entry.ID
			p.Image = img

			// process the image file (possibly replacing the payload e.g. if oversized, a GIF, etc.)
			payload, err := r.processImageFile(ctx, p, entry)
			if err != nil {
				releasePayload(p)
				_ = r.report(ctx, fatalResult("failed to process image file", err))
				return
			}

			if p != payload {
				// release original payload if it was replaced e.g. oversized, a GIF, etc.
				releasePayload(p)
			}

			// send the payload to the writer
			select {
			case r.writerC <- payload:
			case <-ctx.Done():
				releasePayload(payload)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// processImageFile reads the image file and checks its format.
// It returns the payload and any error that occurred during processing.
func (r *reader) processImageFile(ctx context.Context, p *payload, entry *images.ManifestEntry) (*payload, error) {
	// open the file
	if err := context.Cause(ctx); err != nil {
		return nil, err
	}
	f, err := os.Open(entry.Path)
	if err != nil {
		return nil, copier.NewFileError("failed to open file", entry, err)
	}
	defer f.Close()

	// fill the pre-allocated buffer
	//
	// Although io.ReadFull could block on a network resource, there is little we can do about it here.
	// If the user interrupts the program, the goroutine implementing the iterator in CopyAll will be
	// cancelled and it will return.  That at least allows its caller to make forward progress.  If the
	// I/O eventually unwedges, then the reader goroutines will cleanly tidy themselves up.  Otherwise
	// they will be blocked until the runtime exits.  The feeder, writer, and coordinator will all also
	// tidy themselves up when the context is cancelled.
	if err := context.Cause(ctx); err != nil {
		return nil, err
	}
	n, err := io.ReadFull(f, p.Data)
	switch err {
	case nil, io.ErrUnexpectedEOF:
		p.Data = p.Data[:n]
	case io.EOF:
		return nil, copier.NewFileEmptyError(entry)
	default:
		return nil, copier.NewFileError("failed to read file", entry, err)
	}

	// check that the format actually matches
	format, err := images.DetectFormat(p.Data)
	if err != nil {
		return nil, copier.NewFileError("failed to detect supported image format", entry, err)
	}
	if (!entry.IsGIF && format != entry.Format) || (entry.IsGIF && format != "gif") {
		warnErr := copier.NewFormatMismatchError(entry, format)
		if err := r.report(ctx, warnResult("problem with image file", entry, warnErr)); err != nil {
			return nil, err
		}
	}

	// check if file fits in buffer
	if len(p.Data) >= cap(p.Data) {
		// buffer full, delegate reading the remainder (if any)
		switch p, err = r.readOversizedData(ctx, f, p, entry); err {
		case nil:
			// read more data, no errors encountered
			warnErr := copier.NewFileTooLargeError(entry, int64(len(p.Data)))
			if err := r.report(ctx, warnResult("problem with image file", entry, warnErr)); err != nil {
				return nil, err
			}
		case io.EOF:
			// no more data to read, proceed with payload as-is
		default:
			return nil, err
		}
	}

	// deal with GIF files
	if entry.IsGIF {
		if p, err = r.convertToPNG(ctx, p, entry); err != nil {
			return nil, err
		}
		if err := r.report(ctx, infoResult("GIF converted to PNG", entry, nil)); err != nil {
			return nil, err
		}
	}

	return p, nil
}

// readOversizedData handles files that exceed the pre-allocated buffer size.
// It reads the remaining data from the file and constructs a new oversized payload.
// The caller is responsible for releasing the original payload.
// If io.EOF is returned, no further data needed to be read (and the returned payload
// is the original payload object.)
func (r *reader) readOversizedData(ctx context.Context, f *os.File, p *payload, entry *images.ManifestEntry) (*payload, error) {
	// check if there's more
	if err := context.Cause(ctx); err != nil {
		return nil, err
	}
	var peekBuf [1]byte
	switch n, err := f.Read(peekBuf[:]); err {
	case nil:
		// file is (unexpectedly) larger than the pre-allocated buffer
	case io.EOF:
		if n == 0 {
			// file fits exactly in buffer
			return p, io.EOF
		}
		// file is (unexpectedly) larger than the pre-allocated buffer
	default:
		return nil, copier.NewFileError("failed to read file", entry, err)
	}

	// file is (unexpectedly) larger than the pre-allocated buffer
	if err := context.Cause(ctx); err != nil {
		return nil, err
	}
	rest, err := io.ReadAll(f)
	if err != nil {
		return nil, copier.NewFileError("failed to read file", entry, err)
	}

	// construct final payload with full content
	oversized := &payload{
		ID:    p.ID,
		Image: p.Image,
		Data:  make([]byte, 0, len(p.Data)+1+len(rest)),
	}
	oversized.Data = append(oversized.Data, p.Data...)
	oversized.Data = append(oversized.Data, peekBuf[0])
	oversized.Data = append(oversized.Data, rest...)

	return oversized, nil
}

// convertToPNG converts a GIF image to PNG format.
//
// We choose to just convert GIFs to PNGs when adding them to the Tellico collection.
// Although Tellico supports GIFs via Qt QImage, it handles them by on-the-fly
// conversion to in-memory PNGs on each access.  This is inefficient, so we convert
// them to PNGs up front.  (Tellico also logs an unstructured message to stderr each
// time it does its conversion, which is annoying.)
func (r *reader) convertToPNG(_ context.Context, p *payload, entry *images.ManifestEntry) (result *payload, err error) {
	cpng := copier.NewConvertedPNG(entry.Info)
	fsInfo, _ := cpng.Stat() // cannot fail
	img, err := tcfile.NewImage(entry.ID, fsInfo, entry.Comment())
	if err != nil {
		return nil, copier.NewFileError("failed to create image", entry, err)
	}
	if err := copier.ConvertGIFToPNG(cpng, bytes.NewReader(p.Data)); err != nil {
		return nil, copier.NewFileError("failed to convert GIF to PNG", entry, err)
	}
	return &payload{
		ID:    p.ID,
		Image: img,
		Data:  cpng.Bytes(),
	}, nil
}
