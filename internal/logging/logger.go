package logging

import (
	"io"
	"log/slog"
	"os"
)

// New creates a configured application logger.
// It writes to Stderr (to separate from Stdout flow UI/JSON-RPC).
// It standardizes common keys (e.g., "error" -> "err").
func New(level slog.Level) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Standardize 'error' key to 'err'
			if a.Key == "error" {
				a.Key = "err"
			}
			return a
		},
	}))
}

// NewNop returns a no-op logger.
func NewNop() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
