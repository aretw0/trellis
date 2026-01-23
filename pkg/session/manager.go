package session

import (
	"context"
	"fmt"
	"sync"

	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

// lockEntry holds the mutex and the reference count.
type lockEntry struct {
	mu   sync.Mutex
	refs int
}

// Manager orchestrates session access, ensuring safe concurrent operations.
// It uses Reference Counting to garbage collect unused locks.
type Manager struct {
	store ports.StateStore

	mu    sync.Mutex            // Global lock for the map
	locks map[string]*lockEntry // Map of active locks
}

// NewManager creates a new Session Manager with the given persistence store.
func NewManager(store ports.StateStore) *Manager {
	return &Manager{
		store: store,
		locks: make(map[string]*lockEntry),
	}
}

// acquire gets or creates a lock entry and increments its reference count.
// It returns the entry, which MUST be released via release(sessionID) after use.
func (m *Manager) acquire(sessionID string) *lockEntry {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.locks[sessionID]
	if !exists {
		entry = &lockEntry{}
		m.locks[sessionID] = entry
	}
	entry.refs++
	return entry
}

// release decrements the reference count and deletes the entry if it reaches zero.
func (m *Manager) release(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.locks[sessionID]
	if !exists {
		return // Should not happen if paired correctly
	}

	entry.refs--
	if entry.refs <= 0 {
		delete(m.locks, sessionID)
	}
}

// Load retrieves an existing session from the store.
func (m *Manager) Load(ctx context.Context, sessionID string) (*domain.State, error) {
	entry := m.acquire(sessionID)
	entry.mu.Lock()
	defer func() {
		entry.mu.Unlock()
		m.release(sessionID)
	}()

	return m.store.Load(ctx, sessionID)
}

// LoadOrStart tries to load a session. If not found, it initializes a new one.
func (m *Manager) LoadOrStart(ctx context.Context, sessionID string, startNode string) (*domain.State, error) {
	entry := m.acquire(sessionID)
	entry.mu.Lock()
	defer func() {
		entry.mu.Unlock()
		m.release(sessionID)
	}()

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
func (m *Manager) Save(ctx context.Context, sessionID string, state *domain.State) error {
	entry := m.acquire(sessionID)
	entry.mu.Lock()
	defer func() {
		entry.mu.Unlock()
		m.release(sessionID)
	}()

	return m.store.Save(ctx, sessionID, state)
}

// Delete removes the session from the store.
func (m *Manager) Delete(ctx context.Context, sessionID string) error {
	entry := m.acquire(sessionID)
	entry.mu.Lock()
	defer func() {
		entry.mu.Unlock()
		m.release(sessionID)
	}()

	return m.store.Delete(ctx, sessionID)
}

// List delegates to the store.
func (m *Manager) List(ctx context.Context) ([]string, error) {
	return m.store.List(ctx)
}

// WithLock executes a function while holding the lock for the session.
func (m *Manager) WithLock(ctx context.Context, sessionID string, fn func(context.Context) error) error {
	entry := m.acquire(sessionID)
	entry.mu.Lock()
	defer func() {
		entry.mu.Unlock()
		m.release(sessionID)
	}()

	return fn(ctx)
}
