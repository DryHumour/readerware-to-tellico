package copier

import (
	"errors"
	"fmt"
	"io"

	"github.com/DryHumour/readerware-to-tellico/internal/images"
)

const (
	// MaxReaderwareImageSize is the maximum size of a Readerware large image (256KiB).
	// Thumbnails are 64KiB. This value is used as the buffer size for image data transfer.
	MaxReaderwareImageSize = 256 * 1024
)

var (
	// ErrFormatMismatch is returned when the detected image format does not match the expected format.
	ErrFormatMismatch = errors.New("format mismatch")
	// ErrFileTooLarge is returned when a file exceeds the expected buffer size.
	ErrFileTooLarge = errors.New("file larger than expected")
	// ErrFileEmpty is returned when a file is empty.
	ErrFileEmpty = errors.New("file is empty")
)

// NewFileEmptyError creates an error for an empty file, joining the file path error with io.EOF.
func NewFileEmptyError(entry *images.ManifestEntry) error {
	return errors.Join(fmt.Errorf("%s: %w", entry.Path, ErrFileEmpty), io.EOF)
}

// NewFileError creates a file-related error with context text and the file path.
func NewFileError(text string, entry *images.ManifestEntry, err error) error {
	return fmt.Errorf("%s: %s: %w", text, entry.Path, err)
}

// NewFileTooLargeError creates an error for a file that exceeds the expected buffer size.
func NewFileTooLargeError(entry *images.ManifestEntry, size int64) error {
	return fmt.Errorf("%s: size %d bytes: %w", entry.Path, size, ErrFileTooLarge)
}

// NewFormatMismatchError creates an error when the detected format does not match the expected format.
func NewFormatMismatchError(entry *images.ManifestEntry, format string) error {
	return fmt.Errorf("%s: expected %s, got %s: %w", entry.Path, entry.Format, format, ErrFormatMismatch)
}

// MergeErrors combines multiple errors into a single error.
// If there is only one non-nil error, it is returned directly.
// If there are multiple non-nil errors, they are joined using errors.Join.
// If all errors are nil, nil is returned.
func MergeErrors(errs ...error) error {
	var singleErr error
	for _, err := range errs {
		if err == nil {
			continue
		}
		if singleErr != nil {
			return errors.Join(errs...)
		}
		singleErr = err
	}
	return singleErr
}
