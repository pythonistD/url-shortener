package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Counter struct {
	client *redis.Client
	key    string
}

func New(client *redis.Client, key string) *Counter {
	if key == "" {
		key = "url_shortener:counter"
	}
	return &Counter{
		client: client,
		key:    key,
	}
}

func (c *Counter) Inc(ctx context.Context) (uint64, error) {
	val, err := c.client.Incr(ctx, c.key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment redis counter: %w", err)
	}
	if val < 0 {
		return 0, fmt.Errorf("unexpected negative redis counter value: %d", val)
	}
	return uint64(val), nil
}
