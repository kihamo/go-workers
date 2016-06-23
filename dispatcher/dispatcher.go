package dispatcher

import (
	"container/heap"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"sync"

	"github.com/kihamo/go-workers"
	"github.com/kihamo/go-workers/collection"
	"github.com/kihamo/go-workers/task"
	"github.com/kihamo/go-workers/worker"
)

const (
	DispatcherStatusWait = int64(iota)
	DispatcherStatusProcess
)

var (
	funcNameRegexp      *regexp.Regexp
	funcNameSubexpNames []string
)

func init() {
	funcNameRegexp = regexp.MustCompile("" +
		// package
		"^(?P<package>[^/]*[^.]*)?" +

		".*?" +

		// name
		"(" +
		"(?:glob\\.)?(?P<name>func)(?:\\d+)" + // anonymous func in go >= 1.5 dispatcher.glob.func1 or method.func1
		"|(?P<name>func)(?:·\\d+)" + // anonymous func in go < 1.5, ex. dispatcher.func·002
		"|(?P<name>[^.]+?)(?:\\)[-·]fm)?" + // dispatcher.jobFunc or dispatcher.jobSleepSixSeconds)·fm
		")?$")
	funcNameSubexpNames = funcNameRegexp.SubexpNames()
}

type Dispatcher struct {
	mutex     sync.RWMutex
	waitGroup *sync.WaitGroup

	workers   *collection.Workers
	tasks     *collection.Tasks
	waitTasks *collection.Tasks

	status      int64
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

		workers:   collection.NewWorkers(),
		tasks:     collection.NewTasks(),
		waitTasks: collection.NewTasks(),

		status:      DispatcherStatusWait,
		workersBusy: 0,

		newQueue:     make(chan task.Tasker),
		executeQueue: make(chan task.Tasker),

		done:            make(chan worker.Worker),
		quit:            make(chan bool, 1),
		allowProcessing: make(chan bool),
	}
}

func (d *Dispatcher) Run() error {
	if d.GetStatus() == DispatcherStatusProcess {
		return errors.New("Dispatcher is running")
	}

	d.setStatus(DispatcherStatusProcess)

	defer func() {
		d.setStatus(DispatcherStatusWait)
	}()

	// отслеживание квоты на занятость исполнителей
	go func() {
		for {
			d.executeQueue <- <-d.newQueue

			<-d.allowProcessing
		}
	}()

	for d.waitTasks.Len() > 0 {
		d.AddTask(d.waitTasks.Pop().(task.Tasker))
	}

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
			return nil
		}
	}
}

func (d *Dispatcher) AddWorker() worker.Worker {
	w := worker.NewWorkman(d.done)
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
	add := func() {
		if d.GetStatus() == DispatcherStatusProcess {
			d.tasks.Push(t)
			d.newQueue <- t
		} else {
			d.waitTasks.Push(t)
		}
	}

	duration := t.GetDuration()
	if duration > 0 {
		timer := workers.Clock.NewTimer(duration)
		go func() {
			<-timer.C()
			add()
		}()
	} else {
		add()
	}
}

func (d *Dispatcher) AddNamedTaskByFunc(n string, f task.TaskFunction, a ...interface{}) task.Tasker {
	task := task.NewTask(f, a)
	task.SetName(n)

	d.AddTask(task)

	return task
}

func (d *Dispatcher) AddTaskByFunc(f task.TaskFunction, a ...interface{}) task.Tasker {
	name := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()

	parts := funcNameRegexp.FindAllStringSubmatch(name, -1)
	if len(parts) > 0 {
		fmt.Println(parts)

		for i, value := range parts[0] {
			switch funcNameSubexpNames[i] {
			case "name":
				if value != "" {
					name += "." + value
				}
			case "package":
				name = value
			}
		}
	}

	return d.AddNamedTaskByFunc(name, f, a...)
}

func (d *Dispatcher) GetTasks() *collection.Tasks {
	return d.tasks
}

func (d *Dispatcher) GetWaitTasks() *collection.Tasks {
	return d.waitTasks
}

func (d *Dispatcher) Kill() error {
	if d.GetStatus() == DispatcherStatusProcess {
		d.quit <- true
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
