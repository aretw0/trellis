package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/aretw0/trellis/internal/adapters/redis"
	backend "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisLocker_LockUnlock(t *testing.T) {
	// Setup miniredis
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := backend.NewClient(&backend.Options{
		Addr: mr.Addr(),
	})
	locker := redis.NewLocker(client, "test:lock:")
	ctx := context.Background()
	key := "resource1"

	// 1. Acquire Lock
	unlock, err := locker.Lock(ctx, key, 5*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, unlock)

	// Verify key set in redis
	assert.True(t, mr.Exists("test:lock:lock:resource1"), "Lock key should be set in Redis")

	// 2. Release Lock
	err = unlock(ctx)
	assert.NoError(t, err)

	// Verify key removed
	assert.False(t, mr.Exists("test:lock:lock:resource1"), "Lock key should be removed after unlock")
}

func TestRedisLocker_Contention(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := backend.NewClient(&backend.Options{
		Addr: mr.Addr(),
	})
	locker1 := redis.NewLocker(client, "test:lock:")
	locker2 := redis.NewLocker(client, "test:lock:") // Same prefix -> contention
	ctx := context.Background()
	key := "shared-resource"

	// 1. Client 1 acquires lock
	unlock1, err := locker1.Lock(ctx, key, 5*time.Second)
	assert.NoError(t, err)
	assert.NotNil(t, unlock1)

	// 2. Client 2 tries to acquire (should block/retry until timeout or success)
	// Since we use polling in implementation, we need a timeout Context for Client 2
	ctxTimeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond) // Short timeout
	defer cancel()

	start := time.Now()
	_, err = locker2.Lock(ctxTimeout, key, 5*time.Second)

	// Should fail due to timeout (Client 1 holds it)
	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.WithinDuration(t, start.Add(500*time.Millisecond), time.Now(), 100*time.Millisecond, "Should block until timeout")

	// 3. Client 1 unlocks
	err = unlock1(ctx)
	assert.NoError(t, err)

	// 4. Client 2 tries again (should succeed)
	unlock2, err := locker2.Lock(ctx, key, 5*time.Second)
	assert.NoError(t, err)
	defer unlock2(ctx)

	assert.True(t, mr.Exists("test:lock:lock:shared-resource"))
}
