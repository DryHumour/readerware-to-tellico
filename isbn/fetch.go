package isbn

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	ErrOversizeResponse = errors.New("oversize response")
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// FetchISBNRanges attempts to fetch ISBN range data from multiple sources,
// falling back to embedded data if all sources fail.
func FetchISBNRanges(ctx context.Context, client HTTPClient) (ISBNRanges, error) {
	var errs []error

	if data, err := fetchURL(ctx, client, isbnRangesURL); err == nil {
		if ranges, err := ParseISBNRangesJSON(data); err == nil {
			return ranges, nil
		} else {
			errs = append(errs, fmt.Errorf("failed to parse ISBN ranges JSON: %w", err))
		}
	} else {
		errs = append(errs, fmt.Errorf("failed to fetch ISBN ranges JSON: %s: %w", isbnRangesURL, err))
	}

	if data, err := fetchURL(ctx, client, rangeMessageURL); err == nil {
		if msg, err := ParseRangeMessageXML(data); err == nil {
			return ISBNRangesFromXML(msg), nil
		} else {
			errs = append(errs, fmt.Errorf("failed to parse range message XML: %w", err))
		}
	} else {
		errs = append(errs, fmt.Errorf("failed to fetch range message XML: %s: %w", rangeMessageURL, err))
	}

	if fallback, err := ParseISBNRangesJSON(isbnRangesJSON); err == nil {
		return fallback, nil
	} else {
		errs = append(errs, fmt.Errorf("failed to parse fallback ISBN ranges JSON: %w", err))
	}

	return ISBNRanges{}, fmt.Errorf("all ISBN range fetch methods failed: %w", errors.Join(errs...))
}

// fetchURL fetches data from a URL using the provided HTTP client.
func fetchURL(ctx context.Context, client HTTPClient, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// avoid potential DOS from large error bodies, at the cost of connection re-use
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	const maxPayloadSize = 2 * 1024 * 1024 // 2MB (arbitrary, but we only expect 250kB or so anyway)
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxPayloadSize))
	if err != nil {
		return nil, err
	}
	if len(data) == maxPayloadSize {
		return nil, ErrOversizeResponse
	}

	return data, nil
}
