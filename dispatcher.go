package workers

import (
	"container/heap"
	"reflect"
	"runtime"
	"sync"
	"time"
)

type Dispatcher struct {
	mutex     sync.RWMutex
	waitGroup *sync.WaitGroup

	workers *Workers
	tasks   *Tasks

	workersBusy int

	newQueue     chan *Task // очередь новых заданий
	executeQueue chan *Task // очередь выполняемых заданий

	done            chan *Worker // канал уведомления о завершении выполнения заданий
	quit            chan bool    // канал для завершения диспетчера
	allowProcessing chan bool    // канал для блокировки выполнения новых задач для случая, когда все исполнители заняты
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		waitGroup: new(sync.WaitGroup),

		workers: NewWorkers(),
		tasks:   NewTasks(),

		workersBusy: 0,

		newQueue:     make(chan *Task),
		executeQueue: make(chan *Task),

		done:            make(chan *Worker),
		quit:            make(chan bool, 1),
		allowProcessing: make(chan bool),
	}
}

func (d *Dispatcher) Run() {
	// отслеживание квоты на занятость исполнителей
	go func() {
		for {
			d.executeQueue <- <-d.newQueue

			<-d.allowProcessing
		}
	}()

	heap.Init(d.workers)

	for {
		select {
		// пришел новый таск на выполнение от flow контроллера
		case task := <-d.executeQueue:
			worker := heap.Pop(d.workers).(*Worker)
			d.runWorker(worker)
			worker.sendTask(task)
			heap.Push(d.workers, worker)

			// проверяем есть ли еще свободные исполнители для задач
			if d.workersBusy++; d.workersBusy < d.workers.Len() {
				d.allowProcessing <- true
			}

		// пришло уведомление, что рабочий закончил выполнение задачи
		case worker := <-d.done:
			heap.Remove(d.workers, d.workers.GetIndexById(worker.GetId()))
			heap.Push(d.workers, worker)

			task := worker.GetTask()

			d.tasks.removeById(task.GetId())
			worker.setTask(nil)

			repeats := task.GetRepeats()
			if repeats == -1 || task.GetAttempts() < repeats {
				task.setStatus(TaskStatusRepeatWait)
				d.AddTask(task)
			}

			// проверяем не освободился ли какой-нибудь исполнитель
			if d.workersBusy--; d.workersBusy == d.workers.Len()-1 {
				d.allowProcessing <- true
			}

		case <-d.quit:
			d.waitGroup.Wait()
			return
		}
	}
}

func (d *Dispatcher) AddWorker() *Worker {
	w := NewWorker(d.done)
	heap.Push(d.workers, w)

	return w
}

func (d *Dispatcher) runWorker(worker *Worker) {
	d.waitGroup.Add(1)
	go func() {
		defer d.waitGroup.Done()
		worker.run()
	}()
}

func (d *Dispatcher) GetWorkers() *Workers {
	return d.workers
}

func (d *Dispatcher) AddTask(task *Task) {
	time.AfterFunc(task.GetDuration(), func() {
		d.tasks.Push(task)
		d.newQueue <- task
	})
}

func (d *Dispatcher) AddNamedTaskByFunc(name string, fn TaskFunction, args ...interface{}) *Task {
	task := NewTask(name, 0, 0, 1, fn, args)
	d.AddTask(task)

	return task
}

func (d *Dispatcher) AddTaskByFunc(fn TaskFunction, args ...interface{}) *Task {
	return d.AddNamedTaskByFunc(runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name(), fn, args...)
}

func (d *Dispatcher) GetTasks() *Tasks {
	return d.tasks
}

func (d *Dispatcher) Kill() {
	d.quit <- true
}
