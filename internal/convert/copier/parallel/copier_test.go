package parallel

import (
	"bytes"
	"context"
	"io"
	"iter"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/DryHumour/readerware-to-tellico/internal/images"
	"github.com/DryHumour/readerware-to-tellico/internal/tellico/tcfile"
)

func TestCopierCopyAll(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Create dummy image data
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00}
	path := filepath.Join(tmpDir, "test.jpg")
	if err := os.WriteFile(path, jpegData, 0644); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(path)

	entry := &images.ManifestEntry{
		ID:     "test.jpg",
		Path:   path,
		Format: "jpg",
		Info:   info,
	}

	t.Run("single file copy", func(t *testing.T) {
		var buf bytes.Buffer
		tcf := tcfile.New(&buf)
		c := New(logger, tcf, 2)

		entries := func(yield func(*images.ManifestEntry) bool) {
			yield(entry)
		}

		ctx := t.Context()
		for report, err := range c.CopyAll(ctx, iter.Seq[*images.ManifestEntry](entries)) {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			t.Logf("Report: %v", report.Message)
		}

		if err := tcf.Close(); err != nil {
			t.Fatal(err)
		}

		if buf.Len() == 0 {
			t.Error("expected data to be written to buffer")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		var buf bytes.Buffer
		tcf := tcfile.New(&buf)
		c := New(logger, tcf, 2)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		entries := func(yield func(*images.ManifestEntry) bool) {
			yield(entry)
		}

		// Should exit quickly without error (or with context error, but iterator handles it)
		for _, err := range c.CopyAll(ctx, iter.Seq[*images.ManifestEntry](entries)) {
			if err != nil && err != context.Canceled {
				t.Errorf("expected no error or context.Canceled, got %v", err)
			}
		}
	})
}
