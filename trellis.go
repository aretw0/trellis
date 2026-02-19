package trellis

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/aretw0/loam"
	"github.com/aretw0/trellis/internal/runtime"
	loamAdapter "github.com/aretw0/trellis/pkg/adapters/loam"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

// Engine is the high-level entry point for the Trellis library.
// It wraps the internal runtime and provides a simplified API for consumers.
type Engine struct {
	runtime            *runtime.Engine
	loader             ports.GraphLoader
	evaluator          runtime.ConditionEvaluator
	interpolator       runtime.Interpolator
	defaultErrorNodeID string
	runtimeOpts        []runtime.EngineOption
	hooks              domain.LifecycleHooks
	logger             *slog.Logger
	Name               string
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

// WithDefaultErrorNode sets a global fallback node for tool errors.
func WithDefaultErrorNode(nodeID string) Option {
	return func(e *Engine) {
		e.defaultErrorNodeID = nodeID
	}
}

// WithEntryNode configures the initial node ID (default: "start").
func WithEntryNode(nodeID string) Option {
	return func(e *Engine) {
		e.runtimeOpts = append(e.runtimeOpts, runtime.WithEntryNode(nodeID))
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

		eng.Name = filepath.Base(absPath)

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
		typedRepo := loam.NewTypedRepository[loamAdapter.NodeMetadata](repo)
		eng.loader = loamAdapter.New(typedRepo)
	} else {
		// If custom loader is provided, we can use repoPath as a descriptive label/session prefix.
		if repoPath != "" {
			eng.Name = filepath.Base(repoPath)
		}
	}

	// Ensure logger is initialized (so we don't pass nil to runtime, which would overwrite its default)
	if eng.logger == nil {
		eng.logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}

	// Enrich logger with graph name if available
	if eng.Name != "" {
		eng.logger = eng.logger.With("graph", eng.Name)
	}

	// Initialize Core Runtime with the selected loader
	runtimeOpts := []runtime.EngineOption{
		runtime.WithLifecycleHooks(eng.hooks),
		runtime.WithLogger(eng.logger),
		runtime.WithDefaultErrorNode(eng.defaultErrorNodeID),
	}
	// Append user-defined runtime options (like WithEntryNode)
	runtimeOpts = append(runtimeOpts, eng.runtimeOpts...)

	eng.runtime = runtime.NewEngine(
		eng.loader,
		eng.evaluator,
		eng.interpolator,
		runtimeOpts...,
	)

	return eng, nil
}

// Start creates the initial state for the flow and triggers lifecycle hooks.
func (e *Engine) Start(ctx context.Context, sessionID string, initialContext map[string]any) (*domain.State, error) {
	return e.runtime.Start(ctx, sessionID, initialContext)
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
