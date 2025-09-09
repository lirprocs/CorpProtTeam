package worker_pool

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewWorkerPoolValidation(t *testing.T) {
	tests := []struct {
		name            string
		size            int
		numberOfWorkers int
		shouldPanic     bool
		panicMessage    string
	}{
		{
			name:            "Valid parameters - normal case",
			numberOfWorkers: 5,
			shouldPanic:     false,
		},
		{
			name:            "Invalid - zero workers",
			numberOfWorkers: 0,
			shouldPanic:     true,
			panicMessage:    "numberOfWorkers must be > 0",
		},
		{
			name:            "Invalid - negative workers large",
			numberOfWorkers: -1,
			shouldPanic:     true,
			panicMessage:    "numberOfWorkers must be > 0",
		},
		{
			name:            "Valid - large parameters",
			numberOfWorkers: 100,
			shouldPanic:     false,
		},
		{
			name:            "Valid - minimal valid parameters",
			numberOfWorkers: 1,
			shouldPanic:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.shouldPanic {
					if r == nil {
						t.Errorf("Expected panic, but got none")
						return
					}
					if tt.panicMessage != "" && r != tt.panicMessage {
						t.Errorf("Expected panic message %q, got %q", tt.panicMessage, r)
					}
				} else {
					if r != nil {
						t.Errorf("Unexpected panic: %v", r)
					}
				}
			}()

			pool := NewWorkerPool(tt.numberOfWorkers)

			if !tt.shouldPanic {
				if pool == nil {
					t.Error("Expected pool to be created, got nil")
				}

				pool.Stop()
			}
		})
	}
}

func TestSubmit(t *testing.T) {
	var counter int32
	numTask := 10

	wp := NewWorkerPool(5)

	for i := 0; i < numTask; i++ {
		wp.Submit(func() {
			atomic.AddInt32(&counter, 1)
		})
	}

	wp.StopWait()
	if counter != int32(numTask) {
		t.Errorf("TestSubmit: tasks did not run")
	}
}

func TestParallelExecution(t *testing.T) {
	var counter int32
	numTask := 10

	wp := NewWorkerPool(10)
	start := time.Now()
	for i := 0; i < numTask; i++ {
		wp.Submit(func() {
			atomic.AddInt32(&counter, 1)
			time.Sleep(1000 * time.Millisecond)
		})
	}
	wp.StopWait()
	stop := time.Since(start)

	if stop > 2000*time.Millisecond {
		t.Errorf("TestParallelExecution: tasks did not run in parallel, took %v", stop)
	}
}

func TestSubmitAfterStop(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("TestSubmitAfterStop: expected panic when submitting after Stop, but got none")
		}
	}()

	wp := NewWorkerPool(2)
	wp.Stop()
	wp.Submit(func() {})
}

func TestSubmitWait(t *testing.T) {
	var counter int32
	numTask := 3

	wp := NewWorkerPool(2)
	for i := 0; i < numTask; i++ {
		wp.SubmitWait(func() {
			atomic.AddInt32(&counter, 1)
			time.Sleep(1000 * time.Millisecond)
		})
	}
	wp.Stop()

	if counter != int32(numTask) {
		t.Errorf("TestSubmitWait: expected %d, got %d", numTask, counter)
	}
}

func TestStop(t *testing.T) {
	var counter int32
	numTask := 10

	wp := NewWorkerPool(2)

	for i := 0; i < numTask; i++ {
		wp.Submit(func() {
			time.Sleep(1000 * time.Millisecond)
			fmt.Println("Done")
			atomic.AddInt32(&counter, 1)
		})
	}

	time.Sleep(50 * time.Millisecond)

	wp.Stop()

	if counter != int32(2) {
		t.Errorf("TestStop: expected %d, got %d", 2, counter)
	}
}

func TestStopAfterStop(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("TestStopAfterStop: expected no panic on second Stop, but got: %v", r)
		}
	}()

	wp := NewWorkerPool(2)

	wp.Submit(func() {})

	wp.Stop()
	wp.Stop()
}

func TestStopWait(t *testing.T) {
	var counter int32
	numTask := 5

	wp := NewWorkerPool(2)

	for i := 0; i < numTask; i++ {
		wp.Submit(func() {
			time.Sleep(20 * time.Millisecond)
			atomic.AddInt32(&counter, 1)
		})
	}

	wp.StopWait()

	if counter != 5 {
		t.Errorf("TestStopWait: expected %d, got %d", numTask, counter)
	}
}

func TestStopWaitAfterStopWait(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("TestStopWaitAfterStopWait: expected no panic on second Stop, but got: %v", r)
		}
	}()

	wp := NewWorkerPool(2)

	wp.Submit(func() {})

	wp.StopWait()
	wp.StopWait()
}

func TestStopWaitAfterStop(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("TestStopWaitAfterStop: expected no panic on second Stop, but got: %v", r)
		}
	}()

	wp := NewWorkerPool(2)

	wp.Submit(func() {})

	wp.Stop()
	wp.StopWait()
}

func TestStopAfterStopWait(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("TestStopAfterStopWait: expected no panic on second Stop, but got: %v", r)
		}
	}()

	wp := NewWorkerPool(2)

	wp.Submit(func() {})

	wp.StopWait()
	wp.Stop()
}

func TestConcurrentSubmitWait(t *testing.T) {
	var counter int32
	numTasks := 20
	wp := NewWorkerPool(5)

	var wg sync.WaitGroup
	wg.Add(numTasks)

	for i := 0; i < numTasks; i++ {
		go func() {
			defer wg.Done()
			wp.SubmitWait(func() {
				time.Sleep(50 * time.Millisecond)
				atomic.AddInt32(&counter, 1)
			})
		}()
	}

	wg.Wait()
	wp.Stop()

	if counter != int32(numTasks) {
		t.Errorf("TestConcurrentSubmitWait: expected %d, got %d", numTasks, counter)
	}
}

func TestStopWithRemainingTasks(t *testing.T) {
	var counter int32
	numTasks := 10

	wp := NewWorkerPool(2)

	for i := 0; i < numTasks; i++ {
		wp.Submit(func() {
			time.Sleep(100 * time.Millisecond)
			atomic.AddInt32(&counter, 1)
		})
	}

	time.Sleep(150 * time.Millisecond)

	wp.Stop()

	if counter >= int32(numTasks) {
		t.Errorf("TestStopWithRemainingTasks: expected some tasks to remain unexecuted, got %d/%d", counter, numTasks)
	}
}
