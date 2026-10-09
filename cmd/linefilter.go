package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/DryHumour/readerware-to-tellico/internal/httpclient"
	"github.com/DryHumour/readerware-to-tellico/internal/strutil"
	"github.com/DryHumour/readerware-to-tellico/isbn"
	"github.com/spf13/cobra"
)

// lineFilterOptions configures the output side of runLineFilter.
type lineFilterOptions struct {
	raw        bool             // print bare text instead of JSON string literals
	hyphenate  bool             // print hyphenated ISBNs when the range data allows it
	hyphenator *isbn.Hyphenator // required; used for hyphenation output only
}

// runLineFilter reads inputs (stdin lines when args is empty, otherwise one
// per arg), applies convert to each, and writes exactly one output line per
// input.  Blank lines and lines beginning with '#' pass through verbatim
// (stdin only).  A line beginning with '"' is treated as a JSON string literal
// and decoded before conversion.  On conversion failure the original input is
// echoed (JSON-quoted unless raw), an error is logged identifying the line or
// argument, and processing continues; the function returns an error at the
// end if any input failed ("%d of %d inputs failed to convert").
func runLineFilter(cmd *cobra.Command, args []string, convert func(string) (isbn.ISBN, error), opts lineFilterOptions) error {
	ctx := cmd.Context()
	out := cmd.OutOrStdout()
	logger := slog.Default()

	var total, failed int
	process := func(label string, n int, raw string) error {
		text := raw
		if u, ok := strutil.UnquoteJSON(text); ok {
			text = u
		}
		i, err := convert(text)
		var s string
		if err != nil {
			failed++
			logger.ErrorContext(ctx, "conversion failed", label, n, "input", raw, "error", err)
			s = raw
		} else {
			s = i.String()
			if opts.hyphenate {
				if h, herr := opts.hyphenator.Hyphenate(i); herr == nil {
					s = h
				} else {
					logger.WarnContext(ctx, "hyphenation unavailable, printing unhyphenated", label, n, "isbn", i.String(), "error", herr)
				}
			}
		}
		if !opts.raw {
			s = strutil.QuoteJSON(s)
		}
		return writeLine(out, s)
	}

	if len(args) > 0 {
		for i, arg := range args {
			if err := context.Cause(ctx); err != nil {
				return err
			}
			total++
			if err := process("arg", i+1, arg); err != nil {
				return err
			}
		}
	} else {
		scanner := bufio.NewScanner(cmd.InOrStdin())
		for n := 1; scanner.Scan(); n++ {
			if err := context.Cause(ctx); err != nil {
				return err
			}
			raw := scanner.Text()
			trimmed := strings.TrimSpace(raw)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				if err := writeLine(out, raw); err != nil {
					return err
				}
				continue
			}
			total++
			if err := process("line", n, raw); err != nil {
				return err
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d of %d inputs failed to convert", failed, total)
	}
	return nil
}

// loadHyphenator fetches current ISBN range data (falling back on the built-in
// data) using the shared caching HTTP client.
func loadHyphenator(cmd *cobra.Command) (*isbn.Hyphenator, error) {
	h, err := isbn.LoadHyphenator(cmd.Context(), httpclient.New(slog.Default()))
	if err != nil {
		return nil, fmt.Errorf("failed to load ISBN range data: %w", err)
	}
	return h, nil
}

// lineFilterOptionsFromFlags reads the --raw and --hyphenate flags shared by
// the line-filter commands.
func lineFilterOptionsFromFlags(cmd *cobra.Command, h *isbn.Hyphenator) (lineFilterOptions, error) {
	raw, err := cmd.Flags().GetBool("raw")
	if err != nil {
		return lineFilterOptions{}, fmt.Errorf("failed to read --raw flag: %w", err)
	}
	hyphenate, err := cmd.Flags().GetBool("hyphenate")
	if err != nil {
		return lineFilterOptions{}, fmt.Errorf("failed to read --hyphenate flag: %w", err)
	}
	return lineFilterOptions{raw: raw, hyphenate: hyphenate, hyphenator: h}, nil
}
