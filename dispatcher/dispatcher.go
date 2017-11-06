package dispatcher

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"
	"context"

	"code.cloudfoundry.org/clock"
	"github.com/kihamo/go-workers"
	"github.com/kihamo/go-workers/task"
	"github.com/kihamo/go-workers/worker"
)

const (
	DispatcherStatusWait = int64(iota)
	DispatcherStatusProcess
	DispatcherStatusCancel
)

type Dispatcher struct {
	mutex sync.RWMutex
	wg    sync.WaitGroup
	clock clock.Clock

	ctx context.Context
	cancel context.CancelFunc

	workers *Workers
	tasks   *Tasks

	status         int64
	listeners      *ListenerList
	listenersTasks *ListenerTasks

	doneWorker       chan worker.Worker // канал уведомления о завершении рабочего

	allowExecuteTasks    chan struct{} // канал для блокировки выполнения новых задач для случая, когда все исполнители заняты
	tickerDoExecuteTasks *workers.Ticker

	allowNotifyListeners    chan struct{}
	tickerDoNotifyListeners *workers.Ticker
}

func NewDispatcher() *Dispatcher {
	return NewDispatcherWithClock(clock.NewClock())
}

func NewDispatcherWithClock(c clock.Clock) *Dispatcher {
	d := &Dispatcher{
		clock: c,

		workers: NewWorkers(),
		tasks:   NewTasks(),

		status:         DispatcherStatusWait,
		listeners:      NewListenerList(),
		listenersTasks: NewListenerTasks(),

		doneWorker:       make(chan worker.Worker),

		allowExecuteTasks:    make(chan struct{}, 1),
		tickerDoExecuteTasks: workers.NewTicker(time.Second),

		allowNotifyListeners:    make(chan struct{}, 1),
		tickerDoNotifyListeners: workers.NewTicker(time.Second),
	}

	d.ctx, d.cancel = context.WithCancel(context.Background())

	return d
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

	<-d.ctx.Done()

	d.setStatus(DispatcherStatusCancel)

	d.wg.Wait()
	d.setStatus(DispatcherStatusWait)

	return nil
}

// Обработчик сигналов завершения работы от воркеров. Перезапускает задачи, требующего повторения.
func (d *Dispatcher) doWorkerDone() {
	defer d.wg.Done()

	for {
		select {
		case w := <-d.doneWorker:
			t := w.GetTask()
			d.getSafeTasks().Remove(t)

			w.Reset()
			d.getSafeWorkers().Update(w)

			go d.addNotifyListeners(t)

			if repeats := t.GetRepeats(); repeats == -1 || t.GetAttempts() < repeats {
				t.SetStatus(task.TaskStatusRepeatWait)
				d.AddTask(t)
			} else {
				go d.notifyAllowExecuteTasks()
			}

		case <-d.ctx.Done():
			for _, w := range d.getSafeWorkers().GetItems() {
				w.Kill()
			}

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
		case <-d.ctx.Done():
			d.tickerDoExecuteTasks.Stop()
			return

		case <-d.allowExecuteTasks:
			workersList := d.getSafeWorkers()
			tasksList := d.getSafeTasks()

			for tasksList.HasWait() && workersList.HasWait() {
				t := tasksList.GetWait()
				w := workersList.GetWait()

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
		d.allowExecuteTasks <- struct{}{}
	}
}

// Отдельная очередь на оповещение листенеров, чтобы не блокировать основной процесс.
// Асинхронно принимает сообщения через notifyListeners и последовательно оповещает листенеров.
func (d *Dispatcher) doNotifyListeners() {
	defer d.wg.Done()

	for {
		select {
		case <-d.ctx.Done():
			d.tickerDoNotifyListeners.Stop()
			return

		case <-d.allowNotifyListeners:
			listeners := d.GetListeners()
			if len(listeners) > 0 {
				list := d.getSafeListenersTasks()

				for {
					t := list.Shift()
					if t == nil {
						break
					}

					for _, l := range listeners {
						done := make(chan struct{}, 1)

						go func() {
							if err := l.NotifyTaskDone(t); err != nil {
								log.Printf("Notify listener %s about task done returns error %s", l.GetName(), err.Error())
							}

							done <- struct{}{}
						}()

						select {
						case <-done:
						case <-time.After(time.Second):
							if err := l.NotifyTaskDoneTimeout(t); err != nil {
								log.Printf("Notify listener with timout %s about task done returns error %s", l.GetName(), err.Error())
							}
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
		d.allowNotifyListeners <- struct{}{}
	}
}

func (d *Dispatcher) addWorker(w worker.Worker) {
	d.getSafeWorkers().Add(w)
	d.notifyAllowExecuteTasks()
}

func (d *Dispatcher) AddWorker() worker.Worker {
	w := worker.NewWorkmanWithClock(d.doneWorker, d.GetClock())
	d.addWorker(w)

	return w
}

func (d *Dispatcher) getSafeWorkers() *Workers {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return d.workers
}

func (d *Dispatcher) GetWorkers() []worker.Worker {
	return d.getSafeWorkers().GetItems()
}

func (d *Dispatcher) RemoveWorker(w worker.Worker) {
	d.getSafeWorkers().Remove(w)
}

func (d *Dispatcher) addTask(t task.Tasker) {
	d.getSafeTasks().Add(t)
	d.notifyAllowExecuteTasks()
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

func (d *Dispatcher) getSafeTasks() *Tasks {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return d.tasks
}

func (d *Dispatcher) GetTasks() []task.Tasker {
	return d.getSafeTasks().GetItems()
}

func (d *Dispatcher) RemoveTask(t task.Tasker) {
	d.getSafeTasks().Remove(t)
}

func (d *Dispatcher) RemoveTaskById(id string) {
	d.getSafeTasks().RemoveById(id)
}

func (d *Dispatcher) Reset() {
	d.mutex.Lock()

	ws := d.workers.GetItems()

	go func() {
		for _, w := range ws {
			w.Kill()
		}
	}()

	d.workers = NewWorkers()
	d.tasks = NewTasks()
	d.listeners = NewListenerList()
	d.listenersTasks = NewListenerTasks()

	d.mutex.Unlock()

	for i := 0; i < len(ws); i++ {
		d.AddWorker()
	}
}

func (d *Dispatcher) Kill() error {
	d.cancel()
	return d.ctx.Err()
}

func (d *Dispatcher) GetStatus() int64 {
	return atomic.LoadInt64(&d.status)
}

func (d *Dispatcher) setStatus(s int64) {
	atomic.StoreInt64(&d.status, s)
}

func (d *Dispatcher) GetClock() clock.Clock {
	return d.clock
}

func (d *Dispatcher) getSafeListeners() *ListenerList {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return d.listeners
}

func (d *Dispatcher) GetListeners() []Listener {
	return d.getSafeListeners().GetAll()
}

func (d *Dispatcher) AddListener(l Listener) {
	d.getSafeListeners().Add(l)
}

func (d *Dispatcher) RemoveListener(l Listener) {
	d.getSafeListeners().Remove(l)
}

func (d *Dispatcher) getSafeListenersTasks() *ListenerTasks {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return d.listenersTasks
}

func (d *Dispatcher) GetListenersTasks() []task.Tasker {
	return d.getSafeListenersTasks().GetAll()
}

func (d *Dispatcher) SetTickerExecuteTasksDuration(t time.Duration) {
	d.tickerDoExecuteTasks.SetDuration(t)
}

func (d *Dispatcher) SetTickerNotifyListenersDuration(t time.Duration) {
	d.tickerDoNotifyListeners.SetDuration(t)
}
