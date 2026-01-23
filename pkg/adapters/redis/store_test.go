package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/aretw0/trellis/pkg/adapters/redis"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
	backend "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisStore_Contract(t *testing.T) {
	// Setup miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	// Initialize client
	client := backend.NewClient(&backend.Options{
		Addr: mr.Addr(),
	})

	// Run contract
	store := redis.NewFromClient(client)
	ports.RunStateStoreContract(t, store)
}

func TestRedisStore_TTL_Expiration(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := backend.NewClient(&backend.Options{
		Addr: mr.Addr(),
	})

	// Create store with 1s TTL
	store := redis.NewFromClient(client, redis.WithTTL(1*time.Second))
	ctx := context.Background()
	sessionID := "session-ttl"
	state := &domain.State{
		CurrentNodeID: "node1",
		Context: map[string]interface{}{
			"foo": "bar",
		},
	}

	// 1. Save
	err = store.Save(ctx, sessionID, state)
	assert.NoError(t, err)

	// 2. Verify List (immediately)
	sessions, err := store.List(ctx)
	assert.NoError(t, err)
	assert.Contains(t, sessions, sessionID)

	// 3. Fast Forward time in miniredis (for Key Expiration)
	mr.FastForward(2 * time.Second)

	// 4. Verify Load (should fail)
	_, err = store.Load(ctx, sessionID)
	assert.ErrorIs(t, err, domain.ErrSessionNotFound)

	// 5. Verify List (lazily cleaned up)
	// Workaround for Test:
	// verification of lazy cleanup requires time.Sleep because our implementation relies on time.Now()
	// to calculate the score for ZRemRangeByScore.
	// We wait > 1s so time.Now() > (start + 1s).
	time.Sleep(1200 * time.Millisecond)

	sessions, err = store.List(ctx)
	assert.NoError(t, err)
	assert.Empty(t, sessions)
}

func TestRedisStore_Prefix(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := backend.NewClient(&backend.Options{
		Addr: mr.Addr(),
	})

	// Custom Prefix
	store := redis.NewFromClient(client, redis.WithPrefix("custom:app:"))
	ctx := context.Background()
	sessionID := "my-session"

	err = store.Save(ctx, sessionID, &domain.State{CurrentNodeID: "start"})
	assert.NoError(t, err)

	// Verify keys in Redis directly
	// Key should be "custom:app:my-session"
	exists := mr.Exists("custom:app:my-session")
	assert.True(t, exists, "Expected key with custom prefix to exist")

	// Index should be "custom:app:index"
	existsIndex := mr.Exists("custom:app:index")
	assert.True(t, existsIndex, "Expected index with custom prefix to exist")

	// Verify List works
	list, err := store.List(ctx)
	assert.NoError(t, err)
	assert.Contains(t, list, sessionID)
}
