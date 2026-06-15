package isbn

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type stubHTTPClient struct {
	do func(req *http.Request) (*http.Response, error)
}

func (s stubHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return s.do(req)
}

func httpResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func TestFetchISBNRanges_JSON(t *testing.T) {
	t.Parallel()
	client := stubHTTPClient{do: func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != isbnRangesURL {
			return httpResponse(http.StatusNotFound, ""), nil
		}
		return httpResponse(http.StatusOK, `{"ISBNRangeMessage":{"MessageDate":"2026-01-01","RegistrationGroups":{"Group":[]}}}`), nil
	}}

	ranges, err := FetchISBNRanges(t.Context(), client)
	assert.NoError(t, err)
	assert.Equal(t, "2026-01-01", ranges.ISBNRangeMessage.MessageDate)
}

func TestFetchISBNRanges_XMLFallback(t *testing.T) {
	t.Parallel()
	client := stubHTTPClient{do: func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case isbnRangesURL:
			return httpResponse(http.StatusInternalServerError, ""), nil
		case rangeMessageURL:
			return httpResponse(http.StatusOK, `<?xml version="1.0" encoding="UTF-8"?>
<ISBNRangeMessage>
  <MessageDate>2026-02-02</MessageDate>
  <RegistrationGroups>
    <Group>
      <Prefix>978-0</Prefix>
      <Agency>Test</Agency>
      <Rules>
        <Rule>
          <Range>0000000-1999999</Range>
          <Length>2</Length>
        </Rule>
      </Rules>
    </Group>
  </RegistrationGroups>
</ISBNRangeMessage>`), nil
		default:
			return httpResponse(http.StatusNotFound, ""), nil
		}
	}}

	ranges, err := FetchISBNRanges(t.Context(), client)
	assert.NoError(t, err)
	assert.Equal(t, "2026-02-02", ranges.ISBNRangeMessage.MessageDate)
	if assert.Len(t, ranges.ISBNRangeMessage.RegistrationGroups.Group, 1) {
		assert.Equal(t, "978-0", ranges.ISBNRangeMessage.RegistrationGroups.Group[0].Prefix)
		assert.Len(t, ranges.ISBNRangeMessage.RegistrationGroups.Group[0].Rules.Rule, 1)
		assert.Equal(t, "0000000-1999999", ranges.ISBNRangeMessage.RegistrationGroups.Group[0].Rules.Rule[0].Range)
		assert.Equal(t, "2", ranges.ISBNRangeMessage.RegistrationGroups.Group[0].Rules.Rule[0].Length)
	}
}
