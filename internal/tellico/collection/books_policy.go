package collection

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/DryHumour/readerware-to-tellico/internal/config"
	"github.com/DryHumour/readerware-to-tellico/internal/images"
	"github.com/DryHumour/readerware-to-tellico/isbn"
	"github.com/bartventer/httpcache"
	_ "github.com/bartventer/httpcache/store/fscache"
	_ "github.com/bartventer/httpcache/store/memcache"
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
	// logger is the logger to use for logging.
	logger *slog.Logger
}

// NewBooksPolicy creates a books policy and refreshes ISBN ranges best-effort.
// The genre blocklist is not set here; call Configure after executing the config template.
func NewBooksPolicy(ctx context.Context, cfg *config.Config) *BooksPolicy {
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	logger = logger.With("kind", KindBooks)
	refreshISBNHyphenator(ctx, logger)
	return &BooksPolicy{
		info: collectionInfo{
			kind:          KindBooks,
			templateNames: booksTemplateNames,
		},
		logger: logger,
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

func refreshISBNHyphenator(ctx context.Context, logger *slog.Logger) {
	client := &http.Client{
		Transport: httpcache.NewTransport(
			"fscache://?appname=readerware-to-tellico",
			httpcache.WithSWRTimeout(10*time.Second)),
		Timeout: 10 * time.Second,
	}
	ranges, err := isbn.FetchISBNRanges(ctx, client)
	if err != nil {
		logger.WarnContext(ctx, "failed to fetch latest ISBN ranges; using embedded fallback", "error", err)
		return
	}
	h, err := isbn.NewHyphenator(ranges)
	if err != nil {
		logger.WarnContext(ctx, "failed to compile fetched ISBN ranges; using embedded fallback", "error", err)
		return
	}
	isbn.DefaultHyphenator.Set(h)
}
