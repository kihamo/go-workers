package worker

import (
	"errors"
	"sync"
	"time"

	"code.cloudfoundry.org/clock"
	"github.com/google/uuid"
	"github.com/kihamo/go-workers/task"
)

const (
	WorkerStatusWait = int64(iota)
	WorkerStatusProcess
	WorkerStatusBusy
)

type Worker interface {
	Run() error
	Kill() error
	Reset()
	SendTask(task.Tasker)
	SetChangeStatusChannel(chan int64)
	GetTask() task.Tasker
	GetId() string
	GetStatus() int64
	GetCreatedAt() time.Time
	GetClock() clock.Clock
}

type Workman struct {
	mutex sync.RWMutex
	wg    sync.WaitGroup
	clock clock.Clock

	id        string
	status    int64
	createdAt time.Time

	changeStatus chan int64
	kill         chan bool
	done         chan Worker

	task     task.Tasker
	newTask  chan task.Tasker
	killTask chan bool
}

func NewWorkman(d chan Worker) *Workman {
	return NewWorkmanWithClock(d, clock.NewClock())
}

func NewWorkmanWithClock(d chan Worker, c clock.Clock) *Workman {
	return &Workman{
		clock: c,

		id:        uuid.New().String(),
		status:    WorkerStatusWait,
		createdAt: c.Now(),
		kill:      make(chan bool, 1),
		done:      d,

		newTask:  make(chan task.Tasker, 1),
		killTask: make(chan bool, 1),
	}
}

func (m *Workman) Run() error {
	if m.GetStatus() != WorkerStatusWait {
		return errors.New("Worker is running")
	}

	defer func() {
		m.setStatus(WorkerStatusWait)
		m.done <- m
	}()

	m.setStatus(WorkerStatusProcess)

	for {
		select {
		case t := <-m.newTask:
			m.setStatus(WorkerStatusBusy)
			m.setTask(t)

			m.wg.Add(1)
			go m.processTask()

		case <-m.kill:
			if m.GetStatus() == WorkerStatusBusy {
				m.killTask <- true
			}

			m.wg.Wait()
			return nil
		}
	}
}

func (m *Workman) processTask() {
	defer m.wg.Done()

	t := m.GetTask()

	t.SetStartedAt(m.GetClock().Now())
	t.SetReturns(nil)
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
	t := m.GetTask()
	resultChan := make(chan []interface{}, 1)
	errorChan := make(chan interface{}, 1)
	quitChan := make(chan bool, 1)

	m.wg.Add(1)
	go func() {
		defer func() {
			t.SetFinishedAt(m.GetClock().Now())

			if err := recover(); err != nil {
				// TODO: log stack trace
				errorChan <- err
			}

			m.wg.Done()
		}()

		newRepeats, newDuration, returns, err := t.GetFunction()(t.GetAttempts(), quitChan, t.GetArguments()...)
		resultChan <- []interface{}{newRepeats, newDuration, returns, err}
	}()

	for {
		timeout := t.GetTimeout()

		if timeout > 0 {
			timer := m.GetClock().NewTimer(timeout)

			select {
			case r := <-resultChan:
				timer.Stop()

				t.SetRepeats(r[0].(int64))
				t.SetDuration(r[1].(time.Duration))
				t.SetReturns(r[2])

				if r[3] != nil {
					t.SetStatus(task.TaskStatusFail)
					t.SetLastError(r[3])
				} else {
					t.SetStatus(task.TaskStatusSuccess)
				}

				return

			case err := <-errorChan:
				timer.Stop()

				t.SetStatus(task.TaskStatusFail)
				t.SetLastError(err)
				return

			case <-m.killTask:
				timer.Stop()

				quitChan <- true
				t.SetStatus(task.TaskStatusKill)
				return

			case <-timer.C():
				quitChan <- true
				t.SetStatus(task.TaskStatusFailByTimeout)
				return
			}
		} else {
			select {
			case r := <-resultChan:
				t.SetRepeats(r[0].(int64))
				t.SetDuration(r[1].(time.Duration))
				t.SetReturns(r[2])

				if r[3] != nil {
					t.SetStatus(task.TaskStatusFail)
					t.SetLastError(r[3])
				} else {
					t.SetStatus(task.TaskStatusSuccess)
				}

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

func (m *Workman) Kill() error {
	if m.GetStatus() != WorkerStatusWait {
		m.kill <- true
		return nil
	}

	return errors.New("Worker isn't running")
}

func (m *Workman) Reset() {
	if m.GetStatus() == WorkerStatusBusy {
		m.killTask <- true
	}

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

func (m *Workman) SetChangeStatusChannel(c chan int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.changeStatus = c
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

	if m.changeStatus != nil && m.status != s {
		m.changeStatus <- s
	}

	m.status = s
}

func (m *Workman) GetCreatedAt() time.Time {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.createdAt
}

func (m *Workman) GetClock() clock.Clock {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.clock
}
