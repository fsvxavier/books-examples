package cont

import (
	"sync"
	"sync/atomic"
)

// Mutex protege um contador com sync.Mutex.
type Mutex struct {
	mu sync.Mutex
	n  int64
}

func (c *Mutex) Inc() {
	c.mu.Lock()
	c.n++
	c.mu.Unlock()
}

// Atomico usa sync/atomic.
type Atomico struct{ n atomic.Int64 }

func (c *Atomico) Inc() { c.n.Add(1) }
