# Go Concurrency Patterns Guide

This document provides detailed information about each concurrency pattern in this repository, including usage scenarios, best practices, and performance characteristics.

## Table of Contents

1. [Basic Patterns (1-8)](#basic-patterns)
2. [Google Search Examples (9-12)](#google-search-examples)
3. [Advanced Patterns (13-18)](#advanced-patterns)
4. [Performance Analysis](#performance-analysis)
5. [Best Practices](#best-practices)
6. [Common Pitfalls](#common-pitfalls)

## Basic Patterns (1-8)

### 1. Boring Goroutine (`1-boring`)

**Pattern**: Basic goroutine creation and communication
**Use Cases**: 
- Background processing
- Asynchronous operations
- Event generation

**Key Concepts**:
- Goroutines don't wait for each other by default
- Main goroutine termination kills all other goroutines
- Communication vs. shared memory

**Best Practices**:
- Always ensure main goroutine waits for workers to complete
- Use channels for communication rather than shared variables
- Avoid infinite loops without exit conditions

**Performance**: Low overhead, excellent for I/O bound tasks

### 2. Channel Communication (`2-chan`)

**Pattern**: Synchronous channel communication
**Use Cases**:
- Producer-consumer patterns
- Synchronous data exchange
- Sequential processing

**Key Concepts**:
- Unbuffered channels provide synchronization
- Channel operations block until both sides are ready
- Channel direction types enhance type safety

**Best Practices**:
- Use receive-only (`<-chan`) and send-only (`chan<-`) types for clarity
- Close channels from sender side only
- Use range loops for consuming all channel values

**Performance**: Synchronization overhead, but ensures ordering

### 3. Generator Pattern (`3-generator`)

**Pattern**: Function returns a channel for data generation
**Use Cases**:
- Data stream processing
- Iterator-like functionality
- Lazy evaluation

**Key Concepts**:
- Functions can return channels
- Encapsulates goroutine creation
- Provides clean API for data generation

**Best Practices**:
- Always close channels when generation is complete
- Handle context cancellation for long-running generators
- Use buffered channels for performance when appropriate

**Performance**: Good for streaming data, memory efficient

### 4. Fan-in Pattern (`4-fanin`)

**Pattern**: Multiple input channels merged into one output channel
**Use Cases**:
- Merging results from multiple sources
- Load balancing
- Data aggregation

**Key Concepts**:
- Multiple goroutines can send to the same channel
- Race condition handling
- Graceful shutdown coordination

**Best Practices**:
- Use sync.WaitGroup for proper shutdown
- Consider ordering requirements
- Handle different input channel closing patterns

**Performance**: Excellent for parallel processing, scales with input channels

### 5. Restore Sequence (`5-restore-sequence`)

**Pattern**: Maintaining message ordering with synchronization
**Use Cases**:
- When order matters despite concurrency
- Coordinated parallel processing
- Request-response patterns

**Key Concepts**:
- Two-phase communication (data + acknowledgment)
- Backpressure mechanism
- Synchronization between producer and consumer

**Best Practices**:
- Use for scenarios where ordering is critical
- Understand the performance trade-offs
- Consider alternative patterns for high-throughput scenarios

**Performance**: Higher latency due to synchronization, but guarantees ordering

### 6. Select with Timeout (`6-select-timeout`)

**Pattern**: Non-blocking channel operations with timeout
**Use Cases**:
- Network operations with timeouts
- Preventing goroutine leaks
- Implementing circuit breakers

**Key Concepts**:
- `time.After()` creates a timeout channel
- `select` enables non-blocking operations
- Default case for immediate non-blocking behavior

**Best Practices**:
- Always set reasonable timeouts
- Use context for hierarchical timeouts
- Handle timeout cases gracefully

**Performance**: Low overhead, prevents blocking indefinitely

### 7. Quit Signal (`7-quit-signal`)

**Pattern**: Graceful shutdown using quit channel
**Use Cases**:
- Graceful service shutdown
- Cancelling long-running operations
- Resource cleanup

**Key Concepts**:
- Quit channel for shutdown signaling
- Bidirectional communication for cleanup acknowledgment
- Select statement for multiple channel monitoring

**Best Practices**:
- Use context.Context for modern applications
- Implement proper cleanup before termination
- Handle cleanup timeouts

**Performance**: Minimal overhead, essential for production systems

### 8. Daisy Chain (`8-daisy-chan`)

**Pattern**: Sequential channel pipeline processing
**Use Cases**:
- Pipeline processing
- Demonstrating goroutine overhead
- Understanding Go scheduler behavior

**Key Concepts**:
- Chain of goroutines with channels
- Demonstrates Go's efficient goroutine handling
- Sequential message passing

**Best Practices**:
- Use for educational purposes primarily
- Consider performance implications for large chains
- Understand memory usage with many goroutines

**Performance**: Demonstrates Go's lightweight goroutines, but sequential nature limits throughput

## Google Search Examples (9-12)

### 9. Sequential Search (`9-google1.0`)

**Pattern**: Sequential operation execution
**Use Cases**:
- Baseline for performance comparison
- When order is strictly required
- Simple, predictable behavior

**Performance**: Slowest but most predictable

### 10. Concurrent Search (`10-google2.0`)

**Pattern**: Parallel execution with result collection
**Use Cases**:
- Independent parallel operations
- Performance improvement over sequential
- Multiple service calls

**Performance**: ~3x faster than sequential for 3 operations

### 11. Concurrent Search with Timeout (`11-google2.1`)

**Pattern**: Concurrent execution with global timeout
**Use Cases**:
- Web services with SLA requirements
- Preventing slow operations from blocking
- Circuit breaker patterns

**Performance**: Bounded latency, may lose some results

### 12. Concurrent Search with Replication (`12-google3.0`)

**Pattern**: Multiple replicas with first-response wins
**Use Cases**:
- High availability systems
- Reducing tail latency
- Fault tolerance

**Performance**: Best latency, highest resource usage

## Advanced Patterns (13-18)

### 13. Ping-Pong (`13-adv-pingpong`)

**Pattern**: State sharing through channel passing
**Use Cases**:
- Game simulations
- State machines
- Token passing systems

**Key Concepts**:
- Shared state through message passing
- Coordination between goroutines
- Avoiding race conditions

**Best Practices**:
- Use for exclusive access patterns
- Consider performance vs. simplicity trade-offs
- Handle termination conditions properly

### 14. Advanced Subscription (`14-adv-subscription`)

**Pattern**: Complex publisher-subscriber with backpressure
**Use Cases**:
- RSS feed processing
- Event streaming
- Real-time data processing

**Key Concepts**:
- Asynchronous fetching
- Backpressure handling
- Deduplication
- Graceful shutdown

**Best Practices**:
- Implement proper backpressure
- Handle duplicate events
- Use context for cancellation
- Monitor memory usage

### 15. Bounded Parallelism (`15-bounded-parallelism`)

**Pattern**: Worker pool with limited concurrency
**Use Cases**:
- File processing
- Rate-limited API calls
- Resource-constrained operations

**Key Concepts**:
- Limiting concurrent operations
- Work distribution
- Result collection
- Error handling

**Best Practices**:
- Size worker pool based on resources
- Handle errors gracefully
- Use buffered channels appropriately
- Monitor worker utilization

### 16. Context Usage (`16-context`)

**Pattern**: Request scoped cancellation and timeouts
**Use Cases**:
- HTTP request handling
- Database operations
- Microservice calls

**Key Concepts**:
- Hierarchical cancellation
- Deadline propagation
- Value passing (use sparingly)

**Best Practices**:
- Always respect context cancellation
- Use context.WithTimeout for operations
- Don't store context in structs
- Pass context as first parameter

### 17. Ring Buffer Channel (`17-ring-buffer-channel`)

**Pattern**: Fixed-size buffer with overwrite behavior
**Use Cases**:
- Log buffering
- Real-time data streams
- Memory-bounded queues

**Key Concepts**:
- Bounded memory usage
- Overwrite old data when full
- Non-blocking writes

**Best Practices**:
- Use when data loss is acceptable
- Size buffer appropriately
- Monitor buffer utilization
- Consider alternative patterns for critical data

### 18. Worker Pool (`18-worker-pool`)

**Pattern**: Pool of workers processing jobs from a queue
**Use Cases**:
- Batch processing
- Task distribution
- Load balancing

**Key Concepts**:
- Job distribution
- Worker lifecycle management
- Result collection
- Graceful shutdown

**Best Practices**:
- Size pool based on workload
- Handle worker failures
- Implement proper shutdown
- Monitor queue depth

## Performance Analysis

### Benchmark Results Summary

| Pattern | Throughput | Latency | Memory Usage | CPU Usage |
|---------|------------|---------|--------------|-----------|
| Sequential | Low | High | Low | Low |
| Fan-in | High | Medium | Medium | Medium |
| Worker Pool | Very High | Low | Medium | High |
| Ring Buffer | High | Very Low | Very Low | Medium |
| Context Timeout | Medium | Low | Low | Low |

### Scalability Characteristics

- **Channel Operations**: Scale well with goroutine count
- **Worker Pools**: Linear scaling up to CPU count
- **Fan-in Patterns**: Excellent for I/O bound operations
- **Timeout Patterns**: Constant overhead regardless of scale

## Best Practices

### General Guidelines

1. **Use channels for communication, mutexes for shared state**
2. **Always handle goroutine lifecycle properly**
3. **Implement graceful shutdown patterns**
4. **Use context for cancellation and timeouts**
5. **Size buffers and pools appropriately**
6. **Monitor and measure performance**
7. **Handle errors gracefully**
8. **Avoid goroutine leaks**

### Channel Best Practices

- Close channels from sender side only
- Use buffered channels for performance, unbuffered for synchronization
- Check channel closure with two-value receive
- Use select for non-blocking operations
- Implement proper backpressure mechanisms

### Goroutine Best Practices

- Start goroutines when you have concurrent work
- Ensure goroutines exit when work is done
- Use sync.WaitGroup or channels for coordination
- Avoid creating goroutines in tight loops
- Monitor goroutine count in production

## Common Pitfalls

### 1. Goroutine Leaks
```go
// Bad: goroutine may never exit
go func() {
    for {
        select {
        case data := <-ch:
            process(data)
        }
    }
}()

// Good: provide exit condition
go func() {
    for {
        select {
        case data := <-ch:
            process(data)
        case <-ctx.Done():
            return
        }
    }
}()
```

### 2. Channel Deadlocks
```go
// Bad: deadlock on unbuffered channel
ch := make(chan int)
ch <- 1 // blocks forever

// Good: use buffered channel or separate goroutine
ch := make(chan int, 1)
ch <- 1 // doesn't block
```

### 3. Race Conditions
```go
// Bad: race condition
var counter int
go func() { counter++ }()
go func() { counter++ }()

// Good: use channels or sync primitives
ch := make(chan int, 1)
ch <- 0
go func() {
    val := <-ch
    ch <- val + 1
}()
```

### 4. Incorrect Channel Closing
```go
// Bad: closing from receiver side
go func() {
    for data := range ch {
        process(data)
    }
    close(ch) // Wrong!
}()

// Good: close from sender side
go func() {
    defer close(ch)
    for i := 0; i < 10; i++ {
        ch <- i
    }
}()
```

### 5. Context Misuse
```go
// Bad: storing context in struct
type Service struct {
    ctx context.Context
}

// Good: pass context as parameter
func (s *Service) Process(ctx context.Context, data Data) error {
    // use ctx here
}
```

## Testing Concurrency

### Race Detection
```bash
go test -race ./...
```

### Benchmarking
```bash
go test -bench=. -benchmem ./...
```

### Stress Testing
```bash
go test -count=100 ./...
```

## Memory and Performance Considerations

### Channel Performance
- Unbuffered channels: Higher latency, perfect synchronization
- Small buffers: Good balance of performance and memory
- Large buffers: Lower latency, higher memory usage

### Goroutine Overhead
- Each goroutine: ~2KB initial stack
- Context switching: Microsecond range
- Creation cost: Nanosecond range

### Best Performance Practices
- Benchmark before optimizing
- Profile CPU and memory usage
- Use appropriate buffer sizes
- Monitor goroutine count
- Implement proper backpressure

This guide provides a comprehensive overview of Go concurrency patterns. Use it as a reference when implementing concurrent systems and always measure performance for your specific use case.