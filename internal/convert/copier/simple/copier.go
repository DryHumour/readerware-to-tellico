package simple

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"os"

	"github.com/DryHumour/readerware-to-tellico/internal/convert/copier"
	"github.com/DryHumour/readerware-to-tellico/internal/images"
	"github.com/DryHumour/readerware-to-tellico/internal/tellico/tcfile"
)

// simpleCopier sequentially copies image files from a manifest into a Tellico collection.
// It reads and writes files one at a time without parallelism.
type simpleCopier struct {
	tcf    *tcfile.TCFile
	logger *slog.Logger
}

type Report = copier.Report

// New creates a new sequential copier for writing images to the given Tellico file.
func New(logger *slog.Logger, tcf *tcfile.TCFile) simpleCopier {
	return simpleCopier{
		logger: logger,
		tcf:    tcf,
	}
}

// CopyAll copies all image files from the manifest entries into the Tellico collection.
// It processes files sequentially, yielding reports and errors as they occur.
// Context cancellation will stop all processing and close the iterator.
func (c simpleCopier) CopyAll(ctx context.Context, entries iter.Seq[*images.ManifestEntry]) iter.Seq2[Report, error] {
	return func(yield func(Report, error) bool) {
		for entry := range entries {
			if err := context.Cause(ctx); err != nil {
				return
			}
			if !yield(Report{Level: slog.LevelInfo, Message: fmt.Sprintf("copying %s", entry.Path)}, nil) {
				return
			}
			var report Report
			err := c.Copy(ctx, entry)
			switch {
			case err == nil:
				if !entry.IsGIF {
					report = Report{
						Level:   slog.LevelInfo,
						Message: fmt.Sprintf("copied id=%q", entry.ID),
					}
				} else {
					report = Report{
						Level:   slog.LevelInfo,
						Message: fmt.Sprintf("converted GIF to PNG id=%q", entry.ID),
					}
				}
			case errors.Is(err, copier.ErrFileTooLarge), errors.Is(err, copier.ErrFormatMismatch):
				report = Report{
					Level:   slog.LevelWarn,
					Message: fmt.Sprintf("problem with image file: id=%q", entry.ID),
					Err:     err,
				}
				err = nil
			case entry.IsGIF:
				report = Report{
					Level:   slog.LevelInfo,
					Message: fmt.Sprintf("converted GIF to PNG: id=%q", entry.ID),
				}
			default:
			}
			if !yield(report, err) || err != nil {
				return
			}
		}
	}
}

// Copy reads a single image file and writes it to the Tellico collection.
// It validates the image format, checks for empty and oversized files, and reports errors.
// Context cancellation will stop the operation.
func (c simpleCopier) Copy(ctx context.Context, entry *images.ManifestEntry) error {
	var warnings []error

	// Open the file.
	if err := context.Cause(ctx); err != nil {
		return err
	}
	f, err := os.Open(entry.Path)
	if err != nil {
		return copier.NewFileError("failed to open file", entry, err)
	}
	defer f.Close()

	// Check if file is oversized.
	if err := context.Cause(ctx); err != nil {
		return err
	}
	stat, err := f.Stat()
	if err != nil {
		return copier.NewFileError("failed to stat file", entry, err)
	}
	if n := stat.Size(); n > copier.MaxReaderwareImageSize {
		warnings = append(warnings, copier.NewFileTooLargeError(entry, n))
	}

	// Read first few bytes to detect format and check for empty file.
	if err := context.Cause(ctx); err != nil {
		return err
	}
	buf := make([]byte, 512)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return copier.NewFileError("failed to read file", entry, err)
	}
	if err == io.EOF || n == 0 {
		return copier.NewFileEmptyError(entry)
	}
	buf = buf[:n]

	// Detect format.
	format, err := images.DetectFormat(buf)
	if err != nil {
		return copier.NewFileError("failed to detect supported image format", entry, err)
	}
	if (!entry.IsGIF && format != entry.Format) || (entry.IsGIF && format != "gif") {
		warnings = append(warnings, copier.NewFormatMismatchError(entry, format))
	}

	src := io.MultiReader(bytes.NewReader(buf), f)
	fsInfo := stat

	if entry.IsGIF {
		cpng := copier.NewConvertedPNG(fsInfo)
		fsInfo, _ = cpng.Stat() // cannot fail
		if err := copier.ConvertGIFToPNG(cpng, src); err != nil {
			return copier.NewFileError("failed to convert GIF to PNG", entry, err)
		}
		src = cpng
	}

	// Create the image metadata.
	if err := context.Cause(ctx); err != nil {
		return err
	}
	img, err := tcfile.NewImage(entry.ID, fsInfo, entry.Comment())
	if err != nil {
		return fmt.Errorf("failed to create image header: id=%q: %w", entry.ID, err)
	}

	// Create the image writer.
	if err := context.Cause(ctx); err != nil {
		return err
	}
	dst, err := c.tcf.Image(img)
	if err != nil {
		return fmt.Errorf("failed to get image writer: id=%q: %w", entry.ID, err)
	}

	// Copy the image data.
	if err := context.Cause(ctx); err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy image data: id=%q: %w", entry.ID, err)
	}

	return copier.MergeErrors(warnings...)
}
