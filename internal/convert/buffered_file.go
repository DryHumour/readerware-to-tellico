package convert

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)

var (
	// byteOrderMarkUTF8 is the UTF-8 encoded Byte Order Mark (BOM).
	byteOrderMarkUTF8 = []byte{0xEF, 0xBB, 0xBF}
)

type bufferedFile struct {
	*bufio.Reader
	File *os.File
}

func newBufferedFile(path string) (*bufferedFile, error) {
	// Open the underlying file.
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	// Skip the UTF-8 Byte Order Mark (BOM) if present.
	reader := bufio.NewReader(f)
	peek, err := reader.Peek(3)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to peek file for BOM: %w", err)
	}
	if bytes.Equal(peek[:3], byteOrderMarkUTF8) {
		reader.Discard(3)
	}

	return &bufferedFile{Reader: reader, File: f}, nil
}

func (b *bufferedFile) Name() string {
	return b.File.Name()
}

func (b *bufferedFile) Stat() (os.FileInfo, error) {
	return b.File.Stat()
}

func (b *bufferedFile) Close() error {
	return b.File.Close()
}
