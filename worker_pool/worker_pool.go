package worker_pool

import (
	"sync"
)

const size = 1024

type WorkerPool struct {
	tasks     chan func()
	wg        sync.WaitGroup
	runningWg sync.WaitGroup
	stopOnce  sync.Once
	stopCh    chan struct{}
	accepting bool
	hardStop  bool
	mu        sync.Mutex
}

func NewWorkerPool(numberOfWorkers int) *WorkerPool {
	if numberOfWorkers <= 0 {
		panic("numberOfWorkers must be > 0")
	}
	wp := &WorkerPool{
		tasks:     make(chan func(), size),
		stopCh:    make(chan struct{}),
		accepting: true,
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
			wp.mu.Lock()
			if wp.hardStop {
				wp.mu.Unlock()
				wp.wg.Done()
				continue
			}
			wp.runningWg.Add(1)
			wp.mu.Unlock()

			task()
			wp.runningWg.Done()
			wp.wg.Done()
		}
	}
}

// Submit - добавить таску в воркер пул
func (wp *WorkerPool) Submit(task func()) {
	wp.mu.Lock()
	if !wp.accepting {
		wp.mu.Unlock()
		panic("Submit called after Stop/StopWait")
	}
	wp.mu.Unlock()

	wp.wg.Add(1)
	select {
	case wp.tasks <- task:
	case <-wp.stopCh:
		wp.wg.Done()
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
		wp.accepting = false
		wp.hardStop = true
		wp.mu.Unlock()

		close(wp.stopCh)
		wp.runningWg.Wait()
	})
}

// StopWait - остановить воркер пул, дождаться выполнения всех тасок, даже тех, что не начали выполняться, но лежат в очереди
func (wp *WorkerPool) StopWait() {
	wp.stopOnce.Do(func() {
		wp.mu.Lock()
		wp.accepting = false
		wp.mu.Unlock()

		close(wp.tasks)
		wp.wg.Wait()
		close(wp.stopCh)
	})
}
