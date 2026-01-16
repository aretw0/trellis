package runner

import (
	"context"
	"fmt"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

// SessionManager handles the lifecycle of a durable session.
// It coordinates between the Runner, the Engine, and the StateStore.
type SessionManager struct {
	Store ports.StateStore
}

// NewSessionManager creates a new SessionManager.
func NewSessionManager(store ports.StateStore) *SessionManager {
	return &SessionManager{
		Store: store,
	}
}

// LoadOrStart attempts to load an existing session. If not found, it starts a new one.
// Returns the state and a boolean indicating if it was loaded (true) or new (false).
func (sm *SessionManager) LoadOrStart(
	ctx context.Context,
	engine *trellis.Engine,
	sessionID string,
	initialContext map[string]any,
) (*domain.State, bool, error) {
	if sessionID == "" {
		// Ephemeral session (no ID), just start new
		state, err := engine.Start(ctx, initialContext)
		return state, false, err
	}

	// Try Load
	state, err := sm.Store.Load(ctx, sessionID)
	if err == nil {
		// Resume: We do NOT apply initialContext on resume to avoid overwriting progress
		return state, true, nil
	}

	if err != domain.ErrSessionNotFound {
		return nil, false, fmt.Errorf("failed to load session %s: %w", sessionID, err)
	}

	// Not found, Start New
	state, err = engine.Start(ctx, initialContext)
	if err != nil {
		return nil, false, err
	}

	// Save immediately to reserve the ID
	if err := sm.Store.Save(ctx, sessionID, state); err != nil {
		return nil, false, fmt.Errorf("failed to initialize session %s: %w", sessionID, err)
	}

	return state, false, nil
}

// Save persists the state.
func (sm *SessionManager) Save(ctx context.Context, sessionID string, state *domain.State) error {
	if sessionID == "" || sm.Store == nil {
		return nil
	}
	return sm.Store.Save(ctx, sessionID, state)
}
