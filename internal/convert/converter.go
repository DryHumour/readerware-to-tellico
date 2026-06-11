package convert

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log/slog"

	"github.com/DryHumour/readerware-to-tellico/internal/config"
	"github.com/DryHumour/readerware-to-tellico/internal/convert/copier"
	"github.com/DryHumour/readerware-to-tellico/internal/images"
	"github.com/DryHumour/readerware-to-tellico/internal/tellico/collection"
	"github.com/DryHumour/readerware-to-tellico/internal/tellico/tcfile"
)

const (
	// progressStep is the number of rows to process before reporting progress.
	progressStep = 100
)

var (
	// ErrInvalidSpec is returned when the conversion specification is invalid.
	ErrInvalidSpec = errors.New("invalid conversion specification")
	// ErrEmptyInputFile is returned when the input CSV contains no rows at all.
	ErrEmptyInputFile = errors.New("input file is empty")
)

// Converter handles the conversion of Readerware CSV data to Tellico format.
// It uses a streaming architecture to process large datasets without buffering
// all entries in memory.
type Converter struct {
	cfg    *config.Config
	policy collection.Policy
	images *images.Index
	tcf    *tcfile.TCFile
	done   bool
}

// Report is an alias for copier.Report for convenience within the convert package.
type Report = copier.Report

// Copier defines the interface for copying image files into the TC archive.
type Copier interface {
	CopyAll(ctx context.Context, entries iter.Seq[*images.ManifestEntry]) iter.Seq2[Report, error]
}

// NewConverter creates a new Converter with the given configuration and collection policy.
// It validates the configuration and sets up the logger with appropriate context.
func NewConverter(ctx context.Context, cfg *config.Config, policy collection.Policy) (*Converter, error) {
	if policy == nil {
		return nil, fmt.Errorf("%w:  missing policy", ErrInvalidSpec)
	}
	if cfg.Reader == nil {
		return nil, fmt.Errorf("%w:  missing reader", ErrInvalidSpec)
	}
	if cfg.Writer == nil {
		return nil, fmt.Errorf("%w:  missing writer", ErrInvalidSpec)
	}
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	cfg.Logger = logger.With("kind", policy.Info().Kind())
	return &Converter{cfg: cfg, policy: policy}, nil
}

// Run returns an iterator that performs the full conversion process.
// It indexes images, converts CSV entries to Tellico XML, copies image binaries into
// the TC archive, and yields reports for each step. Non-fatal errors are yielded
// as reports while fatal errors are yielded as the error value.
// It can only be executed once per Converter instance; subsequent calls return immediately.
func (c *Converter) Run(ctx context.Context) iter.Seq2[Report, error] {
	return func(yield func(Report, error) bool) {
		if c.done {
			return
		}
		c.done = true // can only run once

		var err error

		// Index any images in the provided filesystems
		c.images, err = images.BuildIndex(ctx, c.cfg.ImagesDirs)
		if err != nil {
			yield(Report{}, fmt.Errorf("failed to build image index: %w", err))
			return
		}

		// Create TC file early for streaming output.
		// Streaming architecture allows processing large datasets without buffering all entries in memory.
		// Entries are written to the TCFile as they are processed, rather than collecting all XML first.
		c.tcf = tcfile.New(c.cfg.Writer)
		defer func() {
			// Check if already closed on success path ("archive/zip".Writer.Close is not idempotent).
			if c.tcf == nil {
				return
			}
			// Best effort (i.e. unreported) close on early exit / failure path.
			if err := c.tcf.Close(); err != nil {
				// may be too late to yield, so just slog
				c.cfg.Logger.WarnContext(ctx, "failed to close TC file", "error", err)
			}
		}()

		for report, err := range c.convertAllEntries(ctx) {
			if !yield(report, err) || err != nil {
				return
			}
		}

		for report, err := range c.copyAllImages(ctx) {
			if !yield(report, err) || err != nil {
				return
			}
		}

		c.tcf, err = nil, c.tcf.Close()
		if err != nil {
			yield(Report{}, fmt.Errorf("failed to close TC file: %w", err))
			return
		}
	}
}
