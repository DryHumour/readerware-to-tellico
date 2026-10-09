package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/DryHumour/readerware-to-tellico/isbn"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var errSentinel = errors.New("sentinel")

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) {
	return 0, errSentinel
}

func newLineFilterCmd(t *testing.T, in io.Reader, out io.Writer) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{}
	cmd.SetContext(t.Context())
	cmd.SetIn(in)
	cmd.SetOut(out)
	return cmd
}

func stubConvert(s string) (isbn.ISBN, error) {
	return isbn.New(s)
}

func stubConvertTolerant(s string) (isbn.ISBN, error) {
	i, err := isbn.New(s)
	if errors.Is(err, isbn.ErrInvalidCheckDigit) {
		return i, nil
	}
	return i, err
}

func TestRunLineFilterCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancelCause(t.Context())
	cancel(errSentinel)

	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	cmd.SetIn(strings.NewReader("0306406152\n0306406152\n"))
	cmd.SetOut(&out)

	err := runLineFilter(cmd, nil, stubConvert, lineFilterOptions{raw: true, hyphenator: isbn.DefaultHyphenator()})
	assert.ErrorIs(t, err, errSentinel, "a cancelled context should abort the loop")
	assert.Empty(t, out.String(), "no lines should be written once cancelled")
}

func TestRunLineFilterScannerError(t *testing.T) {
	t.Parallel()

	in := io.MultiReader(strings.NewReader("0306406152\n"), iotest.ErrReader(errSentinel))
	var out bytes.Buffer
	cmd := newLineFilterCmd(t, in, &out)

	err := runLineFilter(cmd, nil, stubConvert, lineFilterOptions{raw: true, hyphenator: isbn.DefaultHyphenator()})
	assert.ErrorIs(t, err, errSentinel, "scanner failure should propagate the wrapped sentinel")
	assert.ErrorContains(t, err, "reading stdin", "scanner failure should be labelled")
	assert.Equal(t, "0306406152\n", out.String(), "the line read before the error should still be written")
}

func TestRunLineFilterWriterError(t *testing.T) {
	t.Parallel()

	cmd := newLineFilterCmd(t, strings.NewReader("0306406152\n"), errWriter{})

	err := runLineFilter(cmd, nil, stubConvert, lineFilterOptions{raw: true, hyphenator: isbn.DefaultHyphenator()})
	assert.ErrorIs(t, err, errSentinel, "writer failure should propagate the wrapped sentinel")
	assert.ErrorContains(t, err, "writing output", "writer failure should be labelled")
}

func TestRunLineFilterArgsNotSkipped(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := newLineFilterCmd(t, nil, &out)

	err := runLineFilter(cmd, []string{"", "# not a comment", "0306406152"},
		stubConvert, lineFilterOptions{raw: true, hyphenator: isbn.DefaultHyphenator()})
	assert.ErrorContains(t, err, "2 of 3 inputs failed to convert", "blank and '#' args are converted, not skipped")
	assert.Equal(t, "\n# not a comment\n0306406152\n", out.String(), "raw mode echoes failed inputs bare")
}

func TestRunLineFilterCounts(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	cmd := newLineFilterCmd(t, strings.NewReader("\n# c\n0306406152\n0306406153\nbogus\n"), &out)

	err := runLineFilter(cmd, nil, stubConvertTolerant,
		lineFilterOptions{raw: false, strict: true, hyphenator: isbn.DefaultHyphenator()})
	assert.ErrorContains(t, err, "2 of 3 inputs failed to convert",
		"strict check-digit failure and a parse failure, with passthrough lines uncounted")
	assert.Equal(t, "\n# c\n\"0306406152\"\n\"0306406153\"\n\"bogus\"\n", out.String(),
		"strict failures echo the input")
}
