package convert

import (
	"errors"
	"net/http"
	"os"
	"testing"
)

// errOffline is returned by the stub transport for every request.
var errOffline = errors.New("network access disabled in tests")

// offlineTransport rejects all HTTP requests so ISBN range fetches
// fall back to embedded data instead of hitting the network.
type offlineTransport struct{}

func (offlineTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errOffline
}

func TestMain(m *testing.M) {
	http.DefaultTransport = offlineTransport{}
	os.Exit(m.Run())
}
