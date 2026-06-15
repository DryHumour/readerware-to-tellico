// Package httpclient provides a pre-configured HTTP client with a filesystem-backed cache.
package httpclient

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/bartventer/httpcache"
	_ "github.com/bartventer/httpcache/store/fscache"
	_ "github.com/bartventer/httpcache/store/memcache"
)

// New returns an *http.Client backed by a filesystem-cached HTTP transport.
func New(logger *slog.Logger) *http.Client {
	return &http.Client{
		Transport: httpcache.NewTransport(
			"fscache://?appname=readerware-to-tellico",
			httpcache.WithSWRTimeout(10*time.Second),
			httpcache.WithLogger(logger),
		),
		Timeout: 10 * time.Second,
	}
}
