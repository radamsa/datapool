package datapool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDataPool(t *testing.T) {
	pool := NewDataPool()

	assert.NotNil(t, pool)
}

func TestBucket(t *testing.T) {
	pool := NewDataPool()
	bucket := pool.Bucket("test")
	bucket2 := pool.Bucket("test")
	assert.Equal(t, bucket.id, bucket2.id)
}

func TestUnfreshValue(t *testing.T) {
	pool := NewDataPool()
	bucket := pool.Bucket("test")
	timestamp := bucket.Put(10)

	value, timestamp2, ok := bucket.Get(timestamp + 1)
	assert.Equal(t, 10, value)
	assert.Equal(t, timestamp, timestamp2)
	assert.False(t, ok)
}

func TestFreshValue(t *testing.T) {
	pool := NewDataPool()
	bucket := pool.Bucket("test")
	timestamp := bucket.Put(10)

	value, timestamp2, ok := bucket.Get(timestamp - 1)
	assert.Equal(t, 10, value)
	assert.Equal(t, timestamp, timestamp2)
	assert.True(t, ok)
}
