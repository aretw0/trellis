package session

import (
	"context"
	"fmt"
	"testing"

	"github.com/aretw0/trellis/pkg/domain"
)

// MockStore structure
type MockStore struct{}

func (m *MockStore) Save(ctx context.Context, sessionID string, state *domain.State) error {
	return nil
}
func (m *MockStore) Load(ctx context.Context, sessionID string) (*domain.State, error) {
	return nil, nil
}
func (m *MockStore) Delete(ctx context.Context, sessionID string) error { return nil }
func (m *MockStore) List(ctx context.Context) ([]string, error)         { return nil, nil }

func TestManager_LockLifecycle(t *testing.T) {
	mgr := NewManager(&MockStore{})
	ctx := context.Background()
	count := 10000

	// 1. Create and Delete many sessions
	for i := 0; i < count; i++ {
		sid := fmt.Sprintf("session-%d", i)
		_ = mgr.Save(ctx, sid, &domain.State{})
		_ = mgr.Delete(ctx, sid)
	}

	// 2. Count locks remaining in map
	lockCount := len(mgr.locks)

	// 3. Assert Leak
	// If cleaned up properly, count should be near 0.
	// Current behavior: It equals 'count' (ALL locks leak).
	t.Logf("Sessions Created: %d, Locks Leaked: %d", count, lockCount)

	if lockCount != 0 {
		t.Errorf("Memory Leak Detected: %d locks remaining in memory after Delete", lockCount)
	}
}
