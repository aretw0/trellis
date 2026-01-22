package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aretw0/trellis/pkg/domain"
	backend "github.com/redis/go-redis/v9"
)

// Store implements ports.StateStore using Redis.
type Store struct {
	client *backend.Client
	prefix string
	ttl    time.Duration
}

type Option func(*Store)

// WithTTL sets the expiration for sessions.
func WithTTL(ttl time.Duration) Option {
	return func(s *Store) {
		s.ttl = ttl
	}
}

// WithPrefix sets the key prefix for sessions.
func WithPrefix(prefix string) Option {
	return func(s *Store) {
		s.prefix = prefix
	}
}

// New creates a new Redis store with options.
func New(address, password string, db int, opts ...Option) *Store {
	rdb := backend.NewClient(&backend.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})

	store := &Store{
		client: rdb,
		prefix: "trellis:session:",
		ttl:    0, // No expiration by default
	}

	for _, opt := range opts {
		opt(store)
	}

	return store
}

// NewFromClient creates a new Redis store from an existing client.
func NewFromClient(client *backend.Client, opts ...Option) *Store {
	store := &Store{
		client: client,
		prefix: "trellis:session:",
		ttl:    0,
	}

	for _, opt := range opts {
		opt(store)
	}

	return store
}

func (s *Store) key(sessionID string) string {
	return s.prefix + sessionID
}

func (s *Store) indexKey() string {
	return s.prefix + "index"
}

// Save persists the state to Redis.
func (s *Store) Save(ctx context.Context, sessionID string, state *domain.State) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	pipe := s.client.Pipeline()

	// 1. Save JSON with TTL
	// Use 0 for no expiration if ttl is not set.
	pipe.Set(ctx, s.key(sessionID), data, s.ttl)

	// 2. Add to Index (ZSET)
	// Score = Now + TTL. If TTL = 0, Score = +Inf (approx).
	score := float64(time.Now().Add(s.ttl).Unix())
	if s.ttl == 0 {
		score = 4102444800 // 2100-01-01 (Far enough for now)
	}

	pipe.ZAdd(ctx, s.indexKey(), backend.Z{
		Score:  score,
		Member: sessionID,
	})

	// Execute pipeline
	_, err = pipe.Exec(ctx)
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
	pipe := s.client.Pipeline()

	pipe.Del(ctx, s.key(sessionID))
	pipe.ZRem(ctx, s.indexKey(), sessionID)

	_, err := pipe.Exec(ctx)
	return err
}

// List returns active sessions by scanning keys.
// Updated to use ZSET lazy cleanup.
func (s *Store) List(ctx context.Context) ([]string, error) {
	// Lazy Cleanup: Remove expired keys from Index
	now := float64(time.Now().Unix())

	// If TTL > 0, we can rely on cleanup.
	// If everything is infinite, this removes nothing.
	// ZREMRANGEBYSCORE key -inf (now)
	err := s.client.ZRemRangeByScore(ctx, s.indexKey(), "-inf", fmt.Sprintf("%f", now)).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to prune expired sessions: %w", err)
	}

	// Get remaining sessions
	sessions, err := s.client.ZRange(ctx, s.indexKey(), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	return sessions, nil
}

// Close closes the redis client.
func (s *Store) Close() error {
	return s.client.Close()
}
