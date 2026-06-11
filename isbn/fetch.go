package isbn

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func FetchISBNRanges(ctx context.Context, client HTTPClient) (ISBNRanges, error) {
	var errs []error

	if data, err := fetchURL(ctx, client, isbnRangesURL); err == nil {
		if ranges, err := ParseISBNRangesJSON(data); err == nil {
			return ranges, nil
		} else {
			errs = append(errs, fmt.Errorf("failed to parse ISBN ranges JSON: %w", err))
		}
	} else {
		errs = append(errs, fmt.Errorf("failed to fetch ISBN ranges JSON: %w", err))
	}

	if data, err := fetchURL(ctx, client, rangeMessageURL); err == nil {
		if msg, err := ParseRangeMessageXML(data); err == nil {
			return isbnRangesFromXML(msg), nil
		} else {
			errs = append(errs, fmt.Errorf("failed to parse range message XML: %w", err))
		}
	} else {
		errs = append(errs, fmt.Errorf("failed to fetch range message XML: %w", err))
	}

	if fallback, err := ParseISBNRangesJSON(isbnRangesJSON); err == nil {
		return fallback, nil
	} else {
		errs = append(errs, fmt.Errorf("failed to parse fallback ISBN ranges JSON: %w", err))
	}

	return ISBNRanges{}, fmt.Errorf("all ISBN range fetch methods failed: %w", errors.Join(errs...))
}

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

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func isbnRangesFromXML(msg ISBNRangeMessage) ISBNRanges {
	var out ISBNRanges
	out.ISBNRangeMessage.MessageDate = msg.MessageDate
	for _, g := range msg.RegistrationGroups.Group {
		gr := GroupRule{Prefix: g.Prefix}
		for _, r := range g.Rules.Rule {
			gr.Rules.Rule = append(gr.Rules.Rule, RangeRule{Range: r.Range, Length: r.Length})
		}
		out.ISBNRangeMessage.RegistrationGroups.Group = append(out.ISBNRangeMessage.RegistrationGroups.Group, gr)
	}
	return out
}
