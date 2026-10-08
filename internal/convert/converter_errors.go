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
	Record int    // Record is the record number (1-indexed, 0 if not applicable).
	Line   int    // Line is the file line number (1-indexed).
	Column string // Column is the column name (empty if not applicable).
	Err    error  // Err is the wrapped error (e.g., ColumnError).
}

// Error returns a formatted string representation of the RowError,
// including record and line numbers, column (if applicable), and the
// wrapped error.
func (e RowError) Error() string {
	loc := fmt.Sprintf("%d", e.Line)
	if e.Record > 0 {
		loc = fmt.Sprintf("record %d (line %d)", e.Record, e.Line)
	}
	if e.Column != "" {
		return fmt.Sprintf("%s:%s: %v", loc, e.Column, e.Err)
	}
	return fmt.Sprintf("%s: %v", loc, e.Err)
}

// Unwrap returns the underlying error for compatibility with error unwrapping.
func (e RowError) Unwrap() error {
	return e.Err
}

// wrapColumnErrors wraps ColumnError values with RowError context to include
// the CSV record and line numbers. It handles both single errors and
// multi-error aggregations.
func wrapColumnErrors(record, line int, err error) error {
	switch merr := err.(type) {
	case nil:
		return nil
	case ColumnError:
		return RowError{Record: record, Line: line, Column: merr.Column, Err: merr}
	case interface{ Unwrap() []error }:
		errs := merr.Unwrap()
		newErrs := make([]error, 0, len(errs))
		for _, child := range errs {
			if colErr, ok := child.(ColumnError); ok {
				newErrs = append(newErrs, RowError{Record: record, Line: line, Column: colErr.Column, Err: colErr})
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
// The error message includes the text and the record and line numbers.
func newRowError(text string, record, line int, err error) error {
	return fmt.Errorf("%s: %w", text, RowError{Record: record, Line: line, Err: err})
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
func newProgressReport(records int) Report {
	return Report{
		Level:   slog.LevelInfo,
		Message: fmt.Sprintf("processing... (%d records)", records),
	}
}

// newCompletionReport creates an informational report about completion of CSV reading.
func newCompletionReport(records, lines int) Report {
	return Report{
		Level:   slog.LevelInfo,
		Message: fmt.Sprintf("finished reading input (%d records, %d lines)", records, lines),
	}
}
