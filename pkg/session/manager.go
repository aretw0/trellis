package session

import (
	"context"
	"fmt"
	"sync"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

// Manager orchestrates session access, ensuring safe concurrent operations.
type Manager struct {
	store ports.StateStore
	locks sync.Map // map[string]*sync.Mutex
}

// NewManager creates a new Session Manager with the given persistence store.
func NewManager(store ports.StateStore) *Manager {
	return &Manager{
		store: store,
	}
}

// getLock returns or creates a mutex for the given session ID.
func (m *Manager) getLock(sessionID string) *sync.Mutex {
	lock, _ := m.locks.LoadOrStore(sessionID, &sync.Mutex{})
	return lock.(*sync.Mutex)
}

// Load retrieves an existing session from the store.
// It acquires a lock on the session ID during the operation.
func (m *Manager) Load(ctx context.Context, sessionID string) (*domain.State, error) {
	mu := m.getLock(sessionID)
	mu.Lock()
	defer mu.Unlock()

	return m.store.Load(ctx, sessionID)
}

// LoadOrStart tries to load a session. If not found, it initializes a new one.
// This is atomic: two concurrent calls for the same ID will result in one creation.
func (m *Manager) LoadOrStart(ctx context.Context, sessionID string, startNode string) (*domain.State, error) {
	mu := m.getLock(sessionID)
	mu.Lock()
	defer mu.Unlock()

	state, err := m.store.Load(ctx, sessionID)
	if err == nil {
		return state, nil
	}

	if err != domain.ErrSessionNotFound {
		return nil, fmt.Errorf("failed to check session existence: %w", err)
	}

	// Not found, create new
	if startNode == "" {
		startNode = "start"
	}
	state = domain.NewState(startNode)

	// Persist immediately to reserve the ID
	if err := m.store.Save(ctx, sessionID, state); err != nil {
		return nil, fmt.Errorf("failed to initialize session: %w", err)
	}

	return state, nil
}

// Save persists the session state.
// It acquires a lock to ensure no other process is modifying the session.
func (m *Manager) Save(ctx context.Context, sessionID string, state *domain.State) error {
	mu := m.getLock(sessionID)
	mu.Lock()
	defer mu.Unlock()

	return m.store.Save(ctx, sessionID, state)
}

// Delete removes the session from the store and clears its lock.
func (m *Manager) Delete(ctx context.Context, sessionID string) error {
	mu := m.getLock(sessionID)
	mu.Lock()
	defer mu.Unlock()

	err := m.store.Delete(ctx, sessionID)

	// We keep the lock in the map to valid race, or we could delete it using m.locks.Delete(sessionID)
	// But safely deleting the lock itself while other goroutines might obtain it is tricky.
	// For now, leaking a few mutexes for deleted sessions is acceptable vs complex GC logic.
	// In long running process we might want an LRU or explicit cleanup.

	return err
}

// List delegates to the store. Consistency is eventual as we don't lock the world.
func (m *Manager) List(ctx context.Context) ([]string, error) {
	return m.store.List(ctx)
}

// WithLock executes a function while holding the lock for the session.
// Useful for complex operations that need atomicity (Read-Modify-Write).
func (m *Manager) WithLock(ctx context.Context, sessionID string, fn func(context.Context) error) error {
	mu := m.getLock(sessionID)

	// Option: Implement context-aware locking if we want to support timeouts on lock acquisition.
	// For now, standard mutex.
	mu.Lock()
	defer mu.Unlock()

	return fn(ctx)
}
