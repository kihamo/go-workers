package workers

import (
	"sync"
	"time"

	"github.com/pborman/uuid"
)

const (
	workerStatusWait = iota
	workerStatusBusy
)

type Worker struct {
	id        string
	status    int
	created   time.Time
	waitGroup *sync.WaitGroup
	newTask   chan *Task
	quit      chan bool // канал для завершения исполнителя
}

func NewWorker() *Worker {
	id := uuid.New()

	return &Worker{
		id:        id,
		status:    workerStatusWait,
		created:   time.Now(),
		waitGroup: new(sync.WaitGroup),
		newTask:   make(chan *Task, 1),
		quit:      make(chan bool),
	}
}

// kill worker shutdown
func (w *Worker) Kill() {
	w.quit <- true
}

// work выполняет задачу
func (w *Worker) process(done chan *Worker, repeatQueue chan *Task) {
	for {
		select {
		// пришло новое задание на выполнение
		case task := <-w.newTask:
			w.waitGroup.Add(1)

			func() {
				defer func() {
					task.setFinishedTime(time.Now())

					if err := recover(); err != nil {
						task.setStatus(taskStatusFail)
						task.setLastError(err)
					}

					repeats := task.GetRepeats()
					if repeats == -1 || task.GetAttempts() < repeats {
						task.setStatus(taskStatusRepeatWait)
						repeatQueue <- task
					}

					w.waitGroup.Done()
					done <- w
				}()

				task.setStatus(taskStatusProcess)
				task.setAttempts(task.GetAttempts() + 1)

				newRepeats, newDuration := w.execute(task)
				task.SetRepeats(newRepeats)
				task.SetDuration(newDuration)
			}()

		// пришел сигнал на завершение исполнителя
		case <-w.quit:
			// ждем завершения текущего задания, если таковое есть и выходим
			w.waitGroup.Wait()
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

				task.setStatus(taskStatusSuccess)
				return r[0].(int64), r[1].(time.Duration)

			case <-time.After(timeout):
				quit <- true
				task.setStatus(taskStatusFailByTimeout)
				return task.GetRepeats(), task.GetDuration()
			}
		}
	}

	// execute without timeout
	repeats, duration := task.GetFunction()(task.GetAttempts(), quit, task.GetArguments()...)
	task.setStatus(taskStatusSuccess)

	return repeats, duration
}
