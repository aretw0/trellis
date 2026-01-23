package memory

import (
	"context"
	"sync"

	"github.com/aretw0/trellis/pkg/domain"
)

// Store implements ports.StateStore in memory.
// Safe for concurrent use.
type Store struct {
	data map[string]*domain.State
	mu   sync.RWMutex
}

// NewStore creates a new in-memory store.
func NewStore() *Store {
	return &Store{
		data: make(map[string]*domain.State),
	}
}

// Save persists the state in memory.
func (s *Store) Save(ctx context.Context, sessionID string, state *domain.State) error {
	// Deep copy to ensure isolation, similar to serialization
	copiedState := *state
	copiedState.Context = make(map[string]any)
	for k, v := range state.Context {
		copiedState.Context[k] = v
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[sessionID] = &copiedState
	return nil
}

// Load retrieves the state from memory.
func (s *Store) Load(ctx context.Context, sessionID string) (*domain.State, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.data[sessionID]
	if !ok {
		return nil, domain.ErrSessionNotFound
	}

	// Create a copy on read so caller can't mutate store state directly by pointer
	ret := *state
	ret.Context = make(map[string]any)
	for k, v := range state.Context {
		ret.Context[k] = v
	}

	return &ret, nil
}

// Delete removes the state.
func (s *Store) Delete(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, sessionID)
	return nil
}

// List returns active sessions.
func (s *Store) List(ctx context.Context) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]string, 0, len(s.data))
	for id := range s.data {
		sessions = append(sessions, id)
	}
	return sessions, nil
}
