package parallel

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/DryHumour/readerware-to-tellico/internal/convert/copier"
	"github.com/DryHumour/readerware-to-tellico/internal/images"
)

func TestReaderRun(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Helper to create a dummy image file
	createImage := func(t *testing.T, name string, content []byte) string {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, content, 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	// Valid JPEG magic numbers
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00}

	t.Run("happy path", func(t *testing.T) {
		path := createImage(t, "test.jpg", jpegData)
		info, _ := os.Stat(path)
		entry := &images.ManifestEntry{
			ID:     "123",
			Path:   path,
			Format: "jpg",
			Info:   info,
		}

		resultC := make(chan result, 10)
		writerC := make(chan *payload)
		readerC := make(chan *images.ManifestEntry)

		r := newReader(logger, resultC, writerC, readerC)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		go r.Run(ctx)
		readerC <- entry
		close(readerC)

		// Check results
		res1 := <-resultC // progress
		if res1.Report.Message != "reading "+path {
			t.Errorf("unexpected progress message: %q", res1.Report.Message)
		}

		p := <-writerC
		if p.ID != "123" {
			t.Errorf("expected ID 123, got %q", p.ID)
		}
		if string(p.Data) != string(jpegData) {
			t.Errorf("data mismatch")
		}
		releasePayload(p)
	})

	t.Run("oversized file", func(t *testing.T) {
		// Create data larger than copierpkg.MaxReaderwareImageSize
		largeData := make([]byte, copier.MaxReaderwareImageSize+100)
		copy(largeData, jpegData) // keep valid header
		path := createImage(t, "large.jpg", largeData)
		info, _ := os.Stat(path)
		entry := &images.ManifestEntry{
			ID:     "large",
			Path:   path,
			Format: "jpg",
			Info:   info,
		}

		resultC := make(chan result, 10)
		writerC := make(chan *payload)
		readerC := make(chan *images.ManifestEntry)

		r := newReader(logger, resultC, writerC, readerC)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		go r.Run(ctx)
		readerC <- entry
		close(readerC)

		<-resultC         // progress
		res2 := <-resultC // warning about oversized
		if res2.Report.Level != slog.LevelWarn || res2.Report.Err == nil {
			t.Errorf("expected warning result, got %+v", res2)
		}

		p := <-writerC
		if len(p.Data) != len(largeData) {
			t.Errorf("expected size %d, got %d", len(largeData), len(p.Data))
		}
		releasePayload(p)
	})

	t.Run("format mismatch", func(t *testing.T) {
		path := createImage(t, "mismatch.jpg", jpegData)
		info, _ := os.Stat(path)
		entry := &images.ManifestEntry{
			ID:     "mismatch",
			Path:   path,
			Format: "png", // expecting PNG, but file is JPEG
			Info:   info,
		}

		resultC := make(chan result, 10)
		writerC := make(chan *payload)
		readerC := make(chan *images.ManifestEntry)

		r := newReader(logger, resultC, writerC, readerC)

		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		go r.Run(ctx)
		readerC <- entry
		close(readerC)

		<-resultC         // progress
		res2 := <-resultC // warning about mismatch
		if res2.Report.Level != slog.LevelWarn || res2.Report.Err == nil {
			t.Errorf("expected warning result, got %+v", res2)
		}

		p := <-writerC
		releasePayload(p)
	})

	t.Run("context cancellation", func(t *testing.T) {
		resultC := make(chan result)
		writerC := make(chan *payload)
		readerC := make(chan *images.ManifestEntry)

		r := newReader(logger, resultC, writerC, readerC)

		ctx, cancel := context.WithCancel(t.Context())
		cancel() // cancel immediately

		r.Run(ctx) // should return immediately
	})
}
