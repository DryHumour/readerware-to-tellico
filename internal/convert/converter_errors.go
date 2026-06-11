package convert

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
)

// RowError wraps an error with CSV file location context.
type RowError struct {
	Line   int    // Line is the row number (1-indexed).
	Column string // Column is the column name (empty if not applicable).
	Err    error  // Err is the wrapped error (e.g., ColumnError).
}

// Error returns a formatted string representation of the RowError,
// including line number, column (if applicable), and the wrapped error.
func (e RowError) Error() string {
	if e.Column != "" {
		return fmt.Sprintf("%d:%s: %v", e.Line, e.Column, e.Err)
	}
	return fmt.Sprintf("%d: %v", e.Line, e.Err)
}

// Unwrap returns the underlying error for compatibility with error unwrapping.
func (e RowError) Unwrap() error {
	return e.Err
}

// wrapColumnErrors wraps ColumnError values with RowError context to include
// the CSV line number. It handles both single errors and multi-error aggregations.
func wrapColumnErrors(lineNumber int, err error) error {
	switch merr := err.(type) {
	case nil:
		return nil
	case ColumnError:
		return RowError{Line: lineNumber, Column: merr.Column, Err: merr}
	case interface{ Unwrap() []error }:
		errs := merr.Unwrap()
		newErrs := make([]error, 0, len(errs))
		for _, child := range errs {
			if colErr, ok := child.(ColumnError); ok {
				newErrs = append(newErrs, RowError{Line: lineNumber, Column: colErr.Column, Err: colErr})
			} else {
				newErrs = append(newErrs, child)
			}
		}
		return errors.Join(newErrs...)
	default:
		return err
	}
}

// newRowError creates a new error wrapping the provided error with RowError context.
// The error message includes the text and the line number derived from rowCount.
func newRowError(text string, lineNumber int, err error) error {
	return fmt.Errorf("%s: %w", text, RowError{Line: lineNumber, Err: err})
}

// newHeaderReport creates an informational report about parsed CSV headers.
func newHeaderReport(headers []string) Report {
	return Report{
		Level: slog.LevelInfo,
		Message: fmt.Sprintf(
			"parsed CSV: %d columns: %s",
			len(headers),
			strings.Join(slices.Sorted(slices.Values(headers)), " "),
		),
	}
}

// newProgressReport creates an informational report about conversion progress.
func newProgressReport(lineNumber int) Report {
	return Report{
		Level:   slog.LevelInfo,
		Message: fmt.Sprintf("processing... (%d lines)", lineNumber),
	}
}

// newCompletionReport creates an informational report about completion of CSV reading.
func newCompletionReport(lineNumber int) Report {
	return Report{
		Level:   slog.LevelInfo,
		Message: fmt.Sprintf("finished reading input (%d lines)", lineNumber),
	}
}
