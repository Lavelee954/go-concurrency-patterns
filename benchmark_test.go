package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// BenchmarkBoringPattern benchmarks the basic goroutine communication
func BenchmarkBoringPattern(b *testing.B) {
	b.Run("SingleProducerConsumer", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ch := make(chan string, 100)
			
			go func() {
				defer close(ch)
				for j := 0; j < 100; j++ {
					ch <- fmt.Sprintf("msg %d", j)
				}
			}()
			
			count := 0
			for range ch {
				count++
			}
		}
	})
	
	b.Run("MultipleProducers", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ch := make(chan string, 100)
			var wg sync.WaitGroup
			
			// Start 4 producers
			for p := 0; p < 4; p++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					for j := 0; j < 25; j++ {
						ch <- fmt.Sprintf("producer-%d-msg-%d", id, j)
					}
				}(p)
			}
			
			go func() {
				wg.Wait()
				close(ch)
			}()
			
			count := 0
			for range ch {
				count++
			}
		}
	})
}

// BenchmarkChannelTypes compares different channel configurations
func BenchmarkChannelTypes(b *testing.B) {
	b.Run("UnbufferedChannel", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ch := make(chan int)
			
			go func() {
				defer close(ch)
				for j := 0; j < 1000; j++ {
					ch <- j
				}
			}()
			
			for range ch {
			}
		}
	})
	
	b.Run("BufferedChannel_Small", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ch := make(chan int, 10)
			
			go func() {
				defer close(ch)
				for j := 0; j < 1000; j++ {
					ch <- j
				}
			}()
			
			for range ch {
			}
		}
	})
	
	b.Run("BufferedChannel_Large", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ch := make(chan int, 1000)
			
			go func() {
				defer close(ch)
				for j := 0; j < 1000; j++ {
					ch <- j
				}
			}()
			
			for range ch {
			}
		}
	})
}

// BenchmarkFanInPattern compares different fan-in implementations
func BenchmarkFanInPattern(b *testing.B) {
	b.Run("SimpleFanIn", func(b *testing.B) {
		fanIn := func(inputs ...<-chan int) <-chan int {
			out := make(chan int)
			var wg sync.WaitGroup
			
			for _, ch := range inputs {
				wg.Add(1)
				go func(input <-chan int) {
					defer wg.Done()
					for val := range input {
						out <- val
					}
				}(ch)
			}
			
			go func() {
				wg.Wait()
				close(out)
			}()
			
			return out
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Create input channels
			inputs := make([]<-chan int, 4)
			for j := 0; j < 4; j++ {
				ch := make(chan int, 25)
				inputs[j] = ch
				
				go func(c chan int) {
					defer close(c)
					for k := 0; k < 25; k++ {
						c <- k
					}
				}(ch)
			}
			
			merged := fanIn(inputs...)
			count := 0
			for range merged {
				count++
			}
		}
	})
	
	b.Run("SelectBasedFanIn", func(b *testing.B) {
		fanInSelect := func(c1, c2 <-chan int) <-chan int {
			out := make(chan int)
			go func() {
				defer close(out)
				for {
					select {
					case val, ok := <-c1:
						if !ok {
							c1 = nil
						} else {
							out <- val
						}
					case val, ok := <-c2:
						if !ok {
							c2 = nil
						} else {
							out <- val
						}
					}
					if c1 == nil && c2 == nil {
						break
					}
				}
			}()
			return out
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ch1 := make(chan int, 50)
			ch2 := make(chan int, 50)
			
			go func() {
				defer close(ch1)
				for j := 0; j < 50; j++ {
					ch1 <- j
				}
			}()
			
			go func() {
				defer close(ch2)
				for j := 0; j < 50; j++ {
					ch2 <- j + 100
				}
			}()
			
			merged := fanInSelect(ch1, ch2)
			count := 0
			for range merged {
				count++
			}
		}
	})
}

// BenchmarkWorkerPool compares different worker pool configurations
func BenchmarkWorkerPool(b *testing.B) {
	workFunc := func(n int) int {
		// Simulate some CPU work
		sum := 0
		for i := 0; i < n; i++ {
			sum += i
		}
		return sum
	}
	
	b.Run("SingleWorker", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			jobs := make(chan int, 100)
			results := make(chan int, 100)
			
			// Single worker
			go func() {
				defer close(results)
				for job := range jobs {
					results <- workFunc(job)
				}
			}()
			
			// Send jobs
			go func() {
				defer close(jobs)
				for j := 0; j < 100; j++ {
					jobs <- j + 1
				}
			}()
			
			// Consume results
			for range results {
			}
		}
	})
	
	b.Run("FourWorkers", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			jobs := make(chan int, 100)
			results := make(chan int, 100)
			var wg sync.WaitGroup
			
			// Four workers
			for w := 0; w < 4; w++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for job := range jobs {
						results <- workFunc(job)
					}
				}()
			}
			
			// Close results when workers are done
			go func() {
				wg.Wait()
				close(results)
			}()
			
			// Send jobs
			go func() {
				defer close(jobs)
				for j := 0; j < 100; j++ {
					jobs <- j + 1
				}
			}()
			
			// Consume results
			for range results {
			}
		}
	})
	
	b.Run("TenWorkers", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			jobs := make(chan int, 100)
			results := make(chan int, 100)
			var wg sync.WaitGroup
			
			// Ten workers
			for w := 0; w < 10; w++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for job := range jobs {
						results <- workFunc(job)
					}
				}()
			}
			
			// Close results when workers are done
			go func() {
				wg.Wait()
				close(results)
			}()
			
			// Send jobs
			go func() {
				defer close(jobs)
				for j := 0; j < 100; j++ {
					jobs <- j + 1
				}
			}()
			
			// Consume results
			for range results {
			}
		}
	})
}

// BenchmarkTimeoutPatterns compares different timeout implementations
func BenchmarkTimeoutPatterns(b *testing.B) {
	b.Run("ChannelTimeout", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ch := make(chan string, 1)
			
			go func() {
				time.Sleep(1 * time.Millisecond)
				ch <- "done"
			}()
			
			select {
			case <-ch:
				// Success
			case <-time.After(10 * time.Millisecond):
				// Timeout
			}
		}
	})
	
	b.Run("ContextTimeout", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			
			ch := make(chan string, 1)
			
			go func() {
				time.Sleep(1 * time.Millisecond)
				ch <- "done"
			}()
			
			select {
			case <-ch:
				// Success
			case <-ctx.Done():
				// Timeout
			}
			
			cancel()
		}
	})
}

// BenchmarkSynchronization compares different synchronization methods
func BenchmarkSynchronization(b *testing.B) {
	b.Run("Mutex", func(b *testing.B) {
		var counter int
		var mu sync.Mutex
		
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		})
	})
	
	b.Run("RWMutex_Read", func(b *testing.B) {
		var counter int
		var mu sync.RWMutex
		
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				mu.RLock()
				_ = counter
				mu.RUnlock()
			}
		})
	})
	
	b.Run("Channel", func(b *testing.B) {
		ch := make(chan int, 1)
		ch <- 0
		
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				val := <-ch
				val++
				ch <- val
			}
		})
	})
	
	b.Run("AtomicInt", func(b *testing.B) {
		var counter int64
		
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// Simulate atomic operation
				counter++
			}
		})
	})
}

// BenchmarkRingBuffer benchmarks the ring buffer implementation
func BenchmarkRingBuffer(b *testing.B) {
	b.Run("RingBuffer_Small", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			inCh := make(chan int, 1)
			outCh := make(chan int, 4)
			done := make(chan bool)
			
			// Ring buffer goroutine
			go func() {
				defer close(outCh)
				for v := range inCh {
					select {
					case outCh <- v:
					default:
						select {
						case <-outCh:
						default:
						}
						outCh <- v
					}
				}
			}()
			
			// Producer
			go func() {
				defer close(inCh)
				for j := 0; j < 10; j++ {
					select {
					case inCh <- j:
					default:
					}
				}
			}()
			
			// Consumer
			go func() {
				count := 0
				for range outCh {
					count++
					if count >= 10 {
						break
					}
				}
				done <- true
			}()
			
			<-done
		}
	})
	
	b.Run("RingBuffer_Large", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			inCh := make(chan int, 1)
			outCh := make(chan int, 64)
			done := make(chan bool)
			
			// Ring buffer goroutine
			go func() {
				defer close(outCh)
				for v := range inCh {
					select {
					case outCh <- v:
					default:
						select {
						case <-outCh:
						default:
						}
						outCh <- v
					}
				}
			}()
			
			// Producer
			go func() {
				defer close(inCh)
				for j := 0; j < 10; j++ {
					select {
					case inCh <- j:
					default:
					}
				}
			}()
			
			// Consumer
			go func() {
				count := 0
				for range outCh {
					count++
					if count >= 10 {
						break
					}
				}
				done <- true
			}()
			
			<-done
		}
	})
	
	b.Run("SimpleBufferedChannel", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ch := make(chan int, 64)
			
			// Producer
			go func() {
				defer close(ch)
				for j := 0; j < 100; j++ {
					select {
					case ch <- j:
					default:
						// Drop if full
					}
				}
			}()
			
			// Consumer
			for range ch {
			}
		}
	})
}

// BenchmarkGoogleSearchPattern benchmarks concurrent vs sequential search
func BenchmarkGoogleSearchPattern(b *testing.B) {
	type Result string
	type Search func(query string) Result
	
	fakeSearch := func(kind string) Search {
		return func(query string) Result {
			time.Sleep(1 * time.Millisecond) // Simulate network delay
			return Result(fmt.Sprintf("%s result for %q", kind, query))
		}
	}
	
	Web := fakeSearch("web")
	Image := fakeSearch("image")
	Video := fakeSearch("video")
	
	b.Run("Sequential", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var results []Result
			results = append(results, Web("golang"))
			results = append(results, Image("golang"))
			results = append(results, Video("golang"))
		}
	})
	
	b.Run("Concurrent", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ch := make(chan Result, 3)
			
			go func() { ch <- Web("golang") }()
			go func() { ch <- Image("golang") }()
			go func() { ch <- Video("golang") }()
			
			var results []Result
			for j := 0; j < 3; j++ {
				results = append(results, <-ch)
			}
		}
	})
	
	b.Run("ConcurrentWithTimeout", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ch := make(chan Result, 3)
			
			go func() { ch <- Web("golang") }()
			go func() { ch <- Image("golang") }()
			go func() { ch <- Video("golang") }()
			
			var results []Result
			timeout := time.After(5 * time.Millisecond)
			
			for j := 0; j < 3; j++ {
				select {
				case result := <-ch:
					results = append(results, result)
				case <-timeout:
					goto done
				}
			}
			done:
		}
	})
}