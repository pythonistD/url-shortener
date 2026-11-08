package inmem

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemory_Positive(t *testing.T) {
	ctx := context.Background()

	t.Run("starts from 0 and increments sequentially", func(t *testing.T) {
		c := New(0)

		val1, err := c.Inc(ctx)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), val1)

		val2, err := c.Inc(ctx)
		require.NoError(t, err)
		assert.Equal(t, uint64(2), val2)

		val3, err := c.Inc(ctx)
		require.NoError(t, err)
		assert.Equal(t, uint64(3), val3)
	})

	t.Run("starts from custom initial value", func(t *testing.T) {
		c := New(100)

		val, err := c.Inc(ctx)
		require.NoError(t, err)
		assert.Equal(t, uint64(101), val)
	})
}

func TestInMemory_CanceledContext(t *testing.T) {
	c := New(0)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	val, err := c.Inc(ctx)
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, uint64(0), val)
}
