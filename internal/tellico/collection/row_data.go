package collection

import (
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/DryHumour/readerware-to-tellico/internal/images"
	"github.com/DryHumour/readerware-to-tellico/internal/strutil"
)

var (
	// categoryPathNodeCutRE is a regex to match nodes in a category path that should be removed.
	categoryPathNodeCutRE = regexp.MustCompile(`(?i:Authors|By (?:Author|Period)|\w{2,}, A-Z|^\(\s*[A-Z]\s*\)$)`) // FIXME(nschelle) move this to generic implementation's source file
)

type rowData struct {
	columns   ColumnConfig
	blocklist map[string]bool
	clean     map[string]string
	images    images.Row
}

func newRowData(info CollectionInfo, clean map[string]string, images images.Row) *rowData {
	return &rowData{
		columns:   info.Columns(),
		blocklist: info.Blocklist(),
		clean:     clean,
		images:    images,
	}
}

// Columns returns the column configuration for this collection.
func (d *rowData) Columns() ColumnConfig {
	return d.columns
}

// Clean returns the plain-text cleaned column values for this entry.
func (d *rowData) Clean() map[string]string {
	return d.clean
}

// Images returns the cover image data for this entry.
func (d *rowData) Images() images.Row {
	return d.images
}

// V returns the plain-text cleaned value for the given column.
// Apply xml in the template at the point of XML output: {{ .V "COL" | xml }}
func (d *rowData) V(col string) string {
	return d.clean[col]
}

// L returns the plain-text split values for a semicolon or slash separated column.
// Apply xml in the template at the point of XML output: {{ range .L "COL" }}{{ . | xml }}{{ end }}
func (d *rowData) L(col string) []string {
	return strutil.SplitList(d.clean[col])
}

// D returns a TellicoDate value for the given column.
func (d *rowData) D(col string) *TellicoDate {
	return NewTellicoDate(d.clean[col])
}

// Is returns true if the given column is "true".
func (d *rowData) Is(col string) bool {
	return d.clean[col] == "true"
}

// Categories returns the plain-text Readerware category values for this entry.
func (d *rowData) Categories() []string {
	var cats []string
	for col := range d.columns.Categories {
		if v := d.clean[col]; v != "" {
			cats = append(cats, v)
		}
	}
	return cats
}

// Genres returns the plain-text genre values derived from Readerware category paths,
// with single-character values, path navigation nodes, and blocklisted entries removed.
func (d *rowData) Genres() []string {
	uniq := make(map[string]struct{})
	for _, path := range d.Categories() {
		if path == "" {
			continue
		}
		for node := range strings.FieldsFuncSeq(path, func(r rune) bool { return r == ':' || r == '|' || r == '>' }) {
			trimmed := strutil.Squeeze(node)
			if categoryPathNodeCutRE.MatchString(trimmed) {
				break
			}
			if len(trimmed) <= 1 || d.blocklist[trimmed] {
				continue
			}
			uniq[trimmed] = struct{}{}
		}
	}
	return slices.Collect(maps.Keys(uniq))
}
