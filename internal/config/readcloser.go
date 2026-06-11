package config

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

var (
	// bom is the Unicode Byte Order Mark (BOM) for UTF-8 files.
	bom = []byte{0xEF, 0xBB, 0xBF}
)

type readCloser struct {
	io.Reader
	io.Closer
}

func newReadCloser(path string) (*readCloser, error) {
	// Open the underlying file.
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Skip the UTF-8 Byte Order Mark (BOM) if present.
	reader := bufio.NewReader(f)
	peek, err := reader.Peek(3)
	if err == nil && bytes.Equal(peek[:3], bom) {
		reader.Discard(3)
	} else if err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to peek file for BOM: %w", err)
	}

	return &readCloser{Reader: reader, Closer: f}, nil
}
