package ports_test

import (
	"context"
	"testing"

	"github.com/aretw0/trellis/pkg/domain"
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

func TestStateStore_Contract(t *testing.T) {
	// This test verifies that the MockStore complies with the StateStore logic.
	// It serves as a contract test for future implementations (Adapters).

	ctx := context.Background()
	store := NewMockStore()
	sessionID := "test-session"

	// 1. Load non-existent session
	_, err := store.Load(ctx, sessionID)
	if err != domain.ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got %v", err)
	}

	// 2. Save session
	state := domain.NewState("start")
	state.Context["foo"] = "bar"
	err = store.Save(ctx, sessionID, state)
	if err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// 3. Load session
	loaded, err := store.Load(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}
	if loaded.CurrentNodeID != state.CurrentNodeID {
		t.Errorf("Expected NodeID %s, got %s", state.CurrentNodeID, loaded.CurrentNodeID)
	}
	if loaded.Context["foo"] != "bar" {
		t.Errorf("Expected Context['foo'] = 'bar', got %v", loaded.Context["foo"])
	}

	// 4. Delete session
	err = store.Delete(ctx, sessionID)
	if err != nil {
		t.Fatalf("Failed to delete session: %v", err)
	}

	// 5. Load deleted session
	_, err = store.Load(ctx, sessionID)
	if err != domain.ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound after delete, got %v", err)
	}
}
