package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps redis.Client
type Client struct {
	*redis.Client
}

// Connect creates a new Redis client
func Connect(ctx context.Context, addr, password string, db int) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Verify connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{Client: client}, nil
}

// SetNX sets a key with expiration if it doesn't exist (for deduplication)
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.Client.SetNX(ctx, key, value, expiration).Result()
}

// GetSet retrieves and sets a new value atomically
func (c *Client) GetSet(ctx context.Context, key string, value interface{}) (string, error) {
	return c.Client.GetSet(ctx, key, value).Result()
}
