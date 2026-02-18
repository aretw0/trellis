package ports

import (
	"context"

	"github.com/aretw0/trellis/pkg/domain"
)

// StatelessEngine defines the interface for state machine cores that do not maintain internal state.
// This is the primary interface used by adapters (e.g., HTTP, MCP) that manage state externally or per-request.
type StatelessEngine interface {
	// Render calculates the presentation (actions) for a given state without advancing it.
	Render(ctx context.Context, state *domain.State) ([]domain.ActionRequest, bool, error)

	// Navigate progresses the state machine based on input, returning the new state.
	Navigate(ctx context.Context, state *domain.State, input any) (*domain.State, error)

	// Signal triggers a global event on the state machine, potentially causing a transition.
	Signal(ctx context.Context, state *domain.State, signal string) (*domain.State, error)

	// Inspect returns the current graph structure for introspection.
	Inspect() ([]domain.Node, error)
}
