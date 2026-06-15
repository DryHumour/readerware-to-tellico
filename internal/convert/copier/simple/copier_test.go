package simple

import (
	"bytes"
	"context"
	"errors"
	"io"
	"iter"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DryHumour/readerware-to-tellico/internal/convert/copier"
	"github.com/DryHumour/readerware-to-tellico/internal/images"
	"github.com/DryHumour/readerware-to-tellico/internal/tellico/tcfile"
)

func TestSimpleCopier(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Valid JPEG magic numbers
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00}

	createImage := func(t *testing.T, name string, content []byte) string {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, content, 0644); err != nil {
			t.Fatal(err)
		}
		return path
	}

	t.Run("Copy - happy path", func(t *testing.T) {
		path := createImage(t, "happy.jpg", jpegData)
		info, _ := os.Stat(path)
		entry := &images.ManifestEntry{
			ID:     "happy",
			Path:   path,
			Format: "jpg",
			Info:   info,
		}

		var buf bytes.Buffer
		tcf := tcfile.New(&buf)
		c := New(logger, tcf)

		err := c.Copy(t.Context(), entry)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if err := tcf.Close(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Copy - empty file", func(t *testing.T) {
		path := createImage(t, "empty.jpg", []byte{})
		info, _ := os.Stat(path)
		entry := &images.ManifestEntry{
			ID:     "empty",
			Path:   path,
			Format: "jpg",
			Info:   info,
		}

		var buf bytes.Buffer
		tcf := tcfile.New(&buf)
		c := New(logger, tcf)

		err := c.Copy(t.Context(), entry)
		if !errors.Is(err, copier.ErrFileEmpty) {
			t.Errorf("expected ErrFileEmpty, got %v", err)
		}
	})

	t.Run("Copy - oversized file", func(t *testing.T) {
		data := make([]byte, copier.MaxReaderwareImageSize+100)
		copy(data, jpegData)
		path := createImage(t, "oversized.jpg", data)
		info, _ := os.Stat(path)
		entry := &images.ManifestEntry{
			ID:     "oversized",
			Path:   path,
			Format: "jpg",
			Info:   info,
		}

		var buf bytes.Buffer
		tcf := tcfile.New(&buf)
		c := New(logger, tcf)

		err := c.Copy(t.Context(), entry)
		if !errors.Is(err, copier.ErrFileTooLarge) {
			t.Errorf("expected ErrFileTooLarge, got %v", err)
		}
	})

	t.Run("Copy - format mismatch", func(t *testing.T) {
		path := createImage(t, "mismatch.jpg", jpegData)
		info, _ := os.Stat(path)
		entry := &images.ManifestEntry{
			ID:     "mismatch",
			Path:   path,
			Format: "png", // expect png, but file is jpg
			Info:   info,
		}

		var buf bytes.Buffer
		tcf := tcfile.New(&buf)
		c := New(logger, tcf)

		err := c.Copy(t.Context(), entry)
		if !errors.Is(err, copier.ErrFormatMismatch) {
			t.Errorf("expected ErrFormatMismatch, got %v", err)
		}
	})

	t.Run("CopyAll - mixed results", func(t *testing.T) {
		// 1. Success
		path1 := createImage(t, "ok.jpg", jpegData)
		info1, _ := os.Stat(path1)
		entry1 := &images.ManifestEntry{ID: "ok", Path: path1, Format: "jpg", Info: info1}

		// 2. Format mismatch (non-fatal)
		path2 := createImage(t, "mismatch.jpg", jpegData)
		info2, _ := os.Stat(path2)
		entry2 := &images.ManifestEntry{ID: "mismatch", Path: path2, Format: "png", Info: info2}

		// 3. File not found (fatal)
		entry3 := &images.ManifestEntry{ID: "missing", Path: "/no/such/file", Format: "jpg"}

		var buf bytes.Buffer
		tcf := tcfile.New(&buf)
		c := New(logger, tcf)

		entries := func(yield func(*images.ManifestEntry) bool) {
			if !yield(entry1) {
				return
			}
			if !yield(entry2) {
				return
			}
			if !yield(entry3) {
				return
			}
		}

		var reports []Report
		var finalErr error
		for report, err := range c.CopyAll(t.Context(), iter.Seq[*images.ManifestEntry](entries)) {
			if !report.IsEmpty() {
				reports = append(reports, report)
			}
			if err != nil {
				finalErr = err
				break
			}
		}

		if len(reports) != 5 {
			t.Errorf("expected 5 reports, got %d", len(reports))
		}
		if reports[0].Message != "copying "+path1 {
			t.Errorf("expected first report to be progress, got %q", reports[0].Message)
		}
		if reports[1].Level != slog.LevelInfo || !strings.Contains(reports[1].Message, "copied id=\"ok\"") {
			t.Errorf("expected Info level success report, got %v: %q", reports[1].Level, reports[1].Message)
		}
		if reports[2].Message != "copying "+path2 {
			t.Errorf("expected third report to be progress, got %q", reports[2].Message)
		}
		if reports[3].Level != slog.LevelWarn {
			t.Errorf("expected Warn level for fourth report, got %v", reports[3].Level)
		}
		if !errors.Is(reports[3].Err, copier.ErrFormatMismatch) {
			t.Errorf("expected ErrFormatMismatch in fourth report, got %v", reports[3].Err)
		}
		if reports[4].Message != "copying /no/such/file" {
			t.Errorf("expected fifth report to be progress, got %q", reports[4].Message)
		}
		if finalErr == nil {
			t.Error("expected fatal error for missing file, got nil")
		}
	})

	t.Run("CopyAll - context cancellation", func(t *testing.T) {
		path := createImage(t, "cancel.jpg", jpegData)
		info, _ := os.Stat(path)
		entry := &images.ManifestEntry{ID: "cancel", Path: path, Format: "jpg", Info: info}

		var buf bytes.Buffer
		tcf := tcfile.New(&buf)
		c := New(logger, tcf)

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		entries := func(yield func(*images.ManifestEntry) bool) {
			yield(entry)
		}

		for _, err := range c.CopyAll(ctx, iter.Seq[*images.ManifestEntry](entries)) {
			if err != nil && !errors.Is(err, context.Canceled) {
				t.Errorf("expected nil or context.Canceled, got %v", err)
			}
		}
	})
}
