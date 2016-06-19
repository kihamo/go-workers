package worker

import (
	"sync"
	"time"

	"github.com/kihamo/go-workers"
	"github.com/kihamo/go-workers/task"
	"github.com/pborman/uuid"
)

const (
	WorkerStatusWait = int64(iota)
	WorkerStatusProcess
	WorkerStatusBusy
)

type Worker interface {
	Run()
	Kill()
	Reset()
	SendTask(task.Tasker)
	GetTask() task.Tasker
	GetId() string
	GetStatus() int64
	GetCreatedAt() time.Time
}

type Workman struct {
	mutex sync.RWMutex
	wg    sync.WaitGroup

	id        string
	status    int64
	createdAt time.Time
	kill      chan bool
	done      chan Worker

	task     task.Tasker
	newTask  chan task.Tasker
	killTask chan bool
}

func NewWorker(d chan Worker) *Workman {
	return &Workman{
		id:        uuid.New(),
		status:    WorkerStatusWait,
		createdAt: workers.Clock.Now(),
		kill:      make(chan bool, 1),
		done:      d,

		newTask:  make(chan task.Tasker, 1),
		killTask: make(chan bool, 1),
	}
}

func (m *Workman) Run() {
	defer func() {
		m.setStatus(WorkerStatusWait)
		m.done <- m
	}()

	m.setStatus(WorkerStatusProcess)

	for {
		select {
		case task := <-m.newTask:
			m.wg.Add(1)
			m.setStatus(WorkerStatusBusy)

			m.setTask(task)
			go m.processTask()

		case <-m.kill:
			if m.GetStatus() == WorkerStatusBusy {
				m.killTask <- true
			}

			m.wg.Wait()
			return
		}
	}
}

func (m *Workman) processTask() {
	t := m.GetTask()

	t.SetLastError(nil)
	if t.GetStatus() != task.TaskStatusRepeatWait {
		t.SetAttempts(0)
	}

	t.SetStatus(task.TaskStatusProcess)
	t.SetAttempts(t.GetAttempts() + 1)

	m.executeTask()

	m.setStatus(WorkerStatusWait)
	m.kill <- true
}

func (m *Workman) executeTask() {
	defer func() {
		m.wg.Done()
	}()

	t := m.GetTask()
	resultChan := make(chan []interface{}, 1)
	errorChan := make(chan interface{}, 1)
	quitChan := make(chan bool, 1)

	go func() {
		defer func() {
			t.SetFinishedTime(workers.Clock.Now())

			if err := recover(); err != nil {
				errorChan <- err
			}
		}()

		newRepeats, newDuration := t.GetFunction()(t.GetAttempts(), quitChan, t.GetArguments()...)
		resultChan <- []interface{}{newRepeats, newDuration}
	}()

	for {
		timeout := t.GetTimeout()

		if timeout > 0 {
			// execute witch timeout
			select {
			case r := <-resultChan:
				t.SetStatus(task.TaskStatusSuccess)
				t.SetRepeats(r[0].(int64))
				t.SetDuration(r[1].(time.Duration))
				return

			case err := <-errorChan:
				t.SetStatus(task.TaskStatusFail)
				t.SetLastError(err)
				return

			case <-m.killTask:
				quitChan <- true
				t.SetStatus(task.TaskStatusKill)
				return

			case <-time.After(timeout):
				quitChan <- true
				t.SetStatus(task.TaskStatusFailByTimeout)
				return
			}
		} else {
			// execute without timeout
			select {
			case r := <-resultChan:
				t.SetStatus(task.TaskStatusSuccess)
				t.SetRepeats(r[0].(int64))
				t.SetDuration(r[1].(time.Duration))
				return

			case err := <-errorChan:
				t.SetStatus(task.TaskStatusFail)
				t.SetLastError(err)
				return

			case <-m.killTask:
				quitChan <- true
				t.SetStatus(task.TaskStatusKill)
				return
			}
		}
	}
}

func (m *Workman) Kill() {
	m.kill <- true
}

func (m *Workman) Reset() {
	m.setTask(nil)
}

func (m *Workman) GetTask() task.Tasker {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.task
}

func (m *Workman) setTask(t task.Tasker) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.task = t
}

func (m *Workman) SendTask(t task.Tasker) {
	m.newTask <- t
}

func (m *Workman) GetId() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.id
}

func (m *Workman) GetStatus() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.status
}

func (m *Workman) setStatus(s int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.status = s
}

func (m *Workman) GetCreatedAt() time.Time {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.createdAt
}
