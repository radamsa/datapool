// Package datapool provides a concurrent-safe key-value store with timestamp tracking
// for freshness detection. It's useful for caching data that needs to be periodically
// refreshed or validated against a source of truth.
package datapool

import (
	"fmt"
	"sync"
	"time"
)

// DataPool is a concurrent-safe key-value store with timestamp tracking
// that allows checking for data freshness based on timestamps.
type DataPool struct {
	buckets []*bucket
}

// Bucket represents a named entry in the DataPool with methods to get and update values.
type Bucket struct {
	pool *DataPool
	id   int
}

type bucket struct {
	name      string
	value     any
	timestamp int64
	guard     sync.RWMutex
}

// NewDataPool creates a new empty DataPool instance.
func NewDataPool() *DataPool {
	return &DataPool{
		buckets: make([]*bucket, 0),
	}
}

func (p *DataPool) dump() {
	fmt.Println("Dump of DataPool:")
	for i, b := range p.buckets {
		fmt.Printf("[%d] (%d) %v\n", i, b.timestamp, b.value)
	}
	fmt.Println("--- end ---")
}

func (p *DataPool) get(id int, timestamp int64) (any, int64, bool) {
	if id < 0 || id >= len(p.buckets) {
		return nil, timestamp, false
	}

	p.buckets[id].guard.RLock()
	defer p.buckets[id].guard.RUnlock()

	return p.buckets[id].value, p.buckets[id].timestamp, p.buckets[id].timestamp > timestamp
}

func (p *DataPool) put(id int, value any) int64 {
	if id < 0 || id >= len(p.buckets) {
		return 0
	}

	p.buckets[id].guard.Lock()
	defer p.buckets[id].guard.Unlock()

	p.buckets[id].value = value
	p.buckets[id].timestamp = time.Now().UnixNano()

	return p.buckets[id].timestamp
}

// Bucket gets a bucket by name or creates a new one if it doesn't exist.
// It returns a Bucket reference that can be used for future operations.
func (p *DataPool) Bucket(name string) Bucket {
	for i, b := range p.buckets {
		if b.name == name {
			return Bucket{
				pool: p,
				id:   i,
			}
		}
	}

	b := &bucket{
		name:      name,
		timestamp: 0,
	}
	p.buckets = append(p.buckets, b)

	return Bucket{
		pool: p,
		id:   len(p.buckets) - 1,
	}
}

// Get returns the value of the bucket, its timestamp, and whether the value is fresher
// than the provided comparison timestamp. The boolean return value will be true if
// the bucket's timestamp is newer than the provided timestamp.
func (b *Bucket) Get(timestamp int64) (any, int64, bool) {
	return b.pool.get(b.id, timestamp)
}

// Put updates the value of the bucket and returns the new timestamp.
func (b *Bucket) Put(value any) int64 {
	return b.pool.put(b.id, value)
}
