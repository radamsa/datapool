package datapool

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDataPool(t *testing.T) {
	pool := NewDataPool()

	assert.NotNil(t, pool)
	assert.Empty(t, pool.buckets, "New pool should have no buckets")
}

func TestBucket(t *testing.T) {
	pool := NewDataPool()
	bucket := pool.Bucket("test")
	bucket2 := pool.Bucket("test")

	assert.Equal(t, bucket.id, bucket2.id, "Same name should return same bucket")

	// Different bucket names should create different buckets
	bucket3 := pool.Bucket("different")
	assert.NotEqual(t, bucket.id, bucket3.id, "Different name should return different bucket")

	// Test pool has correct number of buckets
	assert.Equal(t, 2, len(pool.buckets), "Pool should have two buckets")
}

func TestUnfreshValue(t *testing.T) {
	pool := NewDataPool()
	bucket := pool.Bucket("test")
	timestamp := bucket.Put(10)

	value, timestamp2, ok := bucket.Get(timestamp + 1)
	assert.Equal(t, 10, value)
	assert.Equal(t, timestamp, timestamp2)
	assert.False(t, ok, "Value should not be considered fresh when comparison timestamp is higher")
}

func TestFreshValue(t *testing.T) {
	pool := NewDataPool()
	bucket := pool.Bucket("test")
	timestamp := bucket.Put(10)

	value, timestamp2, ok := bucket.Get(timestamp - 1)
	assert.Equal(t, 10, value)
	assert.Equal(t, timestamp, timestamp2)
	assert.True(t, ok, "Value should be considered fresh when comparison timestamp is lower")
}

func TestEmptyBucket(t *testing.T) {
	pool := NewDataPool()
	bucket := pool.Bucket("test")

	// Get on a bucket with no value should return nil and timestamp 0
	value, timestamp, ok := bucket.Get(0)
	assert.Nil(t, value, "Empty bucket should return nil value")
	assert.Equal(t, int64(0), timestamp, "Empty bucket should have 0 timestamp")
	assert.False(t, ok, "Empty bucket should not be fresh")
}

func TestNilValue(t *testing.T) {
	pool := NewDataPool()
	bucket := pool.Bucket("test")

	// Store nil value
	timestamp := bucket.Put(nil)

	// Retrieve and verify
	value, timestamp2, ok := bucket.Get(0)
	assert.Nil(t, value, "Retrieved value should be nil")
	assert.Equal(t, timestamp, timestamp2)
	assert.True(t, ok, "Value should be fresh")
}

func TestInvalidBucketID(t *testing.T) {
	pool := NewDataPool()

	// Test with negative ID
	value, ts, ok := pool.get(-1, 0)
	assert.Nil(t, value)
	assert.Equal(t, int64(0), ts)
	assert.False(t, ok)

	// Test with out-of-bounds ID
	value, ts, ok = pool.get(999, 0)
	assert.Nil(t, value)
	assert.Equal(t, int64(0), ts)
	assert.False(t, ok)

	// Test put with invalid ID
	timestamp := pool.put(-1, "test")
	assert.Equal(t, int64(0), timestamp)
}

func TestSequentialUpdates(t *testing.T) {
	pool := NewDataPool()
	bucket := pool.Bucket("test")

	// First update
	timestamp1 := bucket.Put("first")
	value1, ts1, _ := bucket.Get(0)
	assert.Equal(t, "first", value1)
	assert.Equal(t, timestamp1, ts1)

	// Ensure timestamps are different
	time.Sleep(time.Millisecond)

	// Second update
	timestamp2 := bucket.Put("second")
	value2, ts2, _ := bucket.Get(0)
	assert.Equal(t, "second", value2)
	assert.Equal(t, timestamp2, ts2)

	// Timestamps should be increasing
	assert.Greater(t, timestamp2, timestamp1, "Second timestamp should be greater than first")
}

func TestConcurrentBucketAccess(t *testing.T) {
	pool := NewDataPool()
	bucket := pool.Bucket("concurrent")

	// Setup synchronized access
	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch goroutines that put values
	for i := 0; i < numGoroutines; i++ {
		go func(val int) {
			defer wg.Done()
			bucket.Put(val)
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Verify we have a value - which one isn't predictable
	value, ts, fresh := bucket.Get(0)
	assert.NotNil(t, value)
	assert.Greater(t, ts, int64(0))
	assert.True(t, fresh)
}

func TestConcurrentBucketCreation(t *testing.T) {
	pool := NewDataPool()

	// Create buckets with unique names from multiple goroutines
	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			name := fmt.Sprintf("bucket-%d", id)
			bucket := pool.Bucket(name)
			bucket.Put(id)
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Check that buckets were created
	assert.GreaterOrEqual(t, len(pool.buckets), 1)
	assert.LessOrEqual(t, len(pool.buckets), numGoroutines)

	// Check that we can retrieve buckets
	for i := 0; i < len(pool.buckets); i++ {
		assert.NotEmpty(t, pool.buckets[i].name)
	}
}

func TestConcurrentSameBucketName(t *testing.T) {
	pool := NewDataPool()

	// First create a bucket before the concurrent access
	bucket := pool.Bucket("same-name")
	bucket.Put("initial")

	// Multiple goroutines try to access the same bucket
	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(val int) {
			defer wg.Done()
			b := pool.Bucket("same-name")
			b.Put(val)
		}(i)
	}

	wg.Wait()

	// We should have exactly one bucket
	assert.Equal(t, 1, len(pool.buckets), "Should have only one bucket despite concurrent access")

	// The bucket should be accessible and have a value
	b := pool.Bucket("same-name")
	val, _, _ := b.Get(0)
	assert.NotNil(t, val)
}

func TestDifferentDataTypes(t *testing.T) {
	pool := NewDataPool()

	// Test string
	t.Run("String", func(t *testing.T) {
		b := pool.Bucket("string")
		b.Put("hello")
		val, _, _ := b.Get(0)
		assert.Equal(t, "hello", val)
	})

	// Test integer
	t.Run("Int", func(t *testing.T) {
		b := pool.Bucket("int")
		b.Put(42)
		val, _, _ := b.Get(0)
		assert.Equal(t, 42, val)
	})

	// Test float
	t.Run("Float", func(t *testing.T) {
		b := pool.Bucket("float")
		b.Put(3.14)
		val, _, _ := b.Get(0)
		assert.Equal(t, 3.14, val)
	})

	// Test struct
	t.Run("Struct", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		b := pool.Bucket("struct")
		person := Person{Name: "Alice", Age: 30}
		b.Put(person)

		val, _, _ := b.Get(0)
		retrieved, ok := val.(Person)
		assert.True(t, ok, "Should be able to cast to Person type")
		assert.Equal(t, person, retrieved)
	})

	// Test map
	t.Run("Map", func(t *testing.T) {
		b := pool.Bucket("map")
		data := map[string]int{"one": 1, "two": 2}
		b.Put(data)

		val, _, _ := b.Get(0)
		retrieved, ok := val.(map[string]int)
		assert.True(t, ok, "Should be able to cast to map type")
		assert.Equal(t, data, retrieved)
	})

	// Test slice
	t.Run("Slice", func(t *testing.T) {
		b := pool.Bucket("slice")
		data := []string{"apple", "banana", "cherry"}
		b.Put(data)

		val, _, _ := b.Get(0)
		retrieved, ok := val.([]string)
		assert.True(t, ok, "Should be able to cast to slice type")
		assert.Equal(t, data, retrieved)
	})
}

func TestComplexNestedStructure(t *testing.T) {
	pool := NewDataPool()
	bucket := pool.Bucket("complex")

	// Define nested structure
	type Address struct {
		Street string
		City   string
		Zip    string
	}

	type Person struct {
		Name      string
		Age       int
		Addresses []Address
		Metadata  map[string]interface{}
	}

	// Create complex data
	person := Person{
		Name: "John Doe",
		Age:  30,
		Addresses: []Address{
			{Street: "123 Main St", City: "Anytown", Zip: "12345"},
			{Street: "456 Oak Ave", City: "Othertown", Zip: "67890"},
		},
		Metadata: map[string]interface{}{
			"employed": true,
			"salary":   75000.50,
			"skills":   []string{"Go", "Python", "JavaScript"},
		},
	}

	// Store and retrieve
	bucket.Put(person)
	val, _, _ := bucket.Get(0)

	// Verify
	retrieved, ok := val.(Person)
	require.True(t, ok, "Should be able to cast to Person type")
	assert.Equal(t, "John Doe", retrieved.Name)
	assert.Equal(t, 30, retrieved.Age)
	assert.Len(t, retrieved.Addresses, 2)
	assert.Equal(t, "Anytown", retrieved.Addresses[0].City)
	assert.Equal(t, "67890", retrieved.Addresses[1].Zip)

	// Check nested structures
	assert.Equal(t, true, retrieved.Metadata["employed"])
	skills, ok := retrieved.Metadata["skills"].([]string)
	assert.True(t, ok)
	assert.Contains(t, skills, "Go")
}

func TestManyBuckets(t *testing.T) {
	pool := NewDataPool()
	const numBuckets = 100

	// Create many buckets
	for i := 0; i < numBuckets; i++ {
		name := fmt.Sprintf("bucket-%d", i)
		bucket := pool.Bucket(name)
		bucket.Put(i)
	}

	// Check all buckets were created
	assert.Equal(t, numBuckets, len(pool.buckets))

	// Verify values in each bucket
	for i := 0; i < numBuckets; i++ {
		name := fmt.Sprintf("bucket-%d", i)
		bucket := pool.Bucket(name)
		val, _, fresh := bucket.Get(0)
		assert.Equal(t, i, val)
		assert.True(t, fresh)
	}
}

func BenchmarkBucketPut(b *testing.B) {
	pool := NewDataPool()
	bucket := pool.Bucket("benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bucket.Put(i)
	}
}

func BenchmarkBucketGet(b *testing.B) {
	pool := NewDataPool()
	bucket := pool.Bucket("benchmark")
	bucket.Put("test value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bucket.Get(0)
	}
}

func BenchmarkConcurrentAccess(b *testing.B) {
	pool := NewDataPool()
	bucket := pool.Bucket("benchmark")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bucket.Put("value")
			bucket.Get(0)
		}
	})
}
