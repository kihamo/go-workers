package dispatcher

import (
	"container/heap"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/kihamo/go-workers/collection"
	"github.com/kihamo/go-workers/task"
	"github.com/kihamo/go-workers/worker"
)

type Dispatcher struct {
	mutex     sync.RWMutex
	waitGroup *sync.WaitGroup

	workers *collection.Workers
	tasks   *collection.Tasks

	workersBusy int

	newQueue     chan task.Tasker // очередь новых заданий
	executeQueue chan task.Tasker // очередь выполняемых заданий

	done            chan worker.Worker // канал уведомления о завершении выполнения заданий
	quit            chan bool          // канал для завершения диспетчера
	allowProcessing chan bool          // канал для блокировки выполнения новых задач для случая, когда все исполнители заняты
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		waitGroup: new(sync.WaitGroup),

		workers: collection.NewWorkers(),
		tasks:   collection.NewTasks(),

		workersBusy: 0,

		newQueue:     make(chan task.Tasker),
		executeQueue: make(chan task.Tasker),

		done:            make(chan worker.Worker),
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
		case t := <-d.executeQueue:
			worker := heap.Pop(d.workers).(worker.Worker)
			d.runWorker(worker)
			worker.SendTask(t)
			heap.Push(d.workers, worker)

			// проверяем есть ли еще свободные исполнители для задач
			if d.workersBusy++; d.workersBusy < d.workers.Len() {
				d.allowProcessing <- true
			}

		// пришло уведомление, что рабочий закончил выполнение задачи
		case w := <-d.done:
			heap.Remove(d.workers, d.workers.GetIndexById(w.GetId()))
			heap.Push(d.workers, w)

			t := w.GetTask()

			d.tasks.RemoveById(t.GetId())
			w.Reset()

			repeats := t.GetRepeats()
			if repeats == -1 || t.GetAttempts() < repeats {
				t.SetStatus(task.TaskStatusRepeatWait)
				d.AddTask(t)
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

func (d *Dispatcher) AddWorker() worker.Worker {
	w := worker.NewWorker(d.done)
	heap.Push(d.workers, w)

	return w
}

func (d *Dispatcher) runWorker(w worker.Worker) {
	d.waitGroup.Add(1)
	go func() {
		defer d.waitGroup.Done()
		w.Run()
	}()
}

func (d *Dispatcher) GetWorkers() *collection.Workers {
	return d.workers
}

func (d *Dispatcher) AddTask(t task.Tasker) {
	time.AfterFunc(t.GetDuration(), func() {
		d.tasks.Push(t)
		d.newQueue <- t
	})
}

func (d *Dispatcher) AddNamedTaskByFunc(name string, fn task.TaskFunction, args ...interface{}) task.Tasker {
	task := task.NewTask(name, 0, 0, 1, fn, args)
	d.AddTask(task)

	return task
}

func (d *Dispatcher) AddTaskByFunc(fn task.TaskFunction, args ...interface{}) task.Tasker {
	return d.AddNamedTaskByFunc(runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name(), fn, args...)
}

func (d *Dispatcher) GetTasks() *collection.Tasks {
	return d.tasks
}

func (d *Dispatcher) Kill() {
	d.quit <- true
}