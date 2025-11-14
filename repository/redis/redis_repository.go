package redis

import (
	"context"
	"time"

	redisclient "github.com/muhammadheryan/e-commerce/cmd/redis"
)

// Repository defines methods for interacting with Redis key-values
type Repository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}) error
	SetWithTTL(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	SetSession(ctx context.Context, sessionID string, userID uint64, ttl time.Duration) error
	GetSession(ctx context.Context, sessionID string) (uint64, error)
	DeleteSession(ctx context.Context, sessionID string) error
}

type redis struct {
	// *redis.Client
}

// NewRepository returns a Redis Repository implementation
func NewRepository() Repository {
	return &redis{}
}

// Get retrieves a value by key from Redis
func (r *redis) Get(ctx context.Context, key string) (string, error) {
	client := redisclient.Get()
	if client == nil {
		return "", nil
	}
	val, err := client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

// Set stores a key/value pair without expiration
func (r *redis) Set(ctx context.Context, key string, value interface{}) error {
	client := redisclient.Get()
	if client == nil {
		return nil
	}
	return client.Set(ctx, key, value, 0).Err()
}

// SetWithTTL stores a key/value pair with time-to-live
func (r *redis) SetWithTTL(ctx context.Context, key, value string, ttl time.Duration) error {
	client := redisclient.Get()
	if client == nil {
		return nil
	}
	return client.Set(ctx, key, value, ttl).Err()
}

// Delete removes a key from Redis
func (r *redis) Delete(ctx context.Context, key string) error {
	client := redisclient.Get()
	if client == nil {
		return nil
	}
	return client.Del(ctx, key).Err()
}

// SetSession stores a session with userID and TTL
func (r *redis) SetSession(ctx context.Context, sessionID string, userID uint64, ttl time.Duration) error {
	client := redisclient.Get()
	if client == nil {
		return nil
	}
	key := "session:" + sessionID
	return client.Set(ctx, key, userID, ttl).Err()
}

// GetSession retrieves userID from session
func (r *redis) GetSession(ctx context.Context, sessionID string) (uint64, error) {
	client := redisclient.Get()
	if client == nil {
		return 0, nil
	}
	key := "session:" + sessionID
	val, err := client.Get(ctx, key).Uint64()
	if err != nil {
		return 0, err
	}
	return val, nil
}

// DeleteSession removes a session from Redis
func (r *redis) DeleteSession(ctx context.Context, sessionID string) error {
	client := redisclient.Get()
	if client == nil {
		return nil
	}
	key := "session:" + sessionID
	return client.Del(ctx, key).Err()
}
