// Package copier provides image copy/convert helpers shared by the simple and
// parallel copier implementations, plus the GIF-to-PNG conversion used when
// writing Tellico archives.
package copier

import (
	"bytes"
	"context"
	"fmt"
	"image/gif"
	"image/png"
	"io"
	"io/fs"
	"iter"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/DryHumour/readerware-to-tellico/internal/images"
)

// Copier copies image files from manifest entries into a Tellico collection,
// yielding a report (and possibly an error) for each entry processed.
type Copier interface {
	CopyAll(ctx context.Context, entries iter.Seq[*images.ManifestEntry]) iter.Seq2[Report, error]
}

// Shared, thread-safe PNG encoder.
var pngEncoder = &png.Encoder{
	BufferPool:       &pngBufferPool{},
	CompressionLevel: png.DefaultCompression,
}

// pngBufferPool implements the png.EncoderBufferPool interface.
// It uses a sync.Pool to recycle the heavy zlib/deflate buffers.
type pngBufferPool struct {
	pool sync.Pool
}

// Get fetches a buffer from the pool.
// If the pool is empty, returning nil tells the PNG encoder to allocate a new one.
func (p *pngBufferPool) Get() *png.EncoderBuffer {
	if v := p.pool.Get(); v != nil {
		return v.(*png.EncoderBuffer)
	}
	return nil
}

// Put returns the buffer to the pool after encoding is finished.
func (p *pngBufferPool) Put(b *png.EncoderBuffer) {
	p.pool.Put(b)
}

// ConvertGIFToPNG streams a GIF from the reader, decodes the palette,
// and encodes it as a true-color PNG to the writer using recycled memory buffers.
func ConvertGIFToPNG(w io.Writer, r io.Reader) error {
	img, err := gif.Decode(r)
	if err != nil {
		return fmt.Errorf("failed to decode GIF stream: %w", err)
	}
	err = pngEncoder.Encode(w, img)
	if err != nil {
		return fmt.Errorf("failed to encode pooled PNG stream: %w", err)
	}
	return nil
}

type ConvertedPNG struct {
	bytes.Buffer
	name     string
	fileMode fs.FileMode
	modTime  time.Time
}

func NewConvertedPNG(fi fs.FileInfo) *ConvertedPNG {
	name := fi.Name()
	if ext := filepath.Ext(name); strings.EqualFold(ext, ".gif") {
		name = strings.TrimSuffix(name, ext)
	}
	name += ".png"
	return &ConvertedPNG{
		name:     name,
		fileMode: fi.Mode(),
		modTime:  time.Now(),
	}
}

func (p *ConvertedPNG) Stat() (fs.FileInfo, error) {
	return (*convertedPNGStatView)(p), nil
}

type convertedPNGStatView ConvertedPNG

func (p *convertedPNGStatView) Name() string {
	return p.name
}

func (p *convertedPNGStatView) Size() int64 {
	return int64(p.Len())
}

func (p *convertedPNGStatView) Mode() fs.FileMode {
	return p.fileMode
}

func (p *convertedPNGStatView) ModTime() time.Time {
	return p.modTime
}

func (p *convertedPNGStatView) IsDir() bool {
	return false
}

func (p *convertedPNGStatView) Sys() any {
	return nil
}
