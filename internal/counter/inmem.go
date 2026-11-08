package counter

import "sync/atomic"

type InMemory struct {
	val atomic.Uint64
}

func NewInMemory(initial uint64) *InMemory {
	c := &InMemory{}
	c.val.Store(initial)
	return c
}

func (c *InMemory) Inc() uint64 {
	return c.val.Add(1)
}

func (c *InMemory) Get() uint64 {
	return c.val.Load()
}
