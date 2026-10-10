package cmd

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/assert/yaml"

	"github.com/DryHumour/readerware-to-tellico/isbn"
)

func TestExtractDigits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ISBN-13 with hyphens",
			input:    "978-0-306-40615-7",
			expected: "9780306406157",
		},
		{
			name:     "plain digits pass through",
			input:    "0306406152",
			expected: "0306406152",
		},
		{
			name:     "UPC with spaces",
			input:    "0 70993 00595 5 35740",
			expected: "07099300595535740",
		},
		{
			name:     "ISBN-10 with lowercase x",
			input:    "080442957x",
			expected: "080442957X",
		},
		{
			name:     "no digits",
			input:    "hello # world",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, extractDigits(tt.input), "input: %q", tt.input)
		})
	}
}

func TestValidUPC12(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:  "valid UPC-A",
			input: "070993005955",
		},
		{
			name:    "wrong length",
			input:   "07099300595",
			wantErr: true,
		},
		{
			name:    "invalid check digit",
			input:   "070993005950",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validUPC12(tt.input)
			if tt.wantErr {
				assert.Error(t, err, "input: %q", tt.input)
				return
			}
			assert.NoError(t, err, "input: %q", tt.input)
		})
	}
}

func TestIsbn10FromUPC(t *testing.T) {
	t.Parallel()

	table := map[string][]string{
		"070993": {"0446"},
		"076714": {"0671", "07434"},
	}

	tests := []struct {
		name     string
		upc12    string
		upc5     string
		expected string
		wantErr  bool
	}{
		{
			name:     "UPC-12 with add-on",
			upc12:    "070993005955",
			upc5:     "35740",
			expected: "0446357405",
		},
		{
			name:     "multi-candidate uses first publisher prefix",
			upc12:    "076714000001",
			upc5:     "12345",
			expected: "0671123459",
		},
		{
			name:    "unknown UPC prefix",
			upc12:   "123456005950",
			upc5:    "90000",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := isbn10FromUPC(tt.upc12, tt.upc5, table)
			if tt.wantErr {
				assert.Error(t, err, "upc12: %q upc5: %q", tt.upc12, tt.upc5)
				return
			}
			assert.NoError(t, err, "upc12: %q upc5: %q", tt.upc12, tt.upc5)
			assert.Equal(t, tt.expected, got.String(), "upc12: %q upc5: %q", tt.upc12, tt.upc5)
		})
	}
}

func TestPairsToTable(t *testing.T) {
	t.Parallel()

	got, err := upcPairsToTable([]string{
		"076714", "0671",
		"070993", "0446",
		"076714", "07434",
	})
	assert.NoError(t, err, "pairsToTable returned an unexpected error")
	assert.Equal(t, map[string][]string{
		"070993": {"0446"},
		"076714": {"0671", "07434"},
	}, got, "pairsToTable mismatch")

	_, err = upcPairsToTable([]string{"070993"})
	assert.Error(t, err, "odd-length pair list should return an error")

	_, err = upcPairsToTable([]string{"070993", "0446", "4", "0345"})
	assert.ErrorContains(t, err, `"4"`, "short UPC key should return an error")

	_, err = upcPairsToTable([]string{"070993", ""})
	assert.ErrorContains(t, err, "empty ISBN prefix", "empty ISBN prefix should return an error")

	_, err = upcPairsToTable([]string{"070993", "0123456789"})
	assert.ErrorContains(t, err, "expected 4-9", "over-long ISBN prefix should return an error")

	_, err = upcPairsToTable([]string{"070993", "446"})
	assert.ErrorContains(t, err, "expected 4-9", "too-short ISBN prefix should return an error")

	_, err = upcPairsToTable([]string{"07099A", "0446", "070993", "04X6"})
	assert.ErrorContains(t, err, `"07099A"`, "non-digit UPC key should return an error")
	assert.ErrorContains(t, err, `"04X6"`, "non-digit ISBN prefix should return an error")
}

// The upc list tests are parallel-safe only because they never write to the
// global viper; tests that do must stay sequential so they finish before the
// parallel tests resume.
func TestRunUPCList(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.Flags().Bool("raw", false, "")
	var out bytes.Buffer
	cmd.SetOut(&out)

	assert.NoError(t, runUPCList(cmd), "runUPCList returned an unexpected error")

	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	want := 2
	for _, pubs := range upcToISBNPrefix {
		want += 2 * len(pubs)
	}
	assert.Len(t, lines, want, "header plus two YAML list items per pair")
	assert.Equal(t, "upc:", lines[0], "output should be a complete upc.table fragment")
	assert.Equal(t, "  table:", lines[1], "output should be a complete upc.table fragment")
	assert.Equal(t, `    - "011271"`, lines[2], "first pair should lead the output")
	assert.Equal(t, `    - "0939001"`, lines[3], "first pair should lead the output")
	assert.Equal(t, `    - "777506"`, lines[len(lines)-2], "last pair should end the output")
	assert.Equal(t, `    - "155166"`, lines[len(lines)-1], "last pair should end the output")
	assert.Contains(t, out.String(),
		`    - "076714"`+"\n"+`    - "0671"`+"\n"+`    - "076714"`+"\n"+`    - "07434"`+"\n"+`    - "076714"`+"\n"+`    - "0743"`+"\n",
		"multi-candidate key should repeat the key, one pair per candidate in order")

	var dto upcConfigDTO
	assert.NoError(t, yaml.Unmarshal(out.Bytes(), &dto),
		"output should round-trip through the config DTO")
	assert.NotEmpty(t, dto.UPC.Table, "round-tripped table should not be empty")
}

func TestRunUPCListRaw(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.Flags().Bool("raw", true, "")
	var out bytes.Buffer
	cmd.SetOut(&out)

	assert.NoError(t, runUPCList(cmd), "runUPCList returned an unexpected error")

	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	want := 0
	for _, pubs := range upcToISBNPrefix {
		want += len(pubs)
	}
	assert.Len(t, lines, want, "one output line per pair")
	assert.Equal(t, "011271 0939001", lines[0], "first pair should lead the output")
	assert.Equal(t, "777506 155166", lines[len(lines)-1], "last pair should end the output")
	assert.Contains(t, out.String(), "076714 0671\n076714 07434\n076714 0743\n",
		"multi-candidate key should print one line per candidate in order")
}

func TestRunUPCListDefault(t *testing.T) {
	// Not parallel: mutates the global viper.
	v := viper.GetViper()
	defer v.Set("upc", nil)
	v.Set("upc", map[string]any{"table": []string{"123456", "0777"}})

	newCmd := func(out *bytes.Buffer) *cobra.Command {
		cmd := &cobra.Command{}
		cmd.Flags().Bool("raw", true, "")
		cmd.SetOut(out)
		return cmd
	}

	var eff bytes.Buffer
	assert.NoError(t, runUPCList(newCmd(&eff)), "effective listing should not error")
	assert.Contains(t, eff.String(), "123456 0777", "effective table should include the user pair")

	var bi bytes.Buffer
	assert.NoError(t, runUPCListDefault(newCmd(&bi)), "default listing should not error")
	assert.NotContains(t, bi.String(), "123456", "list default should ignore upc.table config")
	assert.Contains(t, bi.String(), "011271 0939001", "list default should still list built-in pairs")
}

func TestEffectiveUPCTable(t *testing.T) {
	// Not parallel: mutates the global viper.
	v := viper.GetViper()
	defer v.Set("upc", nil)

	table, err := effectiveUPCTable()
	assert.NoError(t, err, "no config section should yield the built-in table")
	assert.Equal(t, []string{"0446"}, table["070993"], "unexpected 070993 candidates")

	v.Set("upc", map[string]any{"table": []string{"070993", "0999"}})
	table, err = effectiveUPCTable()
	assert.NoError(t, err, "valid upc.table should not error")
	assert.Equal(t, []string{"0999", "0446"}, table["070993"], "user pair should take precedence over built-in")
	assert.Equal(t, []string{"1595"}, table["645573"], "built-in pairs should be preserved")

	v.Set("upc", map[string]any{"table": []any{70993, 446}})
	_, err = effectiveUPCTable()
	assert.ErrorContains(t, err, "unquoted scalar", "unquoted numeric scalars should be rejected")
}

func TestRunUPCToISBN(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(t.Context())
	cmd.Flags().Bool("raw", false, "")
	cmd.Flags().Bool("hyphenate", false, "")
	cmd.Flags().Bool("strict", false, "")

	var out, errBuf bytes.Buffer
	cmd.SetIn(strings.NewReader("9780306406157\n" +
		"9780306406157 51695\n" +
		"0306406152\n" +
		"\n" +
		"# comment\n" +
		"070993005955\n" +
		"0 70993 00595 5 35740\n" +
		"64557300001112345\n" +
		"05069400001512345\n" +
		"034057000010 12345\n" +
		"07099300595X\n" +
		"0 70993 00595 5 3574x\n" +
		"99999999999312345\n" +
		"0070993005955\n" +
		"0070993005950\n" +
		"0 070993 005955 35740\n" +
		"4006381333931\n" +
		"0306406153\n" +
		"9781060000001\n" +
		"812532570\n" +
		"70999003757\n" +
		"7099900375735740\n" +
		"not a barcode\n"))
	cmd.SetOut(&out)

	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&errBuf, nil)))
	defer slog.SetDefault(old)

	assert.ErrorContains(t, runUPCToISBN(cmd, nil, isbn.DefaultHyphenator()), "9 of 21 inputs failed to convert",
		"unconvertible inputs should fail the command after all lines are processed")

	expectedOut := strings.Join([]string{
		`"9780306406157"`,
		`"9780306406157"`,
		`"0306406152"`,
		"",
		"# comment",
		`"070993005955"`,
		`"0446357405"`,
		`"1595123458"`,
		`"034512345X"`,
		`"0931873452"`,
		`"07099300595X"`,
		`"0 70993 00595 5 3574x"`,
		`"99999999999312345"`,
		`"0070993005955"`,
		`"0070993005950"`,
		`"0446357405"`,
		`"4006381333931"`,
		`"0306406153"`,
		`"9781060000001"`,
		`"0812532570"`,
		`"70999003757"`,
		`"034535740X"`,
		`"not a barcode"`,
	}, "\n") + "\n"

	assert.Equal(t, expectedOut, out.String(), "stdout output mismatch")

	errOut := errBuf.String()
	assert.Contains(t, errOut, "requires a 5-digit add-on", "bare UPC-12 should require its add-on")
	assert.Contains(t, errOut, "invalid UPC-A check digit", "bad-check-digit UPC should report its own error")
	assert.Contains(t, errOut, "unexpected 'X'", "X in UPC input should report a non-digit error")
	assert.Contains(t, errOut, "unknown UPC prefix 999999", "missing UPC error message")
	assert.Contains(t, errOut, "unrecognised EAN-13 prefix 400", "missing EAN-13 prefix error message")
	assert.Contains(t, errOut, "unrecognized input format", "missing format error message")
	assert.Equal(t, 2, strings.Count(errOut, "output is not a valid, range-resolvable ISBN"),
		"bad-check ISBN-10 and unassigned publisher block should each warn once")
	assert.Contains(t, errOut, "invalid ISBN check digit", "check-digit problem should be named in the error attribute")
	assert.Contains(t, errOut, "restored a lost leading zero", "a successful zero-restore retry should warn")
	assert.Contains(t, errOut, "also tried with a leading zero", "a failed zero-restore retry should join both errors")
}

func TestRunUPCToISBNStrict(t *testing.T) {
	// Not parallel: mutates the global slog logger.
	cmd := &cobra.Command{}
	cmd.SetContext(t.Context())
	cmd.Flags().Bool("raw", false, "")
	cmd.Flags().Bool("hyphenate", false, "")
	cmd.Flags().Bool("strict", true, "")

	var out, errBuf bytes.Buffer
	cmd.SetIn(strings.NewReader("0306406153\n" +
		"0 70993 00595 5 35740\n" +
		"9781060000001\n"))
	cmd.SetOut(&out)

	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&errBuf, nil)))
	defer slog.SetDefault(old)

	assert.ErrorContains(t, runUPCToISBN(cmd, nil, isbn.DefaultHyphenator()), "2 of 3 inputs failed to convert",
		"strict mode should fail bad-check and unassigned-publisher outputs")

	assert.Equal(t, `"0306406153"`+"\n"+`"0446357405"`+"\n"+`"9781060000001"`+"\n",
		out.String(), "stdout output mismatch")

	errOut := errBuf.String()
	assert.Contains(t, errOut, "invalid ISBN check digit", "strict bad-check ISBN should fail in convert")
	assert.Contains(t, errOut, "output is not a valid, range-resolvable ISBN", "strict unassigned publisher block should fail in the filter")
	assert.NotContains(t, errOut, "passing through ISBN with invalid check digit",
		"strict mode should not warn-and-passthrough on check digit failures")
}

func TestRunUPCToISBNRaw(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(t.Context())
	cmd.Flags().Bool("raw", true, "")
	cmd.Flags().Bool("hyphenate", false, "")
	cmd.Flags().Bool("strict", false, "")

	var out bytes.Buffer
	cmd.SetIn(strings.NewReader("0 70993 00595 5 35740\n0306406152\n"))
	cmd.SetOut(&out)

	assert.NoError(t, runUPCToISBN(cmd, nil, isbn.DefaultHyphenator()),
		"raw run should not error")
	assert.Equal(t, "0446357405\n0306406152\n", out.String(), "raw output should be unquoted")
}

func TestRunUPCToISBNHyphenate(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.SetContext(t.Context())
	cmd.Flags().Bool("raw", false, "")
	cmd.Flags().Bool("hyphenate", true, "")
	cmd.Flags().Bool("strict", false, "")

	var out bytes.Buffer
	cmd.SetIn(strings.NewReader("0 70993 00595 5 35740\n"))
	cmd.SetOut(&out)

	assert.NoError(t, runUPCToISBN(cmd, nil, isbn.DefaultHyphenator()),
		"hyphenate run should not error")
	assert.Equal(t, `"0-446-35740-5"`+"\n", out.String(), "hyphenated ISBN-10 output mismatch")
}

func TestUPCToISBNRestoringZero(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		strict   bool
		expected string
		wantErr  string
		notErr   string
		wantWarn bool
	}{
		{
			name:     "9-digit ISBN-10 restores a leading zero",
			input:    "812532570",
			expected: "0812532570",
			wantWarn: true,
		},
		{
			name:    "9-digit with bad check digit after padding",
			input:   "812532571",
			strict:  true,
			wantErr: "also tried with a leading zero",
		},
		{
			name:    "11-digit UPC-12 still needs its add-on",
			input:   "70999003757",
			wantErr: "requires a 5-digit add-on",
		},
		{
			name:     "16-digit UPC-12 with add-on restores a leading zero",
			input:    "7099900375735740",
			expected: "034535740X",
			wantWarn: true,
		},
		{
			name:    "8-digit input does not retry",
			input:   "12345678",
			wantErr: "unrecognized input format",
			notErr:  "also tried",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var errBuf bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&errBuf, nil))
			got, err := upcToISBNRestoringZero(t.Context(), logger, tt.input, upcToISBNPrefix, isbn.DefaultHyphenator(), tt.strict)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr, "input %q", tt.input)
				if tt.notErr != "" {
					assert.NotContains(t, err.Error(), tt.notErr, "input %q", tt.input)
				}
				return
			}
			assert.NoError(t, err, "input %q", tt.input)
			assert.Equal(t, tt.expected, got.String(), "input %q", tt.input)
			if tt.wantWarn {
				assert.Contains(t, errBuf.String(), "restored a lost leading zero", "input %q", tt.input)
			}
		})
	}
}
