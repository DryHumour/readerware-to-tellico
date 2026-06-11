// Package config handles CLI configuration and file I/O setup.
// It parses command-line arguments into a DTO, opens input/output files, and
// prepares image directory filesystems.
package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"log/slog"
)

// Config holds the resources and configuration for a single conversion run.
// The caller must call Close() to release the Reader and Writer when done.
// Close is idempotent and safe to call multiple times.
type Config struct {
	Reader       io.ReadCloser  // Reader is the input CSV reader.
	Writer       io.WriteCloser // Writer is the output TC file writer.
	ImagesDirs   Directories    // ImagesDirs are the image directories to search for images.
	TemplateDirs []string       // TemplateDirs are the user-provided template directories.
	Concurrency  int            // Concurrency is the number of parallel readers for image copying
	Logger       *slog.Logger   // Logger is the logger for conversion output
	closed       bool           // closed tracks whether Close has been called
}

// New creates a Config from a DTO by opening files.
// It opens the input CSV reader, creates the output file writer, and prepares
// image directory filesystems. The caller must call Close() on the returned
// Config to release resources.
func New(dto DTO, logger *slog.Logger) (cfg *Config, err error) {
	// Open the input file for reading.
	reader, err := newReadCloser(dto.InputFile)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			reader.Close()
		}
	}()

	// Ensure the output directory exists.
	if dir := filepath.Dir(dto.OutputFile); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Open the output file for writing.
	writer, err := os.Create(dto.OutputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		if err != nil {
			writer.Close()
		}
	}()

	// Return the configuration with the opened reader, images filesystem, and writer.
	return &Config{
		Reader:       reader,
		Writer:       writer,
		ImagesDirs:   dto.ImagesDirs.DefaultToExtracted(dto.ExtractedImagesDir),
		TemplateDirs: dto.TemplateDirs,
		Logger:       logger,
		Concurrency:  dto.Concurrency,
	}, nil
}

// Close releases the resources held by the Config (Reader and Writer).
// It aggregates errors from both Close calls. Safe to call multiple times
// - subsequent calls after the first will return nil.
func (c *Config) Close() error {
	if c.closed {
		return nil
	}

	var errs []error
	if err := c.Reader.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close reader: %w", err))
	}
	if err := c.Writer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close writer: %w", err))
	}

	c.closed = true
	return errors.Join(errs...)
}
