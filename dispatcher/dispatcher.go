package dispatcher

import (
	"errors"
	"sync"
	"time"

	"code.cloudfoundry.org/clock"
	"github.com/kihamo/go-workers"
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

	status         int64
	listeners      *ListenerList
	listenersTasks *ListenerTasks

	doneWorker       chan worker.Worker // канал уведомления о завершении рабочего
	quitDoWorkerDone chan bool

	allowExecuteTasks    chan bool // канал для блокировки выполнения новых задач для случая, когда все исполнители заняты
	quitDoExecuteTasks   chan bool
	tickerDoExecuteTasks *workers.Ticker

	allowNotifyListeners    chan bool
	quitDoNotifyListeners   chan bool
	tickerDoNotifyListeners *workers.Ticker

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

		status:         DispatcherStatusWait,
		listeners:      NewListenerList(),
		listenersTasks: NewListenerTasks(),

		doneWorker:       make(chan worker.Worker),
		quitDoWorkerDone: make(chan bool, 1),

		allowExecuteTasks:    make(chan bool, 1),
		quitDoExecuteTasks:   make(chan bool, 1),
		tickerDoExecuteTasks: workers.NewTicker(time.Second),

		allowNotifyListeners:    make(chan bool, 1),
		quitDoNotifyListeners:   make(chan bool, 1),
		tickerDoNotifyListeners: workers.NewTicker(time.Second),

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
	go d.doExecuteTasks()

	d.wg.Add(1)
	go d.doNotifyListeners()

	d.setStatus(DispatcherStatusProcess)

	for {
		select {
		case <-d.quitDispatcher:
			d.quitDoExecuteTasks <- true
			d.quitDoWorkerDone <- true
			d.quitDoNotifyListeners <- true

			d.wg.Wait()
			d.setStatus(DispatcherStatusWait)
			return nil
		}
	}
}

// Обработчик сигналов завершения работы от воркеров. Перезапускает задачи, требующего повторения.
func (d *Dispatcher) doWorkerDone() {
	defer d.wg.Done()

	for {
		select {
		case w := <-d.doneWorker:
			t := w.GetTask()
			d.GetTasks().Remove(t)

			w.Reset()
			d.GetWorkers().Update(w)

			go d.addNotifyListeners(t)

			if repeats := t.GetRepeats(); repeats == -1 || t.GetAttempts() < repeats {
				t.SetStatus(task.TaskStatusRepeatWait)
				d.AddTask(t)
			} else {
				go d.notifyAllowExecuteTasks()
			}

		case <-d.quitDoWorkerDone:
			return
		}
	}
}

// Основной цикл диспетчера, после получения сигнала на обработку через канал allowExecuteTasks начинает поиск
// задач на обработку. Сигнал allowExecuteTasks посылается при добавлении нового задания, а так же по таймеру,
// который находится внутри doExecuteTasks, что бы исключить блокировку в обработке задач
func (d *Dispatcher) doExecuteTasks() {
	defer d.wg.Done()

	for {
		select {
		case <-d.quitDoExecuteTasks:
			d.tickerDoExecuteTasks.Stop()
			return

		case <-d.allowExecuteTasks:
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

		case <-d.tickerDoExecuteTasks.C():
			d.notifyAllowExecuteTasks()
		}
	}
}

func (d *Dispatcher) notifyAllowExecuteTasks() {
	if d.GetStatus() == DispatcherStatusProcess && len(d.allowExecuteTasks) == 0 {
		d.allowExecuteTasks <- true
	}
}

// Отдельная очередь на оповещение листенеров, чтобы не блокировать основной процесс.
// Асинхронно принимает сообщения через notifyListeners и последовательно оповещает листенеров.
func (d *Dispatcher) doNotifyListeners() {
	defer d.wg.Done()

	for {
		select {
		case <-d.quitDoNotifyListeners:
			d.tickerDoNotifyListeners.Stop()
			return

		case <-d.allowNotifyListeners:
			listeners := d.listeners.GetAll()
			if len(listeners) > 0 {
				for {
					t := d.listenersTasks.Shift()
					if t == nil {
						break
					}

					for _, l := range listeners {
						done := make(chan struct{}, 1)

						go func() {
							l.NotifyTaskDone(t)
							done <- struct{}{}
						}()

						select {
						case <-done:
						case <-time.After(time.Second):
							l.NotifyTaskDoneTimeout(t)
						}
					}
				}
			}

		case <-d.tickerDoNotifyListeners.C():
			d.notifyAllowNotifyListeners()
		}
	}
}

func (d *Dispatcher) addNotifyListeners(t task.Tasker) {
	d.listenersTasks.Add(t)
	d.notifyAllowNotifyListeners()
}

func (d *Dispatcher) notifyAllowNotifyListeners() {
	if len(d.allowNotifyListeners) == 0 {
		d.allowNotifyListeners <- true
	}
}

func (d *Dispatcher) addWorker(w worker.Worker) {
	d.GetWorkers().Add(w)
	d.notifyAllowExecuteTasks()
}

func (d *Dispatcher) addTask(t task.Tasker) {
	d.GetTasks().Add(t)
	d.notifyAllowExecuteTasks()
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

func (d *Dispatcher) AddTaskByPriorityAndFunc(p int64, f task.TaskFunction, a ...interface{}) task.Tasker {
	task := task.NewTask(f, a...)
	task.SetPriority(p)

	d.AddTask(task)

	return task
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

func (d *Dispatcher) GetListeners() []Listener {
	return d.listeners.GetAll()
}

func (d *Dispatcher) AddListener(l Listener) {
	d.listeners.Add(l)
}

func (d *Dispatcher) RemoveListener(l Listener) {
	d.listeners.Remove(l)
}

func (d *Dispatcher) GetListenersTasks() []task.Tasker {
	return d.listenersTasks.GetAll()
}

func (d *Dispatcher) SetTickerExecuteTasksDuration(t time.Duration) {
	d.tickerDoExecuteTasks.SetDuration(t)
}

func (d *Dispatcher) SetTickerNotifyListenersDuration(t time.Duration) {
	d.tickerDoNotifyListeners.SetDuration(t)
}
