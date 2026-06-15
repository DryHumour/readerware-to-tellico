package convert

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

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
	// ErrEmptyInputFile is returned when the input CSV contains no rows at all.
	ErrEmptyInputFile = fmt.Errorf("input file is empty")
)

// Converter handles the conversion of Readerware CSV data to Tellico format.
// It uses a streaming architecture to process large datasets without buffering
// all entries in memory.
type Converter struct {
	cfg        config.Config
	policy     collection.Policy
	logger     *slog.Logger
	httpClient HTTPClient
	images     *images.Index
	tcf        *tcfile.TCFile
	done       bool
}

// Option configures a Converter.
type Option func(*Converter)

// WithLogger sets the logger used during conversion.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Converter) { c.logger = logger }
}

// WithHTTPClient sets the HTTP client used for outbound requests (e.g. fetching ISBN range data).
func WithHTTPClient(h HTTPClient) Option {
	return func(c *Converter) { c.httpClient = h }
}

// HTTPClient is the interface used for outbound HTTP requests (e.g. ISBN range data).
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Report is an alias for copier.Report for convenience within the convert package.
type Report = copier.Report

// NewConverter creates a new Converter with the given configuration and collection policy.
// No I/O is performed; Run() opens and closes all files.
func NewConverter(cfg config.Config, policy collection.Policy, opt ...Option) *Converter {
	c := &Converter{cfg: cfg, policy: policy}
	for _, o := range opt {
		o(c)
	}
	if c.logger == nil {
		c.logger = slog.Default()
	}
	c.logger = c.logger.With("kind", policy.Info().Kind())
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}
	return c
}

// Run returns an iterator that performs the full conversion process.
// It opens the input and output files, indexes images, converts CSV entries to
// Tellico XML, copies image binaries into the TC archive, and closes all files
// before the iterator is exhausted. Non-fatal errors are yielded as reports while
// fatal errors are yielded as the error value.
// It can only be executed once per Converter instance; subsequent calls return immediately.
func (c *Converter) Run(ctx context.Context) iter.Seq2[Report, error] {
	return func(yield func(Report, error) bool) {
		if c.done {
			return
		}
		c.done = true // can only run once

		logger := c.logger

		// Open the input file.
		reader, err := newBufferedFile(c.cfg.InputFile)
		if err != nil {
			yield(Report{}, fmt.Errorf("failed to open input file: %w", err))
			return
		}
		defer reader.Close()

		// Ensure the output directory exists.
		if dir := filepath.Dir(c.cfg.OutputFile); dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				yield(Report{}, fmt.Errorf("failed to create output directory: %w", err))
				return
			}
		}

		// Open the output file.
		writer, err := os.Create(c.cfg.OutputFile)
		if err != nil {
			yield(Report{}, fmt.Errorf("failed to create output file: %w", err))
			return
		}
		defer writer.Close()

		// Apply extracted images dir fallback to image dirs.
		imageDirs := c.cfg.ImagesDirs.DefaultToExtracted(c.cfg.ExtractedImagesDir)

		// Index any images in the provided filesystems.
		c.images, err = images.BuildIndex(ctx, imageDirs)
		if err != nil {
			yield(Report{}, fmt.Errorf("failed to build image index: %w", err))
			return
		}

		// Create TC file early for streaming output.
		// Streaming architecture allows processing large datasets without buffering all entries in memory.
		// Entries are written to the TCFile as they are processed, rather than collecting all XML first.
		c.tcf = tcfile.New(writer)
		defer func() {
			// Check if already closed on success path ("archive/zip".Writer.Close is not idempotent).
			if c.tcf == nil {
				return
			}
			// Best effort (i.e. unreported) close on early exit / failure path.
			if err := c.tcf.Close(); err != nil {
				// may be too late to yield, so just slog
				logger.WarnContext(ctx, "failed to close TC file", "error", err)
			}
		}()

		for report, err := range c.convertAllEntries(ctx, reader) {
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
