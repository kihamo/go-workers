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
					} else {
						task.setStatus(taskStatusSuccess)
					}

					repeats := task.GetRepeats()
					if repeats == -1 || task.GetAttempts() < repeats {
						task.setStatus(taskStatusRepeatWait)
						repeatQueue <- task
					}

					w.waitGroup.Done()
					done <- w
				}()

				nextAttempt := task.GetAttempts() + 1

				task.setStatus(taskStatusProcess)
				task.setAttempts(nextAttempt)

				newRepeats, newDuration := task.GetFunction()(nextAttempt, task.GetArguments()...)
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
