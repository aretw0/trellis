package trellis

import (
	"fmt"
	"path/filepath"

	"github.com/aretw0/loam"
	"github.com/aretw0/trellis/internal/adapters"
	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/domain"
)

// Engine is the high-level entry point for the Trellis library.
// It wraps the internal runtime and provides a simplified API for consumers.
type Engine struct {
	runtime *runtime.Engine
}

// New initializes a new Trellis Engine backed by a Loam repository at the given path.
// It sets up the necessary adapters and loads the content.
func New(repoPath string) (*Engine, error) {
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	// Initialize Loam in read-only mode (Game Mode)
	// We explicitly disable versioning side-effects for the runtime.
	repo, err := loam.Init(absPath, loam.WithVersioning(false))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize loam: %w", err)
	}

	// Setup Typed Repository and Adapter
	typedRepo := loam.NewTypedRepository[adapters.NodeMetadata](repo)
	loader := adapters.NewLoamLoader(typedRepo)

	// Initialize Core Runtime
	rt := runtime.NewEngine(loader)

	return &Engine{
		runtime: rt,
	}, nil
}

// Start creates the initial state for the flow.
// It acts as a factory for the first generic State.
func (e *Engine) Start() *domain.State {
	return domain.NewState("start")
}

// Step executes a single transition step in the flow based on the input.
func (e *Engine) Step(state *domain.State, input string) ([]domain.ActionRequest, *domain.State, error) {
	return e.runtime.Step(state, input)
}
