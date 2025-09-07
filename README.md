Необходимо реализовать WorkerPool и покрыть тестами.             [![Test](https://github.com/lirprocs/CorpProtTeam/actions/workflows/test.yaml/badge.svg)](https://github.com/lirprocs/CorpProtTeam/actions/workflows/test.yaml)

package worker_pool

type WorkerPool struct {
}

func NewWorkerPool(numberOfWorkers int) *WorkerPool {
	return &WorkerPool{}
}

// Submit - добавить таску в воркер пул
func (wp *WorkerPool) Submit(task func()) {

}

// SubmitWait - добавить таску в воркер пул и дождаться окончания ее выполнения
func (wp *WorkerPool) SubmitWait(task func()) {

}

// Stop - остановить воркер пул, дождаться выполнения только тех тасок, которые выполняются сейчас
func (wp *WorkerPool) Stop() {

}

// StopWait - остановить воркер пул, дождаться выполнения всех тасок, даже тех, что не начали выполняться, но лежат в очереди
func (wp *WorkerPool) StopWait() {

}
