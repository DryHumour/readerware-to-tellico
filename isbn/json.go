package isbn

import (
	_ "embed"
	"encoding/json/jsontext"
	json "encoding/json/v2"
	"fmt"
)

const (
	// isbnRangesURL is the canonocal URL for the ISBN ranges JSON file.
	isbnRangesURL = "https://isbnbarcode.org/api/isbn-ranges.json"
)

// isbnRangesJSON is the fallback ISBN ranges JSON file.
//
//go:embed isbn-ranges.json
var isbnRangesJSON []byte

// ISBNRanges represents the structure of an ISBN ranges document.
//
//nolint:revive // name mirrors the ISBN Range Message schema
type ISBNRanges struct {
	ISBNRangeMessage struct {
		MessageDate        string `json:"MessageDate"`
		RegistrationGroups struct {
			Group []GroupRule `json:"Group"`
		} `json:"RegistrationGroups"`
	} `json:"ISBNRangeMessage"`
}

// GroupRule represents a group rule in an ISBN ranges document.
type GroupRule struct {
	Prefix string `json:"Prefix"`
	Rules  struct {
		Rule RangeRuleList `json:"Rule"`
	} `json:"Rules"`
}

// RangeRuleList is a JSON decoding helper for fields that are sometimes encoded
// as a single object and sometimes as an array.
type RangeRuleList []RangeRule

// UnmarshalJSONFrom implements json.UnmarshalerFrom for the object-or-array
// encoding used by the ISBN ranges document. It must read exactly one JSON
// value from the decoder; nested values are handled by UnmarshalDecode.
func (l *RangeRuleList) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	switch dec.PeekKind() {
	case jsontext.KindNull:
		return dec.SkipValue()
	case jsontext.KindBeginObject:
		var r RangeRule
		if err := json.UnmarshalDecode(dec, &r); err != nil {
			return err
		}
		*l = RangeRuleList{r}
		return nil
	case jsontext.KindBeginArray:
		var rs []RangeRule
		if err := json.UnmarshalDecode(dec, &rs); err != nil {
			return err
		}
		*l = RangeRuleList(rs)
		return nil
	default:
		return fmt.Errorf("unexpected JSON token %s", dec.PeekKind())
	}
}

// RangeRule represents a range rule in an ISBN ranges document.
type RangeRule struct {
	Range  string `json:"Range"`
	Length string `json:"Length"`
}

// ParseISBNRangesJSON parses the JSON data from an ISBN ranges document into the
// corresponding ISBNRanges struct.
func ParseISBNRangesJSON(data []byte) (ISBNRanges, error) {
	var result ISBNRanges
	if err := json.Unmarshal(data, &result); err != nil {
		return ISBNRanges{}, fmt.Errorf("parse ISBN ranges: %w", err)
	}
	return result, nil
}
