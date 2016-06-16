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
	newTask   chan *Task
	quit      chan bool
}

func NewWorker() *Worker {
	id := uuid.New()

	return &Worker{
		id:        id,
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

			if task.GetStatus() != TaskStatusRepeatWait {
				task.setAttempts(0)
			}

			func() {
				defer func() {
					task.setFinishedTime(time.Now())

					if err := recover(); err != nil {
						task.setStatus(TaskStatusFail)
						task.setLastError(err)
					}

					repeats := task.GetRepeats()
					if repeats == -1 || task.GetAttempts() < repeats {
						task.setStatus(TaskStatusRepeatWait)
						repeatQueue <- task
					}

					w.waitGroup.Done()
					w.setStatus(WorkerStatusWait)
					done <- w
				}()

				task.setStatus(TaskStatusProcess)
				task.setAttempts(task.GetAttempts() + 1)

				newRepeats, newDuration := w.execute(task)
				task.SetRepeats(newRepeats)
				task.SetDuration(newDuration)
			}()

		case <-w.quit:
			w.waitGroup.Wait()
			w.setStatus(WorkerStatusWait)
			return
		}
	}
}

func (w *Worker) execute(task *Task) (int64, time.Duration) {
	quit := make(chan bool, 1)
	timeout := task.GetTimeout()

	// execute with timeout
	if timeout > 0 {
		result := make(chan []interface{}, 1)

		go func() {
			defer func() {
				if err := recover(); err != nil {
					result <- []interface{}{nil, nil, err}
				}
			}()

			repeats, duration := task.GetFunction()(task.GetAttempts(), quit, task.GetArguments()...)
			result <- []interface{}{repeats, duration, nil}
		}()

		for {
			select {
			case r := <-result:
				if r[2] != nil {
					panic(r[2])
				}

				task.setStatus(TaskStatusSuccess)
				return r[0].(int64), r[1].(time.Duration)

			case <-time.After(timeout):
				quit <- true
				task.setStatus(TaskStatusFailByTimeout)
				return task.GetRepeats(), task.GetDuration()
			}
		}
	}

	// execute without timeout
	repeats, duration := task.GetFunction()(task.GetAttempts(), quit, task.GetArguments()...)
	task.setStatus(TaskStatusSuccess)

	return repeats, duration
}

func (w *Worker) Kill() {
	w.quit <- true
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
