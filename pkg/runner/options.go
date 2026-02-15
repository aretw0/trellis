package runner

import (
	"log/slog"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

// DefaultInputBufferSize is the default number of lines to buffer for input handlers.
const DefaultInputBufferSize = 64

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

// WithRenderer configures the content renderer (e.g. TUI, Markdown).
func WithRenderer(renderer ContentRenderer) Option {
	return func(r *Runner) {
		r.Renderer = renderer
	}
}

// WithInterceptor configures the tool execution middleware.
func WithInterceptor(interceptor ToolInterceptor) Option {
	return func(r *Runner) {
		r.Interceptor = interceptor
	}
}

// WithToolRunner configures the strategy for executing side-effects.
func WithToolRunner(tr ToolRunner) Option {
	return func(r *Runner) {
		r.ToolRunner = tr
	}
}

// WithInterruptSource sets a channel that signals the runner to interrupt current execution.
func WithInterruptSource(ch <-chan struct{}) Option {
	return func(r *Runner) {
		r.InterruptSource = ch
	}
}

// WithEngine configures the Trellis engine for execution.
// This is required for the Runner to operate as a lifecycle.Worker.
func WithEngine(engine *trellis.Engine) Option {
	return func(r *Runner) {
		r.engine = engine
	}
}

// WithInitialState configures the initial state for the Runner.
// If not provided, the Runner will call Engine.Start() to create one.
func WithInitialState(state *domain.State) Option {
	return func(r *Runner) {
		r.initialState = state
	}
}
