package copier

import (
	"context"
	"log/slog"
)

// Report represents the outcome of a single pipeline step.
// It carries a log level, a human-readable message, and an optional error.
type Report struct {
	Level   slog.Level
	Message string
	Err     error
}

// IsEmpty tests whether the report has no message and no error.
func (r Report) IsEmpty() bool {
	return r.Message == "" && r.Err == nil
}

// Log writes the report to the given logger at the appropriate level.
func (r Report) Log(ctx context.Context, logger *slog.Logger) {
	if r.IsEmpty() {
		return
	}
	if logger == nil {
		logger = slog.Default()
	}
	var attrs []any
	if r.Err != nil {
		attrs = append(attrs, "error", r.Err)
	}
	logger.Log(ctx, r.Level, r.Message, attrs...)
}
