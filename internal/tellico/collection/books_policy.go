package collection

import (
	"fmt"
	"slices"

	"github.com/DryHumour/readerware-to-tellico/internal/images"
)

var (
	_ Policy = (*BooksPolicy)(nil)

	booksTemplateNames = TemplateNames{
		Config: "books.config",
		Header: "books.header",
		Entry:  "books.entry",
		Footer: "books.footer",
	}
)

// BooksPolicy implements collection policy behavior for Readerware books exports.
type BooksPolicy struct {
	// info holds the collection configuration for this policy.
	info collectionInfo
}

// NewBooksPolicy creates a books policy.
// The genre blocklist is not set here; call Configure after executing the config template.
func NewBooksPolicy() *BooksPolicy {
	return &BooksPolicy{
		info: collectionInfo{
			kind:          KindBooks,
			templateNames: booksTemplateNames,
		},
	}
}

func (p *BooksPolicy) Info() CollectionInfo {
	return &p.info
}

func (p *BooksPolicy) ConfigureHeaders(headers []string, imagesEnabled bool) error {
	headerSet := make(map[string]bool, len(headers))
	for _, h := range headers {
		if headerSet[h] {
			return fmt.Errorf("duplicate header: %q appears more than once", h)
		}
		headerSet[h] = true
	}
	if !headerSet["TITLE"] {
		return fmt.Errorf("missing required header: TITLE")
	}
	if imagesEnabled {
		if !headerSet["ROWKEY"] {
			return fmt.Errorf("missing required header: ROWKEY is required when image directories are specified")
		}
	}
	if !imagesEnabled && !headerSet["ROWKEY"] && !headerSet["ROW#"] {
		return fmt.Errorf("missing required header: either ROWKEY or ROW# must be present")
	}
	p.info.columns.Headers = slices.Clone(headers)
	return nil
}

func (p *BooksPolicy) NewEntry(clean map[string]string, img images.Row) (Entry, error) {
	return newBooksEntry(p.Info(), clean, img)
}
