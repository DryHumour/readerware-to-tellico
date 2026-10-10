package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/DryHumour/readerware-to-tellico/isbn"
)

// isbnCmd represents the isbn command.
var isbnCmd = &cobra.Command{
	Use:   "isbn",
	Short: "Manipulate ISBNs",
	Long:  `Manipulate ISBN values: normalise, convert, and hyphenate.`,
}

// isbnTo13Cmd represents the isbn to13 command.
var isbnTo13Cmd = &cobra.Command{
	Use:     "to13 [isbn...]",
	Aliases: []string{"13"},
	Short:   "Convert ISBNs to ISBN-13",
	Long: `Convert ISBNs to ISBN-13.

With no arguments, reads ISBN values from stdin, one per line.  With one or
more arguments, each argument is treated as a value to convert.  Accepted
inputs are ISBN-10 or ISBN-13 values, with optional hyphens and spaces; a
line that begins with a double quote is decoded as a JSON string literal
before conversion.

Each input produces exactly one output line: a JSON string literal holding
the ISBN-13 ("\"9780306406157\""), or bare digits with --raw.  With
--hyphenate the ISBN-13 prints in hyphenated form where the ISBN range data
allows it.

Check-digit failures are failures: a structurally invalid ISBN is never
repaired.  ISBN-13 prefixes not found in the ISBN range data (e.g. non-book
EAN-13 codes) are also failures.  An ISBN in a known registration group but
an unassigned publisher block is accepted with a warning; --strict
additionally fails any output whose check digit is invalid or that does not
resolve in the ISBN range data (e.g. an unassigned publisher block).

A value that fails conversion prints unchanged and an error is logged to
stderr identifying the input line or argument.  If any value fails, the
command exits non-zero after all inputs have been processed.`,
	Args:         cobra.ArbitraryArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		h, err := loadHyphenator(cmd)
		if err != nil {
			return err
		}
		return runISBNTo13(cmd, args, h)
	},
}

func init() {
	rootCmd.AddCommand(isbnCmd)
	isbnCmd.AddCommand(isbnTo13Cmd)
	isbnTo13Cmd.Flags().BoolP("raw", "r", false, "Output bare ISBNs instead of JSON string literals")
	isbnTo13Cmd.Flags().BoolP("hyphenate", "H", false, "Output hyphenated ISBNs where the range data allows")
	isbnTo13Cmd.Flags().Bool("strict", false, "Fail outputs that are not valid, range-resolvable ISBNs")
}

func runISBNTo13(cmd *cobra.Command, args []string, h *isbn.Hyphenator) error {
	opts, err := lineFilterOptionsFromFlags(cmd, h)
	if err != nil {
		return err
	}

	return runLineFilter(cmd, args, func(s string) (isbn.ISBN, error) {
		i, err := isbn.New(s)
		if err != nil {
			return isbn.ISBN{}, fmt.Errorf("invalid ISBN: %w", err)
		}
		i = i.To13()
		_, err = h.Hyphenate(i)
		switch {
		case err == nil, errors.Is(err, isbn.ErrPublisherRangeNotFound), errors.Is(err, isbn.ErrInvalidPublisherLength):
			return i, nil
		case errors.Is(err, isbn.ErrRegistrationGroupNotFound):
			return isbn.ISBN{}, fmt.Errorf("not a book ISBN prefix: %w", err)
		default:
			return isbn.ISBN{}, err
		}
	}, opts)
}
