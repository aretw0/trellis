package ports

import (
	"context"
	"time"
)

// UnlockFunc is a function that releases a distributed lock.
type UnlockFunc func(ctx context.Context) error

// DistributedLocker defines the interface for distributed concurrency control.
// It allows the Session Manager to coordinate access across multiple instances (replicas).
type DistributedLocker interface {
	// Lock attempts to acquire a distributed lock for the given key (e.g., session ID).
	// It blocks until the lock is acquired, the context is canceled, or the TTL expires (implementation specific).
	// Returns an UnlockFunc that MUST be called to release the lock.
	Lock(ctx context.Context, key string, ttl time.Duration) (UnlockFunc, error)
}
