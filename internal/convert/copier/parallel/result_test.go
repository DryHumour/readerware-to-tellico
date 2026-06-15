package parallel

import (
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/DryHumour/readerware-to-tellico/internal/convert/copier"
	"github.com/DryHumour/readerware-to-tellico/internal/images"
)

func TestResultHelpers(t *testing.T) {
	t.Parallel()

	entry := &images.ManifestEntry{
		ID:     "123",
		Path:   "/path/to/image.jpg",
		Format: "jpeg",
	}

	t.Run("progressResult", func(t *testing.T) {
		res := progressResult("doing something")
		if res.Report.Level != slog.LevelInfo {
			t.Errorf("expected level Info, got %v", res.Report.Level)
		}
		if res.Report.Message != "doing something" {
			t.Errorf("expected message 'doing something', got %q", res.Report.Message)
		}
	})

	t.Run("fatalResult", func(t *testing.T) {
		err := errors.New("base error")
		res := fatalResult("something failed", err)
		if res.Err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(res.Err.Error(), "something failed") {
			t.Errorf("expected context in error, got %q", res.Err.Error())
		}
		if !errors.Is(res.Err, err) {
			t.Errorf("expected wrapped error, got %v", res.Err)
		}
	})

	t.Run("fatalImageResult", func(t *testing.T) {
		err := errors.New("base error")
		res := fatalImageResult("image failed", entry, err)
		if !strings.Contains(res.Err.Error(), "id=\"123\"") {
			t.Errorf("expected ID in error message, got %q", res.Err.Error())
		}
	})

	t.Run("warnResult", func(t *testing.T) {
		err := errors.New("warning context")
		res := warnResult("something suspicious", entry, err)
		if res.Report.Level != slog.LevelWarn {
			t.Errorf("expected level Warn, got %v", res.Report.Level)
		}
		if !strings.Contains(res.Report.Message, "id=\"123\"") {
			t.Errorf("expected ID in message, got %q", res.Report.Message)
		}
		if res.Report.Err != err {
			t.Errorf("expected error in report, got %v", res.Report.Err)
		}
	})
}

func TestErrorHelpers(t *testing.T) {
	t.Parallel()

	entry := &images.ManifestEntry{
		ID:     "123",
		Path:   "/path/to/image.jpg",
		Format: "jpeg",
	}

	t.Run("newFileEmptyError", func(t *testing.T) {
		err := copier.NewFileEmptyError(entry)
		if !errors.Is(err, copier.ErrFileEmpty) {
			t.Errorf("expected ErrFileEmpty, got %v", err)
		}
		if !errors.Is(err, io.EOF) {
			t.Errorf("expected io.EOF, got %v", err)
		}
		if !strings.Contains(err.Error(), entry.Path) {
			t.Errorf("expected path in error, got %q", err.Error())
		}
	})

	t.Run("newFileError", func(t *testing.T) {
		baseErr := errors.New("io error")
		err := copier.NewFileError("failed to open", entry, baseErr)
		if !errors.Is(err, baseErr) {
			t.Errorf("expected wrapped error, got %v", err)
		}
		if !strings.Contains(err.Error(), entry.Path) {
			t.Errorf("expected path in error, got %q", err.Error())
		}
	})

	t.Run("newFileTooLargeError", func(t *testing.T) {
		err := copier.NewFileTooLargeError(entry, 1000000)
		if !errors.Is(err, copier.ErrFileTooLarge) {
			t.Errorf("expected ErrFileTooLarge, got %v", err)
		}
		if !strings.Contains(err.Error(), "1000000") {
			t.Errorf("expected size in error, got %q", err.Error())
		}
	})

	t.Run("newFormatMismatchError", func(t *testing.T) {
		err := copier.NewFormatMismatchError(entry, "png")
		if !errors.Is(err, copier.ErrFormatMismatch) {
			t.Errorf("expected ErrFormatMismatch, got %v", err)
		}
		if !strings.Contains(err.Error(), "expected jpeg, got png") {
			t.Errorf("expected format info in error, got %q", err.Error())
		}
	})
}
