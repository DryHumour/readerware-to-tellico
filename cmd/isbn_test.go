package cmd

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/DryHumour/readerware-to-tellico/isbn"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestISBNTo13Command(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "to13", isbnTo13Cmd.Name(), "command name guards the viper key")
	assert.Contains(t, isbnTo13Cmd.Aliases, "13", "to13 should be reachable as \"13\"")
}

// newISBNTo13TestCmd builds a bare command with the to13 flags set to the
// given values and the given stdin input attached.
func newISBNTo13TestCmd(input string, raw, hyphenate, strict bool) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("raw", raw, "")
	cmd.Flags().Bool("hyphenate", hyphenate, "")
	cmd.Flags().Bool("strict", strict, "")
	cmd.SetIn(strings.NewReader(input))
	var out bytes.Buffer
	cmd.SetOut(&out)
	var errBuf bytes.Buffer
	return cmd, &out, &errBuf
}

func TestRunISBNTo13(t *testing.T) {
	// Not parallel: mutates the global slog logger.
	cmd, out, errBuf := newISBNTo13TestCmd("0306406152\n"+
		"978-0-306-40615-7\n"+
		"0-306-40615-2\n"+
		"9780306406157\n"+
		"9791090636071\n"+
		`"0306406152"`+"\n"+
		"\n"+
		"# c\n"+
		"0306406153\n"+
		"0070993005955\n"+
		"not an isbn\n", false, false, false)
	cmd.SetContext(t.Context())

	old := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(errBuf, nil)))
	defer slog.SetDefault(old)

	assert.ErrorContains(t, runISBNTo13(cmd, nil, isbn.DefaultHyphenator()), "3 of 9 inputs failed to convert",
		"unconvertible inputs should fail the command after all lines are processed")

	expectedOut := strings.Join([]string{
		`"9780306406157"`,
		`"9780306406157"`,
		`"9780306406157"`,
		`"9780306406157"`,
		`"9791090636071"`,
		`"9780306406157"`,
		"",
		"# c",
		`"0306406153"`,
		`"0070993005955"`,
		`"not an isbn"`,
	}, "\n") + "\n"
	assert.Equal(t, expectedOut, out.String(), "stdout output mismatch")

	errOut := errBuf.String()
	assert.Contains(t, errOut, "check digit", "bad-check-digit input should log a check digit error")
	assert.Contains(t, errOut, "not a book ISBN prefix", "non-book EAN-13 should log a prefix error")
}

func TestRunISBNTo13Raw(t *testing.T) {
	// Not parallel: mutates the global slog logger.
	old := slog.Default()
	defer slog.SetDefault(old)
	slog.SetDefault(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))

	cmd, out, _ := newISBNTo13TestCmd("0306406152\n"+`"bogus"`+"\n", true, false, false)
	cmd.SetContext(t.Context())

	assert.ErrorContains(t, runISBNTo13(cmd, nil, isbn.DefaultHyphenator()), "1 of 2 inputs failed",
		"one bad line should fail the command")
	assert.Equal(t, "9780306406157\nbogus\n", out.String(), "raw output should be unquoted, failed input echoed decoded")
}

func TestRunISBNTo13Hyphenate(t *testing.T) {
	// Not parallel: mutates the global slog logger.
	old := slog.Default()
	defer slog.SetDefault(old)
	slog.SetDefault(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))

	cmd, out, _ := newISBNTo13TestCmd("0306406152\n", false, true, false)
	cmd.SetContext(t.Context())

	assert.NoError(t, runISBNTo13(cmd, nil, isbn.DefaultHyphenator()), "hyphenate run should not error")
	assert.Equal(t, `"978-0-306-40615-7"`+"\n", out.String(), "hyphenated output mismatch")
}

func TestRunISBNTo13UnassignedPublisherBlock(t *testing.T) {
	// Not parallel: mutates the global slog logger.
	old := slog.Default()
	defer slog.SetDefault(old)

	// 9781060000001 sits in registration group 978-1's unassigned publisher
	// block (rule range 0600000-0664999 with Length 0).
	i, err := isbn.Make13("978106000000")
	if err != nil {
		t.Errorf("Make13 should not fail for a 12-digit body: %v", err)
	}
	unassigned := i.String()
	assert.Equal(t, "9781060000001", unassigned, "check digit computation changed")

	t.Run("default accepts with warning", func(t *testing.T) {
		var errBuf bytes.Buffer
		slog.SetDefault(slog.New(slog.NewTextHandler(&errBuf, nil)))

		cmd, out, _ := newISBNTo13TestCmd(unassigned+"\n", false, false, false)
		cmd.SetContext(t.Context())

		assert.NoError(t, runISBNTo13(cmd, nil, isbn.DefaultHyphenator()),
			"unassigned publisher block should be accepted by default")
		assert.Equal(t, `"`+unassigned+`"`+"\n", out.String(), "output mismatch")
		assert.Contains(t, errBuf.String(), "output is not a valid, range-resolvable ISBN", "a warning should be logged")
	})

	t.Run("strict rejects", func(t *testing.T) {
		var errBuf bytes.Buffer
		slog.SetDefault(slog.New(slog.NewTextHandler(&errBuf, nil)))

		cmd, out, _ := newISBNTo13TestCmd(unassigned+"\n", false, false, true)
		cmd.SetContext(t.Context())

		assert.ErrorContains(t, runISBNTo13(cmd, nil, isbn.DefaultHyphenator()), "1 of 1 inputs failed",
			"strict mode should reject the unassigned publisher block")
		assert.Equal(t, `"`+unassigned+`"`+"\n", out.String(), "failed input should echo unchanged")
		assert.Contains(t, errBuf.String(), "output is not a valid, range-resolvable ISBN", "strict failure should be logged")
	})

	t.Run("hyphenate falls back to bare digits", func(t *testing.T) {
		var errBuf bytes.Buffer
		slog.SetDefault(slog.New(slog.NewTextHandler(&errBuf, nil)))

		cmd, out, _ := newISBNTo13TestCmd(unassigned+"\n", false, true, false)
		cmd.SetContext(t.Context())

		assert.NoError(t, runISBNTo13(cmd, nil, isbn.DefaultHyphenator()),
			"unassigned publisher block should be accepted by default")
		assert.Equal(t, `"`+unassigned+`"`+"\n", out.String(), "output should be unhyphenated digits")
		assert.Contains(t, errBuf.String(), "output is not a valid, range-resolvable ISBN", "a warning should be logged")
	})
}

func TestRunISBNTo13Args(t *testing.T) {
	// Not parallel: mutates the global slog logger.
	old := slog.Default()
	defer slog.SetDefault(old)
	slog.SetDefault(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)))

	cmd, out, _ := newISBNTo13TestCmd("", false, false, false)
	cmd.SetContext(t.Context())

	assert.ErrorContains(t, runISBNTo13(cmd, []string{"0306406152", `"bogus"`}, isbn.DefaultHyphenator()),
		"1 of 2 inputs failed", "one bad arg should fail the command")
	assert.Equal(t, `"9780306406157"`+"\n"+`"bogus"`+"\n", out.String(), "stdout output mismatch")
}
