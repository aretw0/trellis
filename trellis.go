package trellis

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/aretw0/loam"
	"github.com/aretw0/trellis/internal/adapters"
	"github.com/aretw0/trellis/internal/dto"
	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

// Engine is the high-level entry point for the Trellis library.
// It wraps the internal runtime and provides a simplified API for consumers.
type Engine struct {
	runtime      *runtime.Engine
	loader       ports.GraphLoader
	evaluator    runtime.ConditionEvaluator
	interpolator runtime.Interpolator
	hooks        domain.LifecycleHooks
	logger       *slog.Logger
}

// Option defines a functional option for configuring the Engine.
type Option func(*Engine)

// WithLifecycleHooks registers observability hooks.
func WithLifecycleHooks(hooks domain.LifecycleHooks) Option {
	return func(e *Engine) {
		e.hooks = hooks
	}
}

// WithLoader injects a custom GraphLoader, bypassing the default Loam initialization.
func WithLoader(l ports.GraphLoader) Option {
	return func(e *Engine) {
		e.loader = l
	}
}

// WithConditionEvaluator sets a custom logic evaluator for the engine.
func WithConditionEvaluator(eval runtime.ConditionEvaluator) Option {
	return func(e *Engine) {
		e.evaluator = eval
	}
}

// WithInterpolator sets a custom interpolator for the engine.
func WithInterpolator(interp runtime.Interpolator) Option {
	return func(e *Engine) {
		e.interpolator = interp
	}
}

// WithLogger sets a custom structured logger for the engine.
func WithLogger(logger *slog.Logger) Option {
	return func(e *Engine) {
		e.logger = logger
	}
}

// New initializes a new Trellis Engine.
// By default, it uses a Loam repository at the given path.
// If WithLoader option is provided, repoPath can be empty and Loam is skipped.
func New(repoPath string, opts ...Option) (*Engine, error) {
	eng := &Engine{}

	// Apply Options first to check if a loader is provided
	for _, opt := range opts {
		opt(eng)
	}

	// If no loader was injected, initialize default Loam adapter
	if eng.loader == nil {
		if repoPath == "" {
			return nil, fmt.Errorf("repoPath is required when no custom loader is provided")
		}

		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			return nil, fmt.Errorf("invalid path: %w", err)
		}

		// Initialize Loam with global strict mode (v0.10.4+)
		// This ensures that all adapters (JSON, Markdown/YAML) return consistent numeric types (json.Number),
		// preventing "float64" ambiguity for large integers.
		// We also enforce ReadOnly mode (v0.10.6+) to avoid Loam's "sandbox" behavior in dev mode.
		// The Engine never modifies the graph structure, only reads it.
		repo, err := loam.Init(absPath,
			loam.WithStrict(true),
			loam.WithReadOnly(true),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize loam: %w", err)
		}

		// Setup Typed Repository and Adapter
		typedRepo := loam.NewTypedRepository[dto.NodeMetadata](repo)
		eng.loader = adapters.NewLoamLoader(typedRepo)
	}

	// Initialize Core Runtime with the selected loader
	eng.runtime = runtime.NewEngine(
		eng.loader,
		eng.evaluator,
		eng.interpolator,
		runtime.WithLifecycleHooks(eng.hooks),
		runtime.WithLogger(eng.logger),
	)

	return eng, nil
}

// Start creates the initial state for the flow and triggers lifecycle hooks.
func (e *Engine) Start(ctx context.Context, initialContext map[string]any) (*domain.State, error) {
	return e.runtime.Start(ctx, initialContext)
}

// Render generates the actions (view) for the current state without transitioning.
// Returns actions, isTerminal (true if no transitions), and error.
func (e *Engine) Render(ctx context.Context, state *domain.State) ([]domain.ActionRequest, bool, error) {
	return e.runtime.Render(ctx, state)
}

// Navigate determines the next state based on input.
func (e *Engine) Navigate(ctx context.Context, state *domain.State, input any) (*domain.State, error) {
	return e.runtime.Navigate(ctx, state, input)
}

// Signal triggers a state transition based on a global signal (e.g. interrupt).
func (e *Engine) Signal(ctx context.Context, state *domain.State, signalName string) (*domain.State, error) {
	return e.runtime.Signal(ctx, state, signalName)
}

// Inspect returns the full graph definition for visualization or introspection tools.
func (e *Engine) Inspect() ([]domain.Node, error) {
	return e.runtime.Inspect()
}

// Watch returns a channel that signals when the underlying graph changes.
// Returns error if the loader does not support watching.
func (e *Engine) Watch(ctx context.Context) (<-chan string, error) {
	if w, ok := e.loader.(ports.Watchable); ok {
		return w.Watch(ctx)
	}
	return nil, fmt.Errorf("current loader does not support watching")
}

// Loader returns the underlying GraphLoader used by the engine.
func (e *Engine) Loader() ports.GraphLoader {
	return e.loader
}
