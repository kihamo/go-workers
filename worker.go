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
	mutex     sync.RWMutex
	waitGroup *sync.WaitGroup

	id        string
	status    int64
	createdAt time.Time
	task      *Task
	newTask   chan *Task
	quit      chan bool
}

func NewWorker() *Worker {
	return &Worker{
		id:        uuid.New(),
		status:    WorkerStatusWait,
		createdAt: time.Now(),
		waitGroup: new(sync.WaitGroup),
		newTask:   make(chan *Task, 1),
		quit:      make(chan bool),
	}
}

func (w *Worker) process(done chan *Worker, repeatQueue chan *Task) {
	for {
		select {
		case task := <-w.newTask:
			w.waitGroup.Add(1)

			w.setStatus(WorkerStatusBusy)
			w.setTask(task)

			task.setLastError(nil)
			if task.GetStatus() != TaskStatusRepeatWait {
				task.setAttempts(0)
			}

			func() {
				task.setStatus(TaskStatusProcess)
				task.setAttempts(task.GetAttempts() + 1)

				w.execute(task)

				repeats := task.GetRepeats()
				if repeats == -1 || task.GetAttempts() < repeats {
					task.setStatus(TaskStatusRepeatWait)
					repeatQueue <- task
				}

				w.waitGroup.Done()
				w.setStatus(WorkerStatusWait)
				done <- w
			}()

		case <-w.quit:
			w.waitGroup.Wait()
			w.setStatus(WorkerStatusWait)
			return
		}
	}
}

func (w *Worker) execute(task *Task) {
	defer func() {
		task.setFinishedTime(time.Now())

		if err := recover(); err != nil {
			task.setStatus(TaskStatusFail)
			task.setLastError(err)
		}
	}()

	quit := make(chan bool, 1)
	timeout := task.GetTimeout()

	// execute with timeout
	if timeout > 0 {
		resultChan := make(chan []interface{}, 1)
		errorChan := make(chan interface{}, 1)

		go func() {
			defer func() {
				if err := recover(); err != nil {
					errorChan <- err
				}
			}()

			repeats, duration := task.GetFunction()(task.GetAttempts(), quit, task.GetArguments()...)
			resultChan <- []interface{}{repeats, duration}
		}()

		for {
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

			case <-time.After(timeout):
				quit <- true

				task.setStatus(TaskStatusFailByTimeout)
				return
			}
		}
	}

	// execute without timeout
	newRepeats, newDuration := task.GetFunction()(task.GetAttempts(), quit, task.GetArguments()...)

	task.setStatus(TaskStatusSuccess)
	task.SetRepeats(newRepeats)
	task.SetDuration(newDuration)
}

func (w *Worker) Kill() {
	w.quit <- true
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
