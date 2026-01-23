package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"log/slog"

	"github.com/aretw0/trellis/internal/logging"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

// lockEntry holds the mutex and the reference count.
type lockEntry struct {
	mu     sync.Mutex
	refs   int
	unlock ports.UnlockFunc // Function to release distributed lock (if any)
}

// Manager orchestrates session access, ensuring safe concurrent operations.
// It uses Reference Counting to garbage collect unused locks.
type Manager struct {
	store ports.StateStore

	mu    sync.Mutex            // Global lock for the map
	locks map[string]*lockEntry // Map of active locks

	locker ports.DistributedLocker // Optional distributed locker
	logger *slog.Logger            // Logger for internal events (like deferred errors)
}

// Option configures the Manager.
type Option func(*Manager)

// WithLocker enables distributed locking.
func WithLocker(locker ports.DistributedLocker) Option {
	return func(m *Manager) {
		m.locker = locker
	}
}

// WithLogger configures a logger for the Manager.
func WithLogger(logger *slog.Logger) Option {
	return func(m *Manager) {
		m.logger = logger
	}
}

// NewManager creates a new Session Manager with the given persistence store.
func NewManager(store ports.StateStore, opts ...Option) *Manager {
	m := &Manager{
		store:  store,
		locks:  make(map[string]*lockEntry),
		logger: logging.NewNop(), // Default to no-op
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// acquire gets or creates a lock entry and increments its reference count.
// The caller MUST Lock the entry.mu, and then call release(sessionID) after unlocking.
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
	var state *domain.State
	err := m.WithLock(ctx, sessionID, func(ctx context.Context) error {
		var err error
		state, err = m.store.Load(ctx, sessionID)
		return err
	})
	return state, err
}

// LoadOrStart tries to load a session. If not found, it initializes a new one.
func (m *Manager) LoadOrStart(ctx context.Context, sessionID string, startNode string) (*domain.State, error) {
	var state *domain.State
	err := m.WithLock(ctx, sessionID, func(ctx context.Context) error {
		var err error
		state, err = m.store.Load(ctx, sessionID)
		if err == nil {
			return nil
		}

		if err != domain.ErrSessionNotFound {
			return fmt.Errorf("failed to check session existence: %w", err)
		}

		// Not found, create new
		if startNode == "" {
			startNode = "start"
		}
		state = domain.NewState(startNode)

		// Persist immediately to reserve the ID
		if err := m.store.Save(ctx, sessionID, state); err != nil {
			return fmt.Errorf("failed to initialize session: %w", err)
		}
		return nil
	})
	return state, err
}

// Save persists the session state.
func (m *Manager) Save(ctx context.Context, sessionID string, state *domain.State) error {
	return m.WithLock(ctx, sessionID, func(ctx context.Context) error {
		return m.store.Save(ctx, sessionID, state)
	})
}

// Delete removes the session from the store.
func (m *Manager) Delete(ctx context.Context, sessionID string) error {
	return m.WithLock(ctx, sessionID, func(ctx context.Context) error {
		return m.store.Delete(ctx, sessionID)
	})
}

// List delegates to the store.
func (m *Manager) List(ctx context.Context) ([]string, error) {
	return m.store.List(ctx)
}

// Store returns the underlying state store.
func (m *Manager) Store() ports.StateStore {
	return m.store
}

// WithLock executes a function while holding the lock for the session.
func (m *Manager) WithLock(ctx context.Context, sessionID string, fn func(context.Context) error) error {
	entry := m.acquire(sessionID)
	entry.mu.Lock()
	defer func() {
		entry.mu.Unlock()
		m.release(sessionID)
	}()

	// Distributed Locking
	if m.locker != nil {
		// TODO: Configure TTL?
		unlock, err := m.locker.Lock(ctx, sessionID, 30*time.Second)
		if err != nil {
			return fmt.Errorf("failed to acquire distributed lock: %w", err)
		}
		defer func() {
			if err := unlock(ctx); err != nil {
				m.logger.Warn("Failed to release distributed lock (will expire via TTL)",
					"session_id", sessionID,
					"err", err,
				)
			}
		}()
	}

	return fn(ctx)
}
