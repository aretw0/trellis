package middleware_test

import (
	"context"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

// MockStore is a simple map-based store for testing middleware.
type MockStore struct {
	data map[string]*domain.State
}

func NewMockStore() *MockStore {
	return &MockStore{
		data: make(map[string]*domain.State),
	}
}

func (s *MockStore) Save(ctx context.Context, sessionID string, state *domain.State) error {
	s.data[sessionID] = state
	return nil
}

func (s *MockStore) Load(ctx context.Context, sessionID string) (*domain.State, error) {
	state, ok := s.data[sessionID]
	if !ok {
		return nil, domain.ErrSessionNotFound
	}
	return state, nil
}

func (s *MockStore) Delete(ctx context.Context, sessionID string) error {
	delete(s.data, sessionID)
	return nil
}

func (s *MockStore) List(ctx context.Context) ([]string, error) {
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys, nil
}

var _ ports.StateStore = (*MockStore)(nil)
