package metrics

import (
	"sync"
	"sync/atomic"
)

// Counters keep track of multiple counters in a map
type Counters struct {
	sync.RWMutex
	m map[string]*uint64
}

// NewCounters returns a Counters instance
func NewCounters() (c *Counters) {
	c = &Counters{}
	c.m = make(map[string]*uint64)
	return c
}

// Count is used to update counters
func (c *Counters) Count(s string, delta uint64) {
	// Exists
	c.RLock()
	_, ok := c.m[s]
	if ok {
		atomic.AddUint64(c.m[s], delta)
		c.RUnlock()
		return
	}
	c.RUnlock()

	// Doesn't exist
	c.Lock()
	// did it get added in between droping read lock and acquiring write lock?
	_, ok = c.m[s]
	if ok {
		atomic.AddUint64(c.m[s], delta)
	} else {
		c.m[s] = &delta
	}
	c.Unlock()
	return
}

// Collect gets the current Counter map snapshot
func (c *Counters) Collect() map[string]uint64 {
	c.RLock()
	old := make(map[string]uint64, len(c.m))
	for s := range c.m {
		old[s] = atomic.SwapUint64(c.m[s], 0)
	}
	c.RUnlock()
	return old
}

// Get gets value of a counter
func (c *Counters) Get(s string) uint64 {
	c.RLock()
	v := c.m[s]
	c.RUnlock()
	return *v
}
