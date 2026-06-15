package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/tidwall/gjson"
)

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
		if !gjson.Valid(jsonField) {
			slog.Warn("invalid edition JSON, skipping")
			continue
		}
		results := gjson.GetMany(jsonField, "isbn_10", "isbn_13")
		for _, r := range results {
			r.ForEach(func(_, v gjson.Result) bool {
				fmt.Fprintln(out, v.String())
				return true
			})
		}
	}
	if err := scanner.Err(); err != nil {
		slog.Error("error reading stdin", "error", err)
		os.Exit(1)
	}
}
