package session_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/session"
	"github.com/stretchr/testify/assert"
)

// SlowStore simulates latency to provoke rac conditions if locking is missing.
type SlowStore struct {
	data map[string]*domain.State
	mu   sync.Mutex
}

func (s *SlowStore) Save(ctx context.Context, sessionID string, state *domain.State) error {
	time.Sleep(10 * time.Millisecond) // Simulate IO
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.data == nil {
		s.data = make(map[string]*domain.State)
	}
	s.data[sessionID] = state
	return nil
}

func (s *SlowStore) Load(ctx context.Context, sessionID string) (*domain.State, error) {
	time.Sleep(10 * time.Millisecond) // Simulate IO
	s.mu.Lock()
	defer s.mu.Unlock()

	if state, ok := s.data[sessionID]; ok {
		return state, nil
	}
	return nil, domain.ErrSessionNotFound
}

func (s *SlowStore) Delete(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, sessionID)
	return nil
}

func (s *SlowStore) List(ctx context.Context) ([]string, error) {
	return nil, nil
}

func TestManager_Locking(t *testing.T) {
	store := &SlowStore{}
	manager := session.NewManager(store)
	ctx := context.Background()
	id := "race-test"

	// Initial save
	_ = manager.Save(ctx, id, domain.NewState(id, "start"))

	var wg sync.WaitGroup
	concurrentWrites := 10

	// We want to verify that writes are serialized.
	// In a real scenario, Read-Modify-Write without locking would lose updates.
	// Here we just ensure no panics or weird state, but the manager.WithLock is where the value lies.

	for i := 0; i < concurrentWrites; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()

			// Just call Save. The Manager must ensure this is safe.
			// The SlowStore simulates IO delay.
			// If locking works, these should happen sequentially (or at least safely).
			err := manager.Save(ctx, id, domain.NewState(id, "updated"))
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()
}

func TestManager_LoadOrStart(t *testing.T) {
	// Verify atomic creation
	store := &SlowStore{}
	manager := session.NewManager(store)
	ctx := context.Background()
	id := "atomic-init"

	var wg sync.WaitGroup
	// Launch 2 routines trying to init same session
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			state, err := manager.LoadOrStart(ctx, id, "start")
			assert.NoError(t, err)
			assert.NotNil(t, state)
		}()
	}
	wg.Wait()

	// Should exist and be valid
	state, err := manager.Load(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, "start", state.CurrentNodeID)
}
