package dispatcher

import (
	"errors"
	"sync"
	"time"

	"code.cloudfoundry.org/clock"
	"github.com/kihamo/go-workers/task"
	"github.com/kihamo/go-workers/worker"
)

const (
	DispatcherStatusWait = int64(iota)
	DispatcherStatusProcess
)

type Dispatcher struct {
	mutex sync.RWMutex
	wg    sync.WaitGroup
	clock clock.Clock

	workers *Workers
	tasks   *Tasks

	status   int64
	doneTask []chan task.Tasker // каналы уведомления о завершении выполнения заданий

	doneWorker       chan worker.Worker // канал уведомления о завершении рабочего
	quitDoWorkerDone chan bool

	allowProcessing       chan bool // канал для блокировки выполнения новых задач для случая, когда все исполнители заняты
	quitDoAllowProcessing chan bool

	quitDispatcher chan bool
}

func NewDispatcher() *Dispatcher {
	return NewDispatcherWithClock(clock.NewClock())
}

func NewDispatcherWithClock(c clock.Clock) *Dispatcher {
	return &Dispatcher{
		clock: c,

		workers: NewWorkers(),
		tasks:   NewTasks(),

		status:   DispatcherStatusWait,
		doneTask: []chan task.Tasker{},

		doneWorker:       make(chan worker.Worker),
		quitDoWorkerDone: make(chan bool, 1),

		allowProcessing:       make(chan bool, 1),
		quitDoAllowProcessing: make(chan bool, 1),

		quitDispatcher: make(chan bool, 1),
	}
}

func (d *Dispatcher) Run() error {
	if d.GetStatus() == DispatcherStatusProcess {
		return errors.New("Dispatcher is running")
	}

	d.wg.Add(1)
	go d.doWorkerDone()

	d.wg.Add(1)
	go d.doAllowProcessing()

	d.setStatus(DispatcherStatusProcess)

	for {
		select {
		case <-d.quitDispatcher:
			d.quitDoAllowProcessing <- true
			d.quitDoWorkerDone <- true

			d.wg.Wait()
			d.setStatus(DispatcherStatusWait)
			return nil
		}
	}
}

func (d *Dispatcher) doWorkerDone() {
	defer d.wg.Done()

	for {
		select {
		case w := <-d.doneWorker:
			t := w.GetTask()
			d.GetTasks().Remove(t)

			w.Reset()
			d.GetWorkers().Update(w)

			go d.notifyAllowProcessing()

			d.mutex.RLock()
			if len(d.doneTask) > 0 {
				// может держать процесс заблокированным, если канал буферизированный и с малым размером
				go func(l []chan task.Tasker) {
					for _, c := range l {
						c <- t
					}
				}(d.doneTask)
			}
			d.mutex.RUnlock()

			if repeats := t.GetRepeats(); repeats == -1 || t.GetAttempts() < repeats {
				t.SetStatus(task.TaskStatusRepeatWait)
				d.AddTask(t)
			}

		case <-d.quitDoAllowProcessing:
			return
		}
	}
}

func (d *Dispatcher) doAllowProcessing() {
	defer d.wg.Done()

	// таймер на случай, если процесс залипнет или заполнение очередей идет не черз диспетчер
	ticker := d.GetClock().NewTicker(time.Second)

	for {
		select {
		case <-d.allowProcessing:
			for d.GetTasks().HasWait() && d.GetWorkers().HasWait() {
				t := d.GetTasks().GetWait()
				w := d.GetWorkers().GetWait()

				changeStatus := make(chan int64, 1)
				w.SetChangeStatusChannel(changeStatus)

				d.wg.Add(1)
				go func() {
					defer d.wg.Done()

					for {
						select {
						case s := <-changeStatus:
							if s == worker.WorkerStatusProcess {
								d.addWorker(w)
								w.SetChangeStatusChannel(nil)
								w.SendTask(t)

								return
							}
						}
					}
				}()

				d.wg.Add(1)
				go func() {
					defer d.wg.Done()

					w.Run()
				}()
			}

		case <-ticker.C():
			d.notifyAllowProcessing()

		case <-d.quitDoWorkerDone:
			return
		}
	}
}

func (d *Dispatcher) notifyAllowProcessing() {
	if d.GetStatus() == DispatcherStatusProcess && len(d.allowProcessing) == 0 {
		d.allowProcessing <- true
	}
}

func (d *Dispatcher) addWorker(w worker.Worker) {
	d.GetWorkers().Add(w)
	d.notifyAllowProcessing()
}

func (d *Dispatcher) addTask(t task.Tasker) {
	d.GetTasks().Add(t)
	d.notifyAllowProcessing()
}

func (d *Dispatcher) AddWorker() worker.Worker {
	w := worker.NewWorkmanWithClock(d.doneWorker, d.GetClock())
	d.addWorker(w)

	return w
}

func (d *Dispatcher) AddTask(t task.Tasker) {
	duration := t.GetDuration()
	if duration > 0 {
		timer := d.GetClock().NewTimer(duration)
		go func() {
			<-timer.C()
			d.addTask(t)
		}()
	} else {
		d.addTask(t)
	}
}

func (d *Dispatcher) AddNamedTaskByFunc(n string, f task.TaskFunction, a ...interface{}) task.Tasker {
	task := task.NewTask(f, a...)

	if n != "" {
		task.SetName(n)
	}

	d.AddTask(task)

	return task
}

func (d *Dispatcher) AddTaskByFunc(f task.TaskFunction, a ...interface{}) task.Tasker {
	return d.AddNamedTaskByFunc("", f, a...)
}

func (d *Dispatcher) GetWorkers() *Workers {
	return d.workers
}

func (d *Dispatcher) GetTasks() *Tasks {
	return d.tasks
}

func (d *Dispatcher) Kill() error {
	if d.GetStatus() == DispatcherStatusProcess {
		d.quitDispatcher <- true
		return nil
	}

	return errors.New("Dispatcher isn't running")
}

func (d *Dispatcher) GetStatus() int64 {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return d.status
}

func (d *Dispatcher) setStatus(s int64) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.status = s
}

func (d *Dispatcher) GetClock() clock.Clock {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return d.clock
}

func (d *Dispatcher) SetTaskDoneChannel(c chan task.Tasker) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.doneTask = []chan task.Tasker{c}
}

func (d *Dispatcher) AddTaskDoneChannel(c chan task.Tasker) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.doneTask = append(d.doneTask, c)
}
