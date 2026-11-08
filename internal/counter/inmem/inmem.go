package inmem

import (
	"context"
	"sync/atomic"
)

type Counter struct {
	val atomic.Uint64
}

func New(initial uint64) *Counter {
	c := &Counter{}
	c.val.Store(initial)
	return c
}

func (c *Counter) Inc(ctx context.Context) (uint64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return c.val.Add(1), nil
}
