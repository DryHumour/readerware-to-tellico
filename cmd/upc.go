package cmd

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"reflect"
	"slices"
	"strings"
	"unicode"

	"github.com/DryHumour/readerware-to-tellico/isbn"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

// upcCmd represents the upc command
var upcCmd = &cobra.Command{
	Use:   "upc",
	Short: "Manage UPC codes",
	Long:  `Manage UPC codes for books and other items.`,
}

// upcListCmd represents the UPC prefix table list command
var upcListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the UPC to ISBN prefix table",
	Long: `Print the UPC company prefix / ISBN publisher prefix table, sorted by
UPC key.  By default the effective table prints as a complete "upc" YAML config
fragment whose "table" list can be pasted directly into a config file; --raw
prints "key value" lines instead.  The "default" sub-command prints the
compiled-in table, ignoring any upc.table config.`,
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runUPCList(cmd)
	},
}

// upcListDefaultCmd represents the UPC prefix table list default command
var upcListDefaultCmd = &cobra.Command{
	Use:   "default",
	Short: "List the built-in UPC to ISBN prefix table",
	Long: `Print the compiled-in UPC/ISBN prefix table, ignoring any upc.table
config.  Useful for comparing against the effective table or inspecting the
built-in pairs when the config table is malformed.`,
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runUPCListDefault(cmd)
	},
}

func init() {
	rootCmd.AddCommand(upcCmd)
	upcCmd.AddCommand(upcISBNCmd)
	upcCmd.AddCommand(upcListCmd)
	upcListCmd.AddCommand(upcListDefaultCmd)
	upcListCmd.PersistentFlags().BoolP("raw", "r", false, "Output unquoted \"key value\" pairs instead of a YAML config fragment")
	upcISBNCmd.Flags().BoolP("raw", "r", false, "Output bare values instead of JSON string literals")
	upcISBNCmd.Flags().BoolP("hyphenate", "H", false, "Output hyphenated ISBNs where the range data allows")
}

// upcISBNCmd represents the UPC to ISBN convert command
var upcISBNCmd = &cobra.Command{
	Use:   "isbn [upc...]",
	Short: "Convert UPC codes to ISBNs",
	Long: `Convert UPC codes to ISBNs.

With no arguments, reads UPC/EAN/ISBN values from stdin, one per line.  With
one or more arguments, each argument is treated as a value to convert.
Accepted inputs are ISBN-10, ISBN-13 (which may carry a 5-digit add-on), and
UPC-12 with a 5-digit add-on; punctuation and formatting characters are
ignored.  An EAN-13-encoded UPC-A (a leading zero followed by the UPC-12,
optionally with a 5-digit add-on) is also accepted: the leading zero is
stripped and the value converts like the equivalent UPC-12.  A UPC-12
without its add-on cannot be converted: the add-on carries the title digits
of the ISBN.

Each input produces exactly one output line as a JSON string literal (--raw
prints bare text instead): valid ISBNs print in normalized form and
resolvable UPCs print their ISBN-10.  With --hyphenate, output ISBNs print in
hyphenated form where the range data allows it.  An ISBN with an invalid
check digit passes through unchanged but logs a warning.  A value that
fails conversion prints unchanged and an error is logged to stderr
identifying the input line or argument.  If any value fails, the command
exits non-zero after all inputs have been processed.`,
	Args:         cobra.ArbitraryArgs,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		h, err := loadHyphenator(cmd)
		if err != nil {
			return err
		}
		return runUPCToISBN(cmd, args, h)
	},
}

// upcToISBNPrefix maps a 6-digit UPC company prefix to the corresponding
// ISBN-10 publisher prefix(es).  This table is treated as read-only.  Some UPC
// prefixes correspond to more than one publisher prefix (e.g. "076714"); the
// first usable candidate wins.
var upcToISBNPrefix = map[string][]string{
	"011271": {"0939001"},
	"014794": {"08041"},
	"016067": {"08092"},
	"018926": {"0445"},
	"027778": {"0449"},
	"030314": {"0385"},
	"034057": {"093187", "155560"},
	"037038": {"088266"},
	"037145": {"0812", "0765"},
	"042799": {"0785"},
	"043144": {"0688"},
	"044903": {"0312"},
	"045863": {"0517"},
	// 0064 = 0-06-4..., 0694 = 0-694-...; both registrants are HarperCollins.
	"046594": {"0064", "0694"},
	"047132": {"0152"},
	"050694": {"0345"},
	"051487": {"08167"},
	"051488": {"0140"},
	"060771": {"0002"},
	"065373": {"0373"},
	"070661": {"0376"},
	"070992": {"0523"},
	"070993": {"0446"},
	"070999": {"0345"},
	"071001": {"0380"},
	"071009": {"0440"},
	"071125": {"088677"},
	"071136": {"0451"},
	"071149": {"0451"},
	"071152": {"0515"},
	"071162": {"0451"},
	"071268": {"08217"},
	"071831": {"0425"},
	"071842": {"08439"},
	"072742": {"0441"},
	"076714": {"0671", "07434", "0743"},
	"076783": {"0553"},
	// 0449 = Fawcett, 0445 = Popular Library; both were Fawcett Publications
	// imprints during the mass-market UPC era.
	"076814": {"0449", "0445"},
	"077434": {"0743"},
	"078021": {"0872"},
	"079808": {"0394"},
	"090129": {"0679"},
	"099455": {"0061"},
	"099769": {"0451"},
	"612264": {"07868"},
	"645573": {"1595"},
	"777506": {"155166"},
}

func runUPCList(cmd *cobra.Command) error {
	table, err := effectiveUPCTable()
	if err != nil {
		return err
	}
	return printTableFromCmd(cmd, table)
}

func runUPCListDefault(cmd *cobra.Command) error {
	return printTableFromCmd(cmd, upcToISBNPrefix)
}

func printTableFromCmd(cmd *cobra.Command, table map[string][]string) error {
	raw, err := cmd.Flags().GetBool("raw")
	if err != nil {
		return fmt.Errorf("failed to read --raw flag: %w", err)
	}
	return printUPCTable(cmd.OutOrStdout(), table, raw)
}

// upcConfigDTO is the shape of the upc.* config namespace; viper populates it
// in effectiveUPCTable and it round-trips back to YAML in printUPCTable.
type upcConfigDTO struct {
	UPC struct {
		Table []string `mapstructure:"table" yaml:"table"`
	} `mapstructure:"upc" yaml:"upc"`
}

// printUPCTable writes the UPC prefix table to w, sorted by UPC key.  Unless
// raw is set, the table prints as a complete "upc.table" YAML fragment
// suitable for pasting into a config file; with raw it prints "key value"
// lines.  Multi-candidate keys repeat the key, printing one pair per
// candidate in consideration order.
func printUPCTable(w io.Writer, table map[string][]string, raw bool) error {
	keys := slices.Sorted(maps.Keys(table))
	if raw {
		for _, key := range keys {
			for _, pub := range table[key] {
				if err := writeLine(w, key+" "+pub); err != nil {
					return err
				}
			}
		}
		return nil
	}

	var dto upcConfigDTO
	for _, key := range keys {
		for _, pub := range table[key] {
			dto.UPC.Table = append(dto.UPC.Table, key, pub)
		}
	}
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	if err := enc.Encode(dto); err != nil {
		return fmt.Errorf("failed to encode UPC table as YAML: %w", err)
	}
	if err := enc.Close(); err != nil {
		return fmt.Errorf("failed to close YAML encoder: %w", err)
	}
	return nil
}

func runUPCToISBN(cmd *cobra.Command, args []string, h *isbn.Hyphenator) error {
	table, err := effectiveUPCTable()
	if err != nil {
		return err
	}
	opts, err := lineFilterOptionsFromFlags(cmd, h)
	if err != nil {
		return err
	}
	return runLineFilter(cmd, args, func(s string) (isbn.ISBN, error) {
		return upcToISBN(s, table, h)
	}, opts)
}

// upcToISBN converts a single UPC/EAN/ISBN string into its ISBN representation.
// It passes through valid ISBN-10 and ISBN-13 values (an ISBN-13 may carry a
// 5-digit add-on, which is discarded), converts UPC-12 values with an
// accompanying UPC-5 add-on to ISBN-10, and returns an error for unrecognised
// inputs.  An ISBN with a bad check digit passes through with a warning.
// An EAN-13-encoded UPC-A (leading zero, which can never be a Bookland
// ISBN-13) is re-dispatched after stripping the zero.
func upcToISBN(s string, table map[string][]string, h *isbn.Hyphenator) (isbn.ISBN, error) {
	digits := extractDigits(s)

	switch len(digits) {
	case 10:
		return tolerantISBN(digits)
	case 13, 18:
		if digits[0] == '0' {
			// A Bookland ISBN-13 always starts with 978/979, so a leading
			// zero marks an EAN-13-encoded UPC-A: strip it and re-dispatch
			// (13 digits → 12, 18 digits → 17).
			return upcToISBN(digits[1:], table, h)
		}
		// For 18 digits this is an ISBN-13 with a 5-digit price add-on:
		// validate and keep the ISBN-13, discard the add-on.
		i, err := tolerantISBN(digits[:13])
		if err != nil {
			return isbn.ISBN{}, err
		}
		if _, herr := h.Hyphenate(i); errors.Is(herr, isbn.ErrRegistrationGroupNotFound) {
			return isbn.ISBN{}, fmt.Errorf("unrecognised EAN-13 prefix %s: not a book ISBN: %w", digits[:3], herr)
		}
		return i, nil
	case 12:
		if !isDigits(digits) {
			return isbn.ISBN{}, fmt.Errorf("unexpected 'X' in UPC input %q", digits)
		}
		if err := validUPC12(digits); err != nil {
			return isbn.ISBN{}, err
		}
		return isbn.ISBN{}, errors.New("UPC-12 requires a 5-digit add-on")
	case 17:
		if !isDigits(digits) {
			return isbn.ISBN{}, fmt.Errorf("unexpected 'X' in UPC input %q", digits)
		}
		upc12 := digits[:12]
		upc5 := digits[12:]
		if err := validUPC12(upc12); err != nil {
			return isbn.ISBN{}, err
		}
		return isbn10FromUPC(upc12, upc5, table)
	default:
		return isbn.ISBN{}, fmt.Errorf("unrecognized input format: %d digits extracted (expected ISBN-10, ISBN-13, ISBN-13+5, or UPC-12+5)", len(digits))
	}
}

// tolerantISBN parses an ISBN-10 or ISBN-13, tolerating a bad check digit:
// the structurally valid ISBN is returned with a warning logged rather than
// an error.  Any other parse failure is fatal.
func tolerantISBN(digits string) (isbn.ISBN, error) {
	i, err := isbn.New(digits)
	if errors.Is(err, isbn.ErrInvalidCheckDigit) {
		slog.Default().Warn("passing through ISBN with invalid check digit", "isbn", i.String())
		return i, nil
	}
	if err != nil {
		return isbn.ISBN{}, fmt.Errorf("invalid ISBN: %w", err)
	}
	return i, nil
}

// extractDigits returns a string containing only the digit and 'X'/'x' runes
// found in s.  It discards whitespace, punctuation, formatting characters, and
// any other non-ISBN/UPC characters.
func extractDigits(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9', r == 'X', r == 'x':
			b.WriteRune(unicode.ToUpper(r))
		}
	}
	return b.String()
}

// validUPC12 reports whether s is a valid 12-digit UPC-A barcode, including
// its final modulo-10 check digit.
func validUPC12(s string) error {
	if len(s) != 12 {
		return fmt.Errorf("expected 12 UPC digits, got %d", len(s))
	}

	sum := 0
	for i := range 12 {
		digit := int(s[i] - '0')
		if i%2 == 0 {
			sum += digit * 3
		} else {
			sum += digit
		}
	}
	if sum%10 != 0 {
		return errors.New("invalid UPC-A check digit")
	}
	return nil
}

// isbn10FromUPC converts a validated 12-digit UPC-A and a 5-digit add-on into
// an ISBN-10 using the supplied publisher prefix table.  On mass-market books
// the add-on is overloaded to carry the trailing ISBN title digits rather than
// a price.  A UPC prefix may map to more than one publisher prefix; the first
// usable candidate wins.
func isbn10FromUPC(upc12, upc5 string, table map[string][]string) (isbn.ISBN, error) {
	if len(upc12) != 12 {
		return isbn.ISBN{}, fmt.Errorf("expected 12 UPC digits, got %d", len(upc12))
	}
	if len(upc5) != 5 {
		return isbn.ISBN{}, fmt.Errorf("expected 5 add-on digits, got %d", len(upc5))
	}

	pubs := table[upc12[:6]]
	if len(pubs) == 0 {
		return isbn.ISBN{}, fmt.Errorf("unknown UPC prefix %s (add a mapping via \"upc.table\" in the config file)", upc12[:6])
	}

	var errs []error
	for _, pub := range pubs {
		need := 9 - len(pub)
		if need < 0 || need > 5 {
			errs = append(errs, fmt.Errorf("incompatible ISBN publisher prefix %q for %s", pub, upc12[:6]))
			continue
		}
		body := pub + upc5[5-need:]
		isbn10, err := isbn.Make10(body)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid ISBN body %q for %s: %w", body, upc12[:6], err))
			continue
		}
		return isbn10, nil
	}

	return isbn.ISBN{}, fmt.Errorf("no usable ISBN publisher prefix for %s: %w", upc12[:6], errors.Join(errs...))
}

// effectiveUPCTable returns the UPC/ISBN prefix table used by the upc
// commands: pairs from the config file's upc.table take precedence over the
// built-in pairs, which remain as fallback candidates.
//
// Quote table elements in the config file: the decode hook rejects non-string
// scalars so that unquoted values like 070993 fail loudly instead of silently
// losing their leading zeros.
func effectiveUPCTable() (map[string][]string, error) {
	v := viper.GetViper()

	var dto upcConfigDTO

	// Viper's weak typing would silently stringify numeric scalars, losing
	// leading zeros; reject them while the raw type is still visible.
	err := v.Unmarshal(&dto, viper.DecodeHook(func(f, t reflect.Type, data any) (any, error) {
		if t.Kind() == reflect.String && f.Kind() != reflect.String {
			return nil, fmt.Errorf("unquoted scalar %v: \"upc.table\" elements must be quoted strings", data)
		}
		return data, nil
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal viper config: %w", err)
	}
	if len(dto.UPC.Table) == 0 {
		return upcToISBNPrefix, nil
	}

	userTable, err := upcPairsToTable(dto.UPC.Table)
	if err != nil {
		return nil, fmt.Errorf("invalid \"upc.table\": %w", err)
	}
	result := maps.Clone(upcToISBNPrefix)
	for key, pubs := range userTable {
		var merged []string
		for _, pub := range slices.Concat(pubs, result[key]) {
			if !slices.Contains(merged, pub) {
				merged = append(merged, pub)
			}
		}
		result[key] = merged
	}
	return result, nil
}

// upcPairsToTable converts an alternating key/value pair list into a map from
// UPC prefix to its ordered candidate publisher prefixes.  Every UPC key must
// be exactly six digits long and every ISBN prefix must be a digit string of
// four to nine digits (the 5-digit add-on supplies the remaining title digits).
func upcPairsToTable(pairs []string) (map[string][]string, error) {
	if len(pairs)%2 != 0 {
		return nil, fmt.Errorf("odd-length UPC/ISBN pair list (%d elements)", len(pairs))
	}
	var errs []error
	m := make(map[string][]string, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		key, val := pairs[i], pairs[i+1]
		valid := true
		switch {
		case len(key) != 6:
			errs = append(errs, fmt.Errorf("pair %d: UPC key %q has %d digits, expected 6", i/2, key, len(key)))
			valid = false
		case !isDigits(key):
			errs = append(errs, fmt.Errorf("pair %d: UPC key %q contains non-digit characters", i/2, key))
			valid = false
		}
		switch {
		case val == "":
			errs = append(errs, fmt.Errorf("pair %d: empty ISBN prefix for UPC key %q", i/2, key))
			valid = false
		case len(val) < 4 || len(val) > 9:
			errs = append(errs, fmt.Errorf("pair %d: ISBN prefix %q has %d digits, expected 4-9 (the 5-digit add-on supplies the rest)", i/2, val, len(val)))
			valid = false
		case !isDigits(val):
			errs = append(errs, fmt.Errorf("pair %d: ISBN prefix %q contains non-digit characters", i/2, val))
			valid = false
		}
		if valid {
			m[key] = append(m[key], val)
		}
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return m, nil
}

// isDigits reports whether s is non-empty and consists only of ASCII digits.
func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}
