package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aretw0/trellis/pkg/ports"
	backend "github.com/redis/go-redis/v9"
)

var (
	// ErrLockAcquire is returned when the lock cannot be acquired.
	ErrLockAcquire = errors.New("failed to acquire distributed lock")
)

// Locker implements ports.DistributedLocker using Redis.
type Locker struct {
	client *backend.Client
	prefix string
}

// NewLocker creates a new Redis locker.
func NewLocker(client *backend.Client, prefix string) *Locker {
	return &Locker{
		client: client,
		prefix: prefix,
	}
}

// Lock acquires a distributed lock for the given key using Redis SET NX PX.
func (l *Locker) Lock(ctx context.Context, key string, ttl time.Duration) (ports.UnlockFunc, error) {
	lockKey := l.prefix + "lock:" + key
	// Value is random (e.g. uuid) to ensure safe release, but for simplicity here we assume
	// that if we hold the lock, we are the ones unlocking it.
	// To be safer, we should check the value on unlock (Lua script).

	val := fmt.Sprintf("%d", time.Now().UnixNano())

	// Use a simple polling loop with backoff to acquire the lock.

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			// Try to acquire
			success, err := l.client.SetNX(ctx, lockKey, val, ttl).Result()
			if err != nil {
				return nil, fmt.Errorf("redis error acquiring lock: %w", err)
			}
			if success {
				// Lock acquired!
				return func(ctx context.Context) error {
					// Safe Unlock using Lua script to check value match
					// script: if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end
					script := `
						if redis.call("get", KEYS[1]) == ARGV[1] then
							return redis.call("del", KEYS[1])
						else
							return 0
						end
					`
					return l.client.Eval(ctx, script, []string{lockKey}, val).Err()
				}, nil
			}
			// Retry...
		}
	}
}
