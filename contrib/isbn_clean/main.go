package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/DryHumour/readerware-to-tellico/isbn"
)

func init() {
	slog.SetLogLoggerLevel(slog.LevelWarn)
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	h, err := isbn.LoadHyphenator(context.Background(), http.DefaultClient)
	if err != nil {
		panic(err)
	}
	hyphenate := func(i isbn.ISBN) {
		if s := i.String(); strings.HasPrefix(s, "999") || (i.Is13() && strings.HasPrefix(s, "978999")) {
			fmt.Fprintln(out, s)
			return
		}
		if s, err := h.Hyphenate(i); err == nil {
			fmt.Fprintln(out, s)
		} else {
			slog.Warn("hyphenate failed", "error", err)
			fmt.Fprintln(out, i.String())
		}
	}
	count := 0
	for scanner.Scan() {
		count++
		if count%1000 == 0 {
			slog.Info("processed", "count", count)
		}
		line := scanner.Text()
		if i, err := try(line); err == nil {
			hyphenate(i)
			continue
		}
		parts := split(line)
		var failed []string
		for _, part := range parts {
			if i, err := try(part); err == nil {
				hyphenate(i)
				continue
			}
			failed = append(failed, part)
		}
		for _, part := range failed {
			for field := range strings.FieldsSeq(part) {
				if i, err := try(field); err == nil {
					hyphenate(i)
				} else {
					slog.Error("ISBN parsing failed", "error", err)
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
		os.Exit(1)
	}
}

func try(s string) (isbn.ISBN, error) {
	i, err := isbn.New(s)
	if err == nil || errors.Is(err, isbn.ErrInvalidCheckDigit) {
		return i, nil
	}
	i, err = isbn.ParseTagged(s)
	if err == nil || (!errors.Is(err, isbn.ErrInvalidKind) && errors.Is(err, isbn.ErrInvalidCheckDigit)) {
		return i, nil
	}
	return isbn.ISBN{}, err
}

func split(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == ';' || r == '/'
	})
}
