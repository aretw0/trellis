package ports_test

import (
	"context"
	"testing"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

// MockStore is an in-memory implementation of StateStore for testing purposes.
type MockStore struct {
	data map[string]*domain.State
}

func NewMockStore() *MockStore {
	return &MockStore{
		data: make(map[string]*domain.State),
	}
}

func (m *MockStore) Save(ctx context.Context, sessionID string, state *domain.State) error {
	// Deep copy to simulate serialization
	copiedState := *state
	copiedState.Context = make(map[string]any)
	for k, v := range state.Context {
		copiedState.Context[k] = v
	}
	m.data[sessionID] = &copiedState
	return nil
}

func (m *MockStore) Load(ctx context.Context, sessionID string) (*domain.State, error) {
	state, ok := m.data[sessionID]
	if !ok {
		return nil, domain.ErrSessionNotFound
	}
	return state, nil
}

func (m *MockStore) Delete(ctx context.Context, sessionID string) error {
	delete(m.data, sessionID)
	return nil
}

func (m *MockStore) List(ctx context.Context) ([]string, error) {
	var sessions []string
	for id := range m.data {
		sessions = append(sessions, id)
	}
	return sessions, nil
}

func TestStateStore_Contract(t *testing.T) {
	// This test verifies that the MockStore complies with the StateStore logic.
	// It serves as a verification that our Contract Test itself is valid against a reference implementation.

	store := NewMockStore()
	ports.RunStateStoreContract(t, store)
}
