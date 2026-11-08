//go:build integration

package redis

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

var testRedisClient *redis.Client

func TestMain(m *testing.M) {
	ctx := context.Background()

	redisContainer, err := tcredis.Run(ctx,
		"redis:8-alpine",
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start redis container: %v\n", err)
		os.Exit(1)
	}

	connStr, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get redis connection string: %v\n", err)
		_ = redisContainer.Terminate(ctx)
		os.Exit(1)
	}

	opts, err := redis.ParseURL(connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse redis url: %v\n", err)
		_ = redisContainer.Terminate(ctx)
		os.Exit(1)
	}

	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to ping redis: %v\n", err)
		_ = client.Close()
		_ = redisContainer.Terminate(ctx)
		os.Exit(1)
	}
	testRedisClient = client

	code := m.Run()

	_ = client.Close()
	_ = redisContainer.Terminate(context.Background())

	os.Exit(code)
}

func TestRedisCounter_SequentialIncrement(t *testing.T) {
	ctx := context.Background()
	key := fmt.Sprintf("test:counter:%d", time.Now().UnixNano())

	c := New(testRedisClient, key)

	for i := uint64(1); i <= 5; i++ {
		val, err := c.Inc(ctx)
		require.NoError(t, err)
		assert.Equal(t, i, val)
	}
}

func TestRedisCounter_ConcurrentIncrements(t *testing.T) {
	ctx := context.Background()
	key := fmt.Sprintf("test:counter:%d", time.Now().UnixNano())
	c := New(testRedisClient, key)

	const goroutines = 50
	const incsPerGoroutine = 20
	const expectedTotal = uint64(goroutines * incsPerGoroutine)

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incsPerGoroutine; j++ {
				_, err := c.Inc(context.Background())
				assert.NoError(t, err)
			}
		}()
	}

	wg.Wait()

	current, err := testRedisClient.Get(ctx, key).Uint64()
	require.NoError(t, err)
	assert.Equal(t, expectedTotal, current)
}

func TestRedisCounter_CanceledContext(t *testing.T) {
	key := fmt.Sprintf("test:counter:%d", time.Now().UnixNano())
	c := New(testRedisClient, key)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	val, err := c.Inc(ctx)
	require.Error(t, err)
	assert.Equal(t, uint64(0), val)
}
