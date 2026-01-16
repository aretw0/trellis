package ports

import (
	"context"

	"github.com/aretw0/trellis/pkg/domain"
)

// StateStore defines the interface for persisting execution state.
// This allows for durable execution, enabling "Stop & Resume" workflows.
type StateStore interface {
	// Save persists the state for a given session ID.
	Save(ctx context.Context, sessionID string, state *domain.State) error

	// Load retrieves the state for a given session ID.
	// Returns domain.ErrSessionNotFound if the session does not exist.
	Load(ctx context.Context, sessionID string) (*domain.State, error)

	// Delete removes the state for a given session ID.
	Delete(ctx context.Context, sessionID string) error
}
