package runner

import (
	"log/slog"

	"github.com/aretw0/trellis/pkg/ports"
)

// Option defines a functional option for configuring the Runner.
type Option func(*Runner)

// WithStore configures the StateStore for persistence.
func WithStore(store ports.StateStore) Option {
	return func(r *Runner) {
		r.Store = store
	}
}

// WithLogger configures the structured logger.
func WithLogger(logger *slog.Logger) Option {
	return func(r *Runner) {
		r.Logger = logger
	}
}

// WithInputHandler configures a custom IOHandler.
func WithInputHandler(handler IOHandler) Option {
	return func(r *Runner) {
		r.Handler = handler
	}
}

// WithHeadless sets the runner to headless mode.
func WithHeadless(headless bool) Option {
	return func(r *Runner) {
		r.Headless = headless
	}
}

// WithSessionID sets the session ID for persistence context.
// This is required if WithStore is used.
func WithSessionID(id string) Option {
	return func(r *Runner) {
		r.SessionID = id
	}
}
