package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRunGetCSVInvalidUTF8(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	cmd.Flags().Bool("raw", false, "")
	cmd.SetIn(strings.NewReader("title\na\xffb\n"))
	var out bytes.Buffer
	cmd.SetOut(&out)

	assert.NoError(t, runGetCSV(cmd, "title"), "invalid UTF-8 in a column value should not fail")
	assert.Equal(t, "\"a\uFFFDb\"\n", out.String(), "invalid UTF-8 should be replaced by U+FFFD in the JSON literal")
}
