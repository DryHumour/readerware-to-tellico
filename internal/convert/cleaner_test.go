package convert

import (
	"errors"
	"testing"
	"text/template"

	"github.com/DryHumour/readerware-to-tellico/internal/tellico/collection"
	"github.com/stretchr/testify/require"
)

func mustTemplate(t *testing.T, src string) *template.Template {
	t.Helper()
	tmpl, err := template.New("test").Parse(src)
	require.NoError(t, err)
	return tmpl
}

func TestCleaner_CleanCell(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		tmpl      string
		cell      CellData
		want      string
		wantError bool
	}{
		{
			name: "falls back to clean.default",
			tmpl: `{{define "clean.default"}}{{.Value}}{{end}}`,
			cell: CellData{Column: "AUTHOR", Value: "Raw"},
			want: "Raw",
		},
		{
			name: "uses column-specific template when present",
			tmpl: `{{define "clean.default"}}{{.Value}}{{end}}{{define "clean.AUTHOR"}}X{{.Value}}Y{{end}}`,
			cell: CellData{Column: "AUTHOR", Value: "Raw"},
			want: "XRawY",
		},
		{
			name:      "returns error when template execution fails",
			tmpl:      `{{define "clean.default"}}{{.Value}}{{end}}{{define "clean.AUTHOR"}}{{.Nope}}{{end}}`,
			cell:      CellData{Column: "AUTHOR", Value: "Raw"},
			want:      "",
			wantError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := cleaner{
				tmpl:   mustTemplate(t, tc.tmpl),
				logger: testLogger,
			}
			got, err := c.CleanCell(t.Context(), tc.cell)
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestCleaner_CleanRow(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                string
		tmpl                string
		raw                 map[string]string
		wantClean           map[string]string
		wantColumnErrorCols []string
	}{
		{
			name: "cleans all columns and aggregates errors",
			tmpl: `
{{define "clean.default"}}{{.Value}}{{end}}
{{define "clean.TITLE"}}T:{{.Value}}{{end}}
{{define "clean.BAD"}}{{.Nope}}{{end}}
`,
			raw: map[string]string{"TITLE": "A", "BAD": "B"},
			wantClean: map[string]string{
				"TITLE": "T:A",
				"BAD":   "",
			},
			wantColumnErrorCols: []string{"BAD"},
		},
		{
			name:      "no errors",
			tmpl:      `{{define "clean.default"}}{{.Value}}{{end}}{{define "clean.TITLE"}}T:{{.Value}}{{end}}`,
			raw:       map[string]string{"TITLE": "A"},
			wantClean: map[string]string{"TITLE": "T:A"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c := cleaner{
				tmpl: mustTemplate(t, tc.tmpl),
				colCfg: collection.ColumnConfig{
					Names:   map[string][]string{"author": {"AUTHOR"}},
					Headers: []string{"TITLE", "BAD"},
				},
				logger: testLogger,
			}
			clean, err := c.CleanRow(t.Context(), tc.raw)
			require.Equal(t, tc.wantClean, clean, "clean map mismatch")

			if len(tc.wantColumnErrorCols) == 0 {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)

			for _, col := range tc.wantColumnErrorCols {
				var ce ColumnError
				require.True(t, errors.As(err, &ce))
				require.Equal(t, col, ce.Column)
			}
		})
	}
}
