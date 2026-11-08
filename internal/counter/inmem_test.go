package counter

import (
	"testing"
)

func TestInMemory_Positive(t *testing.T) {
	t.Run("starts from 0 and increments sequentially", func(t *testing.T) {
		c := NewInMemory(0)

		if got := c.Inc(); got != 1 {
			t.Errorf("1: Inc() = %d, want 1", got)
		}
		if got := c.Inc(); got != 2 {
			t.Errorf("2: Inc() = %d, want 2", got)
		}
		if got := c.Inc(); got != 3 {
			t.Errorf("3: Inc() = %d, want 3", got)
		}
	})

	t.Run("starts from custom initial value", func(t *testing.T) {
		c := NewInMemory(100)

		if got := c.Inc(); got != 101 {
			t.Errorf("Inc() = %d, want 101", got)
		}
	})
}
