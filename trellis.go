package trellis

import (
	"context"
	"fmt"
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
	runtime   *runtime.Engine
	loader    ports.GraphLoader
	evaluator runtime.ConditionEvaluator
}

// Option defines a functional option for configuring the Engine.
type Option func(*Engine)

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
	eng.runtime = runtime.NewEngine(eng.loader, eng.evaluator)

	return eng, nil
}

// Start creates the initial state for the flow.
// It acts as a factory for the first generic State.
func (e *Engine) Start() *domain.State {
	return domain.NewState("start")
}

// Render generates the actions (view) for the current state without transitioning.
// Returns actions, isTerminal (true if no transitions), and error.
func (e *Engine) Render(ctx context.Context, state *domain.State) ([]domain.ActionRequest, bool, error) {
	return e.runtime.Render(ctx, state)
}

// Navigate calculates the next state based on the current state and input.
func (e *Engine) Navigate(ctx context.Context, state *domain.State, input string) (*domain.State, error) {
	return e.runtime.Navigate(ctx, state, input)
}

// Inspect returns the full graph definition for visualization or introspection tools.
func (e *Engine) Inspect() ([]domain.Node, error) {
	return e.runtime.Inspect()
}

// Watch returns a channel that signals when the underlying graph changes.
// Returns error if the loader does not support watching.
func (e *Engine) Watch(ctx context.Context) (<-chan struct{}, error) {
	if w, ok := e.loader.(ports.Watchable); ok {
		return w.Watch(ctx)
	}
	return nil, fmt.Errorf("current loader does not support watching")
}

// Loader returns the underlying GraphLoader used by the engine.
func (e *Engine) Loader() ports.GraphLoader {
	return e.loader
}
