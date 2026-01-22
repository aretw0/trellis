package redis_test

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/aretw0/trellis/internal/adapters/redis"
	"github.com/aretw0/trellis/pkg/ports"
	backend "github.com/redis/go-redis/v9"
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
