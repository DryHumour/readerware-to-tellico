package convert

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"text/template"
)

// convertAllEntries returns an iterator that yields reports for each converted CSV entry.
// It reads and processes all CSV entries, applying cleaning templates and writing the
// results to the Tellico XML file. It uses a streaming approach to handle large datasets
// without buffering all entries in memory.
func (c *Converter) convertAllEntries(ctx context.Context) iter.Seq2[Report, error] {
	return func(yield func(Report, error) bool) {
		logger := c.cfg.Logger

		info := c.policy.Info()
		names := info.TemplateNames()

		//
		// Load templates.
		//

		tmpl, err := Base(template.FuncMap{
			"trace": func(msg string, args ...any) string { logger.Debug(msg, args...); return "" },
		})
		if err != nil {
			yield(Report{}, fmt.Errorf("failed to load base templates: %w", err))
			return
		}

		for _, dir := range c.cfg.TemplateDirs {
			fsys := os.DirFS(dir)
			tmpl, err = LoadAll(tmpl, fsys)
			if err != nil {
				yield(Report{}, fmt.Errorf("failed to load user templates from %s: %w", dir, err))
				return
			}
		}

		//
		// Execute the collection config template to build name options and genre blocklist.
		//

		// Fill in the collection configuration using the config template itself.
		if err := tmpl.ExecuteTemplate(io.Discard, names.Config, info.Data()); err != nil {
			yield(Report{}, fmt.Errorf("failed to execute config template %q: %w", names.Config, err))
			return
		}

		csvReader := csv.NewReader(c.cfg.Reader)
		csvReader.LazyQuotes = true

		lineNumber := 1

		//
		// Read CSV headers.
		//

		headers, err := csvReader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				yield(Report{}, ErrEmptyInputFile)
				return
			}
			yield(Report{}, fmt.Errorf("failed to read CSV headers: %w", err))
			return
		}

		if !yield(newHeaderReport(headers), nil) {
			return
		}

		if err := c.policy.ConfigureHeaders(headers, !c.images.IsEmpty()); err != nil {
			yield(Report{}, fmt.Errorf("invalid headers: %w", RowError{Line: lineNumber, Err: err}))
			return
		}

		//
		// Create tellico.xml file in TC file.
		//

		xmlFile, err := c.tcf.Collection()
		if err != nil {
			yield(Report{}, fmt.Errorf("failed to create Tellico collection file: %w", err))
			return
		}

		//
		// Write header template.
		//

		if err := executeTemplate(ctx, xmlFile, tmpl, names.Header, nil); err != nil {
			yield(Report{}, fmt.Errorf("failed to render Tellico XML header: %w", err))
			return
		}

		//
		// Process the CSV.
		//

		cleaner := cleaner{
			tmpl:   tmpl,
			colCfg: info.Columns(),
			logger: logger,
		}

		for {
			if err := context.Cause(ctx); err != nil {
				yield(Report{}, err)
				return
			}

			//
			// Report progress.
			//

			lineNumber++
			if lineNumber%progressStep == 0 {
				if !yield(newProgressReport(lineNumber), nil) {
					return
				}
			}

			//
			// Read a row from the CSV.
			//

			record, err := csvReader.Read()
			if err != nil {
				if errors.Is(err, io.EOF) {
					if !yield(newCompletionReport(lineNumber), nil) {
						return
					}
					break // finished processing records
				}
				if !yield(Report{}, newRowError("failed to read CSV row", lineNumber, err)) {
					return
				}
				continue // Keep going, best effort; process as many rows as possible for reporting
			}

			//
			// Clean the columns.
			//

			row := make(map[string]string, len(headers))
			for i, h := range headers {
				if i >= len(record) {
					break
				}
				row[h] = record[i]
			}
			clean, err := cleaner.CleanRow(ctx, row)
			if err != nil {
				if !yield(Report{}, wrapColumnErrors(lineNumber, err)) {
					return
				}
				continue // Keep going, best effort; process as many entries as possible for reporting
			}

			//
			// Build entry data (entry satisfies AggregationView for template use).
			//

			data, err := c.policy.NewEntry(clean, c.rowImages(clean["ROWKEY"]))
			if err != nil {
				if !yield(Report{}, newRowError("failed to create entry", lineNumber, err)) {
					return
				}
				continue // Keep going, best effort; process as many entries as possible for reporting
			}

			if err := executeTemplate(ctx, xmlFile, tmpl, names.Entry, data); err != nil {
				if !yield(Report{}, newRowError("failed to render Tellico XML entry", lineNumber, err)) {
					return
				}
				continue // Keep going, best effort; process as many entries as possible for reporting
			}

		}

		//
		// Write footer template (includes <images> section).
		//

		data := c.buildFooterData(ctx)
		if err := executeTemplate(ctx, xmlFile, tmpl, names.Footer, data); err != nil {
			yield(Report{}, fmt.Errorf("failed to render Tellico XML footer: %w", err))
			return
		}
	}
}

// executeTemplate checks context cancellation and executes a template.
// It returns an error if context is cancelled or template execution fails.
func executeTemplate(ctx context.Context, w io.Writer, tmpl *template.Template, name string, data any) error {
	if err := context.Cause(ctx); err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, name, data)
}
