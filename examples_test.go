package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// Test the basic boring goroutine pattern (example 1)
func TestBoringPattern(t *testing.T) {
	ch := make(chan string, 5)
	
	// Simulate the boring function
	go func() {
		defer close(ch)
		for i := 0; i < 5; i++ {
			ch <- fmt.Sprintf("boring! %d", i)
			time.Sleep(10 * time.Millisecond)
		}
	}()
	
	count := 0
	for msg := range ch {
		if msg == "" {
			t.Error("Received empty message")
		}
		count++
	}
	
	if count != 5 {
		t.Errorf("Expected 5 messages, got %d", count)
	}
}

// Test the generator pattern (example 3)
func TestGeneratorPattern(t *testing.T) {
	generator := func(msg string) <-chan string {
		ch := make(chan string)
		go func() {
			defer close(ch)
			for i := 0; i < 3; i++ {
				ch <- fmt.Sprintf("%s %d", msg, i)
				time.Sleep(10 * time.Millisecond)
			}
		}()
		return ch
	}
	
	joe := generator("Joe")
	ann := generator("Ann")
	
	joeCount := 0
	annCount := 0
	
	for i := 0; i < 6; i++ {
		select {
		case msg := <-joe:
			if msg != "" {
				joeCount++
			}
		case msg := <-ann:
			if msg != "" {
				annCount++
			}
		case <-time.After(100 * time.Millisecond):
			// Timeout to prevent infinite waiting
			break
		}
	}
	
	if joeCount != 3 || annCount != 3 {
		t.Errorf("Expected 3 messages from each generator, got Joe: %d, Ann: %d", joeCount, annCount)
	}
}

// Test the fan-in pattern (example 4)
func TestFanInPattern(t *testing.T) {
	boring := func(msg string) <-chan string {
		ch := make(chan string)
		go func() {
			defer close(ch)
			for i := 0; i < 3; i++ {
				ch <- fmt.Sprintf("%s %d", msg, i)
				time.Sleep(10 * time.Millisecond)
			}
		}()
		return ch
	}
	
	fanIn := func(cs ...<-chan string) <-chan string {
		out := make(chan string)
		var wg sync.WaitGroup
		
		for _, ch := range cs {
			wg.Add(1)
			go func(input <-chan string) {
				defer wg.Done()
				for msg := range input {
					out <- msg
				}
			}(ch)
		}
		
		go func() {
			wg.Wait()
			close(out)
		}()
		
		return out
	}
	
	merged := fanIn(boring("Joe"), boring("Ann"))
	
	count := 0
	for range merged {
		count++
	}
	
	if count != 6 {
		t.Errorf("Expected 6 messages from fan-in, got %d", count)
	}
}

// Test the timeout pattern (example 6)
func TestTimeoutPattern(t *testing.T) {
	slowOperation := func() <-chan string {
		ch := make(chan string)
		go func() {
			defer close(ch)
			time.Sleep(100 * time.Millisecond)
			ch <- "completed"
		}()
		return ch
	}
	
	t.Run("Operation completes within timeout", func(t *testing.T) {
		ch := slowOperation()
		timeout := time.After(200 * time.Millisecond)
		
		select {
		case msg := <-ch:
			if msg != "completed" {
				t.Errorf("Expected 'completed', got '%s'", msg)
			}
		case <-timeout:
			t.Error("Operation timed out unexpectedly")
		}
	})
	
	t.Run("Operation times out", func(t *testing.T) {
		ch := slowOperation()
		timeout := time.After(50 * time.Millisecond)
		
		select {
		case msg := <-ch:
			t.Errorf("Expected timeout, but operation completed with: %s", msg)
		case <-timeout:
			// Expected behavior
		}
	})
}

// Test the quit signal pattern (example 7)
func TestQuitSignalPattern(t *testing.T) {
	boring := func(msg string, quit <-chan bool) <-chan string {
		ch := make(chan string)
		go func() {
			defer close(ch)
			for i := 0; ; i++ {
				select {
				case ch <- fmt.Sprintf("%s %d", msg, i):
					time.Sleep(10 * time.Millisecond)
				case <-quit:
					return
				}
			}
		}()
		return ch
	}
	
	quit := make(chan bool)
	ch := boring("Joe", quit)
	
	// Receive a few messages
	count := 0
	for i := 0; i < 3; i++ {
		select {
		case msg := <-ch:
			if msg == "" {
				t.Error("Received empty message")
			}
			count++
		case <-time.After(100 * time.Millisecond):
			t.Error("Timeout waiting for message")
		}
	}
	
	// Send quit signal
	close(quit)
	
	// Channel should close soon
	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				// Channel closed as expected
				goto done
			}
			// Still receiving messages, continue
			_ = msg
		case <-timeout:
			t.Error("Channel did not close after quit signal")
			goto done
		}
	}
	done:
	
	if count != 3 {
		t.Errorf("Expected 3 messages before quit, got %d", count)
	}
}

// Test the context pattern (example 16)
func TestContextPattern(t *testing.T) {
	sleepAndTalk := func(ctx context.Context, d time.Duration, msg string) error {
		select {
		case <-time.After(d):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		
		// Cancel after 50ms
		time.AfterFunc(50*time.Millisecond, cancel)
		
		// Operation that would take 100ms
		err := sleepAndTalk(ctx, 100*time.Millisecond, "hello")
		if err == nil {
			t.Error("Expected context cancellation error")
		}
		if err != context.Canceled {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	})
	
	t.Run("Operation completes before cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		
		// Cancel after 100ms
		time.AfterFunc(100*time.Millisecond, cancel)
		
		// Operation that takes 50ms
		err := sleepAndTalk(ctx, 50*time.Millisecond, "hello")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

// Test the ping-pong pattern (example 13)
func TestPingPongPattern(t *testing.T) {
	type Ball struct{ hits int }
	
	player := func(name string, table chan *Ball, maxHits int, done chan bool) {
		for {
			select {
			case ball, ok := <-table:
				if !ok {
					done <- true
					return
				}
				ball.hits++
				if ball.hits >= maxHits {
					close(table)
					done <- true
					return
				}
				table <- ball
			case <-time.After(100 * time.Millisecond):
				// Timeout to prevent infinite waiting
				done <- true
				return
			}
		}
	}
	
	table := make(chan *Ball, 1)
	done := make(chan bool, 2)
	
	go player("ping", table, 10, done)
	go player("pong", table, 10, done)
	
	table <- &Ball{}
	
	// Wait for one player to finish or timeout
	select {
	case <-done:
		// Game finished
	case <-time.After(200 * time.Millisecond):
		// Test timeout
		close(table)
	}
	
	// Drain the done channel
	select {
	case <-done:
	default:
	}
}

// Test worker pool pattern with bounded parallelism (example 18)
func TestWorkerPoolPattern(t *testing.T) {
	const numJobs = 20
	const numWorkers = 3
	
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)
	
	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				// Simulate work
				time.Sleep(1 * time.Millisecond)
				results <- job * 2
			}
		}()
	}
	
	// Send jobs
	go func() {
		defer close(jobs)
		for j := 1; j <= numJobs; j++ {
			jobs <- j
		}
	}()
	
	// Close results when workers are done
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Collect results
	resultMap := make(map[int]bool)
	for result := range results {
		resultMap[result] = true
	}
	
	// Verify all jobs were processed
	for i := 1; i <= numJobs; i++ {
		expected := i * 2
		if !resultMap[expected] {
			t.Errorf("Missing result for job %d (expected %d)", i, expected)
		}
	}
	
	if len(resultMap) != numJobs {
		t.Errorf("Expected %d results, got %d", numJobs, len(resultMap))
	}
}

// Test the Google search pattern (examples 9-12)
func TestGoogleSearchPattern(t *testing.T) {
	type Result string
	type Search func(query string) Result
	
	fakeSearch := func(kind string) Search {
		return func(query string) Result {
			time.Sleep(time.Duration(10) * time.Millisecond)
			return Result(fmt.Sprintf("%s result for %q", kind, query))
		}
	}
	
	Web := fakeSearch("web")
	Image := fakeSearch("image")
	Video := fakeSearch("video")
	
	t.Run("Sequential search", func(t *testing.T) {
		start := time.Now()
		
		var results []Result
		results = append(results, Web("golang"))
		results = append(results, Image("golang"))
		results = append(results, Video("golang"))
		
		elapsed := time.Since(start)
		
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
		
		// Should take at least 30ms (3 * 10ms)
		if elapsed < 30*time.Millisecond {
			t.Errorf("Sequential search too fast: %v", elapsed)
		}
	})
	
	t.Run("Concurrent search", func(t *testing.T) {
		start := time.Now()
		
		ch := make(chan Result, 3)
		
		go func() { ch <- Web("golang") }()
		go func() { ch <- Image("golang") }()
		go func() { ch <- Video("golang") }()
		
		var results []Result
		for i := 0; i < 3; i++ {
			results = append(results, <-ch)
		}
		
		elapsed := time.Since(start)
		
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
		
		// Should be faster than sequential (closer to 10ms than 30ms)
		if elapsed > 25*time.Millisecond {
			t.Errorf("Concurrent search too slow: %v", elapsed)
		}
	})
	
	t.Run("Concurrent search with timeout", func(t *testing.T) {
		slowSearch := func(kind string) Search {
			return func(query string) Result {
				time.Sleep(100 * time.Millisecond) // Intentionally slow
				return Result(fmt.Sprintf("%s result for %q", kind, query))
			}
		}
		
		SlowWeb := slowSearch("web")
		ch := make(chan Result, 1)
		
		go func() { ch <- SlowWeb("golang") }()
		
		var results []Result
		timeout := time.After(50 * time.Millisecond)
		
		select {
		case result := <-ch:
			results = append(results, result)
		case <-timeout:
			// Expected - operation should timeout
		}
		
		// Should have no results due to timeout
		if len(results) != 0 {
			t.Errorf("Expected 0 results due to timeout, got %d", len(results))
		}
	})
}