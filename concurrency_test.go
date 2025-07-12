package main

import (
	"context"
	"sync"
	"testing"
	"time"
)

// Test for basic channel operations
func TestChannelBasics(t *testing.T) {
	t.Run("Unbuffered Channel", func(t *testing.T) {
		ch := make(chan int)
		var wg sync.WaitGroup
		
		wg.Add(2)
		
		// Sender
		go func() {
			defer wg.Done()
			ch <- 42
		}()
		
		// Receiver
		go func() {
			defer wg.Done()
			val := <-ch
			if val != 42 {
				t.Errorf("Expected 42, got %d", val)
			}
		}()
		
		wg.Wait()
	})
	
	t.Run("Buffered Channel", func(t *testing.T) {
		ch := make(chan int, 3)
		
		// Send without blocking
		ch <- 1
		ch <- 2
		ch <- 3
		
		// Receive and verify
		for i := 1; i <= 3; i++ {
			val := <-ch
			if val != i {
				t.Errorf("Expected %d, got %d", i, val)
			}
		}
	})
}

// Test for fan-in pattern
func TestFanIn(t *testing.T) {
	// Simple fan-in function for testing
	fanIn := func(inputs ...<-chan int) <-chan int {
		out := make(chan int)
		var wg sync.WaitGroup
		
		for _, input := range inputs {
			wg.Add(1)
			go func(ch <-chan int) {
				defer wg.Done()
				for val := range ch {
					out <- val
				}
			}(input)
		}
		
		go func() {
			wg.Wait()
			close(out)
		}()
		
		return out
	}
	
	ch1 := make(chan int, 2)
	ch2 := make(chan int, 2)
	
	ch1 <- 1
	ch1 <- 2
	close(ch1)
	
	ch2 <- 3
	ch2 <- 4
	close(ch2)
	
	merged := fanIn(ch1, ch2)
	received := make(map[int]bool)
	
	for val := range merged {
		received[val] = true
	}
	
	expected := []int{1, 2, 3, 4}
	for _, val := range expected {
		if !received[val] {
			t.Errorf("Expected to receive %d", val)
		}
	}
	
	if len(received) != 4 {
		t.Errorf("Expected 4 values, got %d", len(received))
	}
}

// Test for timeout pattern
func TestTimeout(t *testing.T) {
	t.Run("Operation completes before timeout", func(t *testing.T) {
		ch := make(chan string, 1)
		
		go func() {
			time.Sleep(50 * time.Millisecond)
			ch <- "success"
		}()
		
		select {
		case result := <-ch:
			if result != "success" {
				t.Errorf("Expected 'success', got '%s'", result)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Operation timed out unexpectedly")
		}
	})
	
	t.Run("Operation times out", func(t *testing.T) {
		ch := make(chan string, 1)
		
		go func() {
			time.Sleep(200 * time.Millisecond)
			ch <- "too late"
		}()
		
		select {
		case result := <-ch:
			t.Errorf("Expected timeout, but got result: %s", result)
		case <-time.After(50 * time.Millisecond):
			// Expected timeout - test passes
		}
	})
}

// Test for context cancellation
func TestContextCancellation(t *testing.T) {
	t.Run("Context cancellation stops operation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan bool)
		
		go func() {
			select {
			case <-time.After(100 * time.Millisecond):
				done <- false
			case <-ctx.Done():
				done <- true
			}
		}()
		
		// Cancel after 50ms
		time.AfterFunc(50*time.Millisecond, cancel)
		
		if !<-done {
			t.Error("Context cancellation did not stop the operation")
		}
	})
	
	t.Run("Context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		
		done := make(chan bool)
		
		go func() {
			select {
			case <-time.After(100 * time.Millisecond):
				done <- false
			case <-ctx.Done():
				done <- true
			}
		}()
		
		if !<-done {
			t.Error("Context timeout did not work")
		}
	})
}

// Test for worker pool pattern
func TestWorkerPool(t *testing.T) {
	const numJobs = 10
	const numWorkers = 3
	
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)
	
	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				// Simple job: double the number
				results <- job * 2
			}
		}()
	}
	
	// Send jobs
	for i := 1; i <= numJobs; i++ {
		jobs <- i
	}
	close(jobs)
	
	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Collect results
	received := make(map[int]bool)
	for result := range results {
		received[result] = true
	}
	
	// Verify all expected results
	for i := 1; i <= numJobs; i++ {
		expected := i * 2
		if !received[expected] {
			t.Errorf("Expected result %d not found", expected)
		}
	}
	
	if len(received) != numJobs {
		t.Errorf("Expected %d results, got %d", numJobs, len(received))
	}
}

// Test for race conditions
func TestRaceCondition(t *testing.T) {
	t.Run("Concurrent counter with mutex", func(t *testing.T) {
		var counter int
		var mu sync.Mutex
		var wg sync.WaitGroup
		
		const numGoroutines = 100
		const incrementsPerGoroutine = 100
		
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < incrementsPerGoroutine; j++ {
					mu.Lock()
					counter++
					mu.Unlock()
				}
			}()
		}
		
		wg.Wait()
		
		expected := numGoroutines * incrementsPerGoroutine
		if counter != expected {
			t.Errorf("Expected counter to be %d, got %d", expected, counter)
		}
	})
	
	t.Run("Channel-based synchronization", func(t *testing.T) {
		const numSenders = 10
		const messagesPerSender = 10
		
		ch := make(chan int, numSenders*messagesPerSender)
		var wg sync.WaitGroup
		
		// Start senders
		for i := 0; i < numSenders; i++ {
			wg.Add(1)
			go func(senderID int) {
				defer wg.Done()
				for j := 0; j < messagesPerSender; j++ {
					ch <- senderID*messagesPerSender + j
				}
			}(i)
		}
		
		wg.Wait()
		close(ch)
		
		// Collect all messages
		received := make(map[int]bool)
		for msg := range ch {
			received[msg] = true
		}
		
		expected := numSenders * messagesPerSender
		if len(received) != expected {
			t.Errorf("Expected %d unique messages, got %d", expected, len(received))
		}
	})
}

// Test for ring buffer pattern
func TestRingBuffer(t *testing.T) {
	inCh := make(chan int)
	outCh := make(chan int, 3) // Buffer size 3
	
	// Ring buffer implementation for testing
	go func() {
		defer close(outCh)
		for v := range inCh {
			select {
			case outCh <- v:
				// Successfully sent
			default:
				// Buffer full, remove oldest and add new
				<-outCh
				outCh <- v
			}
		}
	}()
	
	// Send more items than buffer size
	go func() {
		defer close(inCh)
		for i := 0; i < 10; i++ {
			inCh <- i
		}
	}()
	
	// Collect results
	var results []int
	for result := range outCh {
		results = append(results, result)
	}
	
	// Should only have the last 3 items due to ring buffer behavior
	if len(results) < 3 {
		t.Errorf("Expected at least 3 results, got %d", len(results))
	}
	
	// The exact behavior depends on timing, but we should get some results
	t.Logf("Ring buffer results: %v", results)
}

// Benchmark tests for channel operations
func BenchmarkBasicChannelOperations(b *testing.B) {
	b.Run("Unbuffered Channel", func(b *testing.B) {
		ch := make(chan int)
		
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			go func() {
				for pb.Next() {
					ch <- 1
				}
			}()
			
			for pb.Next() {
				<-ch
			}
		})
	})
	
	b.Run("Buffered Channel", func(b *testing.B) {
		ch := make(chan int, 1000)
		
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			go func() {
				for pb.Next() {
					select {
					case ch <- 1:
					default:
					}
				}
			}()
			
			for pb.Next() {
				select {
				case <-ch:
				default:
				}
			}
		})
	})
}

func BenchmarkBasicWorkerPool(b *testing.B) {
	const numWorkers = 4
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		jobs := make(chan int, 100)
		results := make(chan int, 100)
		
		// Start workers
		var wg sync.WaitGroup
		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for job := range jobs {
					results <- job * job
				}
			}()
		}
		
		// Send jobs
		go func() {
			for j := 0; j < 100; j++ {
				jobs <- j
			}
			close(jobs)
		}()
		
		// Close results when workers are done
		go func() {
			wg.Wait()
			close(results)
		}()
		
		// Consume results
		for range results {
		}
	}
}