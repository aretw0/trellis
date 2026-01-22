package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aretw0/trellis/pkg/domain"
	backend "github.com/redis/go-redis/v9"
)

// Store implements ports.StateStore using Redis.
type Store struct {
	client *backend.Client
	prefix string
}

// New creates a new Redis store.
// address: "localhost:6379"
// password: "" (no password set)
// db: 0 (default DB)
func New(address, password string, db int) *Store {
	rdb := backend.NewClient(&backend.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})

	return &Store{
		client: rdb,
		prefix: "trellis:session:",
	}
}

// NewFromClient creates a new Redis store from an existing client.
func NewFromClient(client *backend.Client) *Store {
	return &Store{
		client: client,
		prefix: "trellis:session:",
	}
}

func (s *Store) key(sessionID string) string {
	return s.prefix + sessionID
}

// Save persists the state to Redis.
func (s *Store) Save(ctx context.Context, sessionID string, state *domain.State) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// We use 0 for no expiration, but in production we might want a default TTL.
	// For now, let's keep it indefinite as per "Durable Execution".
	err = s.client.Set(ctx, s.key(sessionID), data, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to save to redis: %w", err)
	}
	return nil
}

// Load retrieves the state from Redis.
func (s *Store) Load(ctx context.Context, sessionID string) (*domain.State, error) {
	val, err := s.client.Get(ctx, s.key(sessionID)).Result()
	if err != nil {
		if err == backend.Nil {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get from redis: %w", err)
	}

	var state domain.State
	if err := json.Unmarshal([]byte(val), &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// Delete removes the session.
func (s *Store) Delete(ctx context.Context, sessionID string) error {
	// Del returns number of keys removed. If 0, it means key didn't exist.
	// But our interface isn't strict about "delete non-existent", usually Delete is idempotent.
	// However, usually we don't return error if not found for delete.

	// Check store_test behavior:
	// "Load after Delete should return ErrSessionNotFound"

	err := s.client.Del(ctx, s.key(sessionID)).Err()
	return err
}

// List returns active sessions by scanning keys.
// Warning: SCAN can be slow on huge datasets.
func (s *Store) List(ctx context.Context) ([]string, error) {
	var sessions []string
	iter := s.client.Scan(ctx, 0, s.prefix+"*", 0).Iterator()

	for iter.Next(ctx) {
		// Key is "trellis:session:xyz", extract "xyz"
		key := iter.Val()
		if len(key) > len(s.prefix) {
			sessions = append(sessions, key[len(s.prefix):])
		}
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan redis keys: %w", err)
	}

	return sessions, nil
}

// Close closes the redis client.
func (s *Store) Close() error {
	return s.client.Close()
}
