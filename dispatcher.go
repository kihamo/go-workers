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

	team        *Team
	workersBusy int

	newQueue     chan *Task // очередь новых заданий
	executeQueue chan *Task // очередь выполняемых заданий
	repeatQueue  chan *Task // канал уведомления о повторном выполнении заданий

	done            chan *Worker // канал уведомления о завершении выполнения заданий
	quit            chan bool    // канал для завершения диспетчера
	allowProcessing chan bool    // канал для блокировки выполнения новых задач для случая, когда все исполнители заняты
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		waitGroup: new(sync.WaitGroup),

		team:        NewTeam(),
		workersBusy: 0,

		newQueue:     make(chan *Task),
		executeQueue: make(chan *Task),
		repeatQueue:  make(chan *Task),

		done:            make(chan *Worker),
		quit:            make(chan bool),
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

	heap.Init(d.team)

	for {
		select {
		// пришел новый таск на выполнение от flow контроллера
		case task := <-d.executeQueue:
			worker := heap.Pop(d.team).(*Worker)
			worker.sendTask(task)
			heap.Push(d.team, worker)

			// проверяем есть ли еще свободные исполнители для задач
			if d.workersBusy++; d.workersBusy < d.team.Len() {
				d.allowProcessing <- true
			}

		// пришло уведомление, что рабочий закончил выполнение задачи
		case worker := <-d.done:
			heap.Remove(d.team, d.team.GetIndexById(worker.id))
			heap.Push(d.team, worker)

			// проверяем не освободился ли какой-нибудь исполнитель
			if d.workersBusy--; d.workersBusy == d.team.Len()-1 {
				d.allowProcessing <- true
			}

		// пришло уведомление, что необходимо повторить задачу
		case task := <-d.repeatQueue:
			d.AddTask(task)

		case <-d.quit:
			d.waitGroup.Wait()
			return
		}
	}
}

func (d *Dispatcher) AddWorker() {
	w := NewWorker()

	d.waitGroup.Add(1)
	go func() {
		defer d.waitGroup.Done()

		w.process(d.done, d.repeatQueue)
	}()

	heap.Push(d.team, w)
}

func (d *Dispatcher) AddTask(task *Task) {
	time.AfterFunc(task.GetDuration(), func() {
		d.newQueue <- task
	})
}

func (d *Dispatcher) AddNamedTaskByFunc(name string, fn TaskFunction, args ...interface{}) *Task {
	task := NewTask(name, time.Duration(0), time.Duration(0), 1, fn, args)
	d.AddTask(task)

	return task
}

func (d *Dispatcher) AddTaskByFunc(fn TaskFunction, args ...interface{}) *Task {
	return d.AddNamedTaskByFunc(runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name(), fn, args...)
}

func (d *Dispatcher) Kill() {
	d.quit <- true
}

func (d *Dispatcher) GetTeam() *Team {
	return d.team
}
