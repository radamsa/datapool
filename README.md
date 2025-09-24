# DataPool

DataPool is a lightweight, concurrent-safe in-memory key-value store with timestamp tracking. It's ideal for caching data that needs to be periodically refreshed or validated against a source of truth.

## Features

- **Simple API**: Easy to use interface for storing and retrieving values
- **Timestamp Tracking**: Each value is automatically tracked with a timestamp
- **Freshness Detection**: Determine if values are fresh based on timestamp comparison
- **Concurrent Safety**: Thread-safe operations for multi-goroutine environments
- **Type-Agnostic Storage**: Store any data type using Go's `interface{}`
- **Named Buckets**: Organize data into named buckets

## Installation

```go
import "github.com/user/datapool"
```

Make sure to replace `github.com/user/datapool` with your actual import path.

## Usage

### Basic Usage

```go
// Create a new data pool
pool := datapool.NewDataPool()

// Get or create a bucket
bucket := pool.Bucket("users")

// Store a value
bucket.Put("user data")

// Retrieve a value with timestamp and freshness
value, timestamp, isFresh := bucket.Get(lastCheckTime)

// 'isFresh' is true if the bucket's timestamp is newer than lastCheckTime
```

### Working with Different Data Types

DataPool can store any type of Go data:

```go
// Store integers
counterBucket := pool.Bucket("counter")
counterBucket.Put(42)

// Store structs
type User struct {
    Name string
    Age  int
}
userBucket := pool.Bucket("user")
userBucket.Put(User{Name: "Alice", Age: 30})

// Store maps
configBucket := pool.Bucket("config")
configBucket.Put(map[string]string{"theme": "dark", "language": "en"})
```

### Checking Freshness

The key feature of DataPool is the ability to check if data is fresh:

```go
// Store a timestamp when you last checked an external resource
lastChecked := time.Now().UnixNano()

// Later, store new data when you update from the external resource
bucket := pool.Bucket("resource")
bucket.Put(newData)

// When retrieving, pass the last checked time to determine if the value is fresh
value, timestamp, isFresh := bucket.Get(lastChecked)

if isFresh {
    // The data was updated after lastChecked
    fmt.Println("Data is fresh!")
} else {
    // The data hasn't been updated since lastChecked
    fmt.Println("Data might be stale")
}
```

### Concurrent Access

DataPool is designed for concurrent access:

```go
// This is safe to call from multiple goroutines
go func() {
    bucket := pool.Bucket("concurrent")
    bucket.Put("value from goroutine 1")
}()

go func() {
    bucket := pool.Bucket("concurrent")
    val, _, _ := bucket.Get(0)
    fmt.Println(val)
}()
```

## How It Works

DataPool organizes data into buckets, each identified by a name. When you put a value into a bucket, it's stored along with the current timestamp. When retrieving a value, you can provide a comparison timestamp to determine if the value is "fresh" (newer than the comparison timestamp).

This is particularly useful for caching scenarios where you need to know if your cached data needs to be refreshed from an authoritative source.

## Thread Safety

All operations in DataPool are thread-safe. Each bucket uses a read-write mutex to ensure that concurrent operations don't conflict. This allows DataPool to be safely used in multi-goroutine environments.

## Development

### Prerequisites

- Go 1.13 or higher

### Building and Testing

The project includes a Makefile with common commands:

```bash
# Run tests
make test

# Run tests with coverage
make coverage

# Format code
make fmt

# Lint code
make lint

# Show all available commands
make help
```

## Benchmarks

DataPool is designed to be lightweight and efficient. Here are some benchmark results:

```
BenchmarkBucketPut-20          15407640    68.27 ns/op     7 B/op    0 allocs/op
BenchmarkBucketGet-20          84577311    13.90 ns/op     0 B/op    0 allocs/op
BenchmarkConcurrentAccess-20    9684880   159.90 ns/op     0 B/op    0 allocs/op
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.