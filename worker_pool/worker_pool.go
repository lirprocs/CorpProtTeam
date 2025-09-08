package worker_pool

import (
	"sync"
)

type WorkerPool struct {
	tasks    chan func()
	wg       sync.WaitGroup
	stopOnce sync.Once
	stopCh   chan struct{}
	stopped  bool
	mu       sync.Mutex
}

func NewWorkerPool(size, numberOfWorkers int) *WorkerPool {
	if numberOfWorkers <= 0 {
		panic("numberOfWorkers must be > 0")
	}
	if size <= 0 {
		panic("size must be > 0")
	}
	wp := &WorkerPool{
		tasks:  make(chan func(), size),
		stopCh: make(chan struct{}),
	}

	for i := 0; i < numberOfWorkers; i++ {
		go wp.worker()
	}

	return wp
}

func (wp *WorkerPool) worker() {
	for {
		select {
		case <-wp.stopCh:
			return
		case task, ok := <-wp.tasks:
			if !ok {
				return
			}
			task()
			wp.wg.Done()
		}
	}
}

// Submit - добавить таску в воркер пул
func (wp *WorkerPool) Submit(task func()) {
	wp.mu.Lock()
	if wp.stopped {
		wp.mu.Unlock()
		panic("Submit called after Stop/StopWait")
	}
	wp.mu.Unlock()

	select {
	case wp.tasks <- task:
		wp.wg.Add(1)
	case <-wp.stopCh:
		panic("Submit called after Stop/StopWait")
	}
}

// SubmitWait - добавить таску в воркер пул и дождаться окончания ее выполнения
func (wp *WorkerPool) SubmitWait(task func()) {
	done := make(chan struct{})
	wp.Submit(func() {
		task()
		close(done)
	})
	<-done
}

// Stop - остановить воркер пул, дождаться выполнения только тех тасок, которые выполняются сейчас
func (wp *WorkerPool) Stop() {
	wp.stopOnce.Do(func() {
		wp.mu.Lock()
		wp.stopped = true
		wp.mu.Unlock()

		close(wp.stopCh)
		close(wp.tasks)
	})
}

// StopWait - остановить воркер пул, дождаться выполнения всех тасок, даже тех, что не начали выполняться, но лежат в очереди
func (wp *WorkerPool) StopWait() {
	wp.stopOnce.Do(func() {
		wp.mu.Lock()
		wp.stopped = true
		wp.mu.Unlock()

		close(wp.tasks)
		wp.wg.Wait()
		close(wp.stopCh)
	})
}
