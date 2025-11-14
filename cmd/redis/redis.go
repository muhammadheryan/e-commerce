package redisclient

import (
	"context"
	"fmt"
	"time"

	"github.com/muhammadheryan/e-commerce/cmd/config"
	"github.com/redis/go-redis/v9"
)

var client *redis.Client

// New initializes the Redis client using provided configuration and verifies connectivity.
func New(cfg *config.Config) error {
	if cfg == nil {
		return fmt.Errorf("nil config provided")
	}

	addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	opt := &redis.Options{
		Addr:     addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	c := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("unable to ping redis at %s: %w", addr, err)
	}

	client = c
	return nil
}

func Get() *redis.Client {
	return client
}

func Close() error {
	if client == nil {
		return nil
	}
	return client.Close()
}
