package convert

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"text/template"

	"github.com/DryHumour/readerware-to-tellico/internal/tellico/collection"
)

// cleaner applies per-column cleaning templates to CSV rows.
//
// It looks up and executes templates named "clean.<Column>" and falls back to
// "clean.DEFAULT" when no column-specific template exists.
//
// The template set is expected to be created by Base() (and optionally extended
// with user overrides).
type cleaner struct {
	tmpl   *template.Template
	colCfg collection.ColumnConfig
	logger *slog.Logger
}

// ColumnError wraps an error produced while transforming a specific CSV column.
//
// The error is intended to be used for row-level error aggregation while still
// preserving which column triggered the failure.
type ColumnError struct {
	Column string
	Err    error
}

func (e ColumnError) Error() string {
	return fmt.Sprintf("column %q: %v", e.Column, e.Err)
}

func (e ColumnError) Unwrap() error {
	return e.Err
}

// CellData is the context object passed to every cleaning template.
// In a template, fields are accessed as .Column, .Value, etc.
type CellData struct {
	// Column is the Readerware column name for this cell (e.g. "AUTHOR", "TITLE").
	// Use this when a template needs to behave differently depending on which column it is cleaning.
	Column string
	// ColumnRole is the credit role associated with this column (e.g. "Authors" for "AUTHOR").
	// This is informational — the cleaning phase does not use it, but a custom template may.
	ColumnRole string
	// Value is the raw text of this cell as read from the CSV export.
	// This is what most cleaning templates transform and return.
	Value string
	// Row is the complete set of raw cell values for the current entry, keyed by column name.
	// Use this when a cleaning template needs to look at another column to decide how to clean this one.
	Row map[string]string
}

// CleanRow applies CleanCell to every cell in the row in the order provided by headers
// and returns a new map containing the cleaned values.
//
// Any errors encountered are aggregated and wrapped in ColumnError values.
func (c cleaner) CleanRow(ctx context.Context, raw map[string]string) (clean map[string]string, err error) {
	var errs []error
	clean = make(map[string]string, len(raw))

	for _, col := range c.colCfg.Headers {
		val, ok := raw[col]
		if !ok {
			continue
		}
		out, err := c.CleanCell(ctx, CellData{
			Column:     col,
			ColumnRole: c.colCfg.ColumnRole(col),
			Value:      val,
			Row:        raw,
		})
		if err != nil {
			errs = append(errs, ColumnError{Column: col, Err: err})
		}
		clean[col] = out
	}
	return clean, errors.Join(errs...)
}

// CleanCell applies a single cleaning template to a cell and returns the
// cleaned string.
func (c cleaner) CleanCell(ctx context.Context, cell CellData) (string, error) {
	name := "clean." + cell.Column
	if cell.Column == "" || c.tmpl.Lookup(name) == nil {
		name = "clean.default"
	}
	if err := context.Cause(ctx); err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := c.tmpl.ExecuteTemplate(&buf, name, cell); err != nil {
		return "", err
	}
	clean := buf.String()
	c.logger.DebugContext(
		ctx,
		"Cleaning cell",
		"column", cell.Column,
		"role", cell.ColumnRole,
		"template", name,
		"input", cell.Value,
		"output", clean,
	)
	return clean, nil
}
