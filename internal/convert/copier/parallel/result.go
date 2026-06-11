package parallel

import (
	"fmt"
	"log/slog"

	"github.com/DryHumour/readerware-to-tellico/internal/convert/copier"
	"github.com/DryHumour/readerware-to-tellico/internal/images"
)

type Report = copier.Report

// result represents the outcome of processing a single image file.
// It contains either a successful report or an error.
type result struct {
	Report Report
	Err    error
}

// progressResult creates a result with an informational progress message.
func progressResult(msg string) result {
	return result{
		Report: Report{
			Level:   slog.LevelInfo,
			Message: msg,
		},
	}
}

// fatalResult creates a result with a fatal error, wrapping the provided error with context text.
func fatalResult(text string, err error) result {
	return result{
		Err: fmt.Errorf("%s: %w", text, err),
	}
}

// fatalImageResult creates a result with a fatal error for an image, including the image ID in the error message.
func fatalImageResult(text string, entry *images.ManifestEntry, err error) result {
	return result{
		Err: fmt.Errorf("%s: id=%q: %w", text, entry.ID, err),
	}
}

// warnResult creates a non-fatal warning result for an image, including the image ID in the message.
// This is used for recoverable errors like format mismatch or oversized files.
func warnResult(text string, entry *images.ManifestEntry, err error) result {
	return result{
		Report: Report{
			Level:   slog.LevelWarn,
			Message: fmt.Sprintf("%s: id=%q", text, entry.ID),
			Err:     err,
		},
	}
}

// infoResult creates an informational result for an image, including the image ID in the message.
// This is used for informational messages like GIF conversion.
func infoResult(text string, entry *images.ManifestEntry, err error) result {
	return result{
		Report: Report{
			Level:   slog.LevelInfo,
			Message: fmt.Sprintf("%s: id=%q", text, entry.ID),
			Err:     err,
		},
	}
}
