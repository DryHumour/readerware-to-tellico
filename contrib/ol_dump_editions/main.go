package main

import (
	"bufio"
	"encoding/json/jsontext"
	json "encoding/json/v2"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// edition is the subset of an Open Library edition record we care about.
// The isbn_10 and isbn_13 members are normally arrays of strings.
type edition struct {
	ISBN10 stringList `json:"isbn_10"`
	ISBN13 stringList `json:"isbn_13"`
}

// stringList decodes a JSON member that may be a single string or an array
// of strings (gjson's ForEach had the same tolerance for non-array values).
type stringList []string

// UnmarshalJSONFrom implements json.UnmarshalerFrom for the string-or-array
// encoding. It must read exactly one JSON value from the decoder.
func (l *stringList) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if dec.PeekKind() == jsontext.KindString {
		var s string
		if err := json.UnmarshalDecode(dec, &s); err != nil {
			return err
		}
		*l = stringList{s}
		return nil
	}
	var ss []string
	if err := json.UnmarshalDecode(dec, &ss); err != nil {
		return err
	}
	*l = stringList(ss)
	return nil
}

func main() {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	const maxLineSize = 16 * 1024 * 1024 // 16 MiB — dump lines can be very large
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, maxLineSize), maxLineSize)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.SplitN(line, "\t", -1)
		if len(fields) == 0 {
			continue
		}
		if fields[0] != "/type/edition" {
			slog.Warn("unexpected record type, skipping", "type", fields[0])
			continue
		}
		jsonField := fields[len(fields)-1]
		var ed edition
		if err := json.Unmarshal([]byte(jsonField), &ed); err != nil {
			slog.Warn("invalid edition JSON, skipping", "error", err)
			continue
		}
		for _, s := range ed.ISBN10 {
			fmt.Fprintln(out, s)
		}
		for _, s := range ed.ISBN13 {
			fmt.Fprintln(out, s)
		}
	}
	if err := scanner.Err(); err != nil {
		slog.Error("error reading stdin", "error", err)
		os.Exit(1)
	}
}
