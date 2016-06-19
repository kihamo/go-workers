package workers

import (
	"sync"
	"time"

	"github.com/pborman/uuid"
)

const (
	WorkerStatusWait = int64(iota)
	WorkerStatusBusy
)

type Worker struct {
	mutex sync.RWMutex

	id        string
	status    int64
	createdAt time.Time
	task      *Task
	newTask   chan *Task
	quit      chan bool
	kill      chan bool

	done chan *Worker
}

func NewWorker(done chan *Worker) *Worker {
	return &Worker{
		id:        uuid.New(),
		status:    WorkerStatusWait,
		createdAt: time.Now(),
		newTask:   make(chan *Task, 1),
		quit:      make(chan bool, 1),
		kill:      make(chan bool, 1),
		done:      done,
	}
}

func (w *Worker) run() {
	for {
		select {
		case task := <-w.newTask:
			go w.processTask(task)

		case <-w.kill:
			w.quit <- true
			w.setStatus(WorkerStatusWait)
			return
		}
	}
}

func (w *Worker) processTask(task *Task) {
	w.setStatus(WorkerStatusBusy)
	w.setTask(task)

	task.setLastError(nil)
	if task.GetStatus() != TaskStatusRepeatWait {
		task.setAttempts(0)
	}

	task.setStatus(TaskStatusProcess)
	task.setAttempts(task.GetAttempts() + 1)

	w.executeTask(task)

	w.setStatus(WorkerStatusWait)
	w.done <- w
}

func (w *Worker) executeTask(task *Task) {
	defer func() {
		task.setFinishedTime(time.Now())

		if err := recover(); err != nil {
			task.setStatus(TaskStatusFail)
			task.setLastError(err)
		}
	}()

	resultChan := make(chan []interface{}, 1)
	errorChan := make(chan interface{}, 1)
	quitChan := make(chan bool, 1)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				errorChan <- err
			}
		}()

		newRepeats, newDuration := task.GetFunction()(task.GetAttempts(), quitChan, task.GetArguments()...)
		resultChan <- []interface{}{newRepeats, newDuration}
	}()

	for {
		timeout := task.GetTimeout()

		if timeout > 0 {
			// execute witch timeout
			select {
			case r := <-resultChan:
				task.setStatus(TaskStatusSuccess)
				task.SetRepeats(r[0].(int64))
				task.SetDuration(r[1].(time.Duration))
				return

			case err := <-errorChan:
				task.setStatus(TaskStatusFail)
				task.setLastError(err)
				return

			case <-w.quit:
				quitChan <- true
				task.setStatus(TaskStatusKill)
				return

			case <-time.After(timeout):
				quitChan <- true
				task.setStatus(TaskStatusFailByTimeout)
				return
			}
		} else {
			// execute without timeout
			select {
			case r := <-resultChan:
				task.setStatus(TaskStatusSuccess)
				task.SetRepeats(r[0].(int64))
				task.SetDuration(r[1].(time.Duration))
				return

			case err := <-errorChan:
				task.setStatus(TaskStatusFail)
				task.setLastError(err)
				return

			case <-w.quit:
				quitChan <- true
				task.setStatus(TaskStatusKill)
				return
			}
		}
	}
}

func (w *Worker) Kill() {
	w.kill <- true
}

func (w *Worker) GetTask() *Task {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.task
}

func (w *Worker) setTask(task *Task) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.task = task
}

func (w *Worker) sendTask(task *Task) {
	w.newTask <- task
}

func (w *Worker) GetId() string {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.id
}

func (w *Worker) GetStatus() int64 {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.status
}

func (w *Worker) GetCreatedAt() time.Time {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.createdAt
}

func (w *Worker) setStatus(status int64) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.status = status
}
