package task

import (
	"sync"
	"time"

	"github.com/kihamo/go-workers"
	"github.com/pborman/uuid"
)

const (
	TaskStatusWait = int64(iota)
	TaskStatusProcess
	TaskStatusSuccess
	TaskStatusFail
	TaskStatusFailByTimeout
	TaskStatusKill
	TaskStatusRepeatWait
)

type Tasker interface {
	GetId() string
	GetName() string
	GetDuration() time.Duration
	SetDuration(time.Duration)
	GetRepeats() int64
	SetRepeats(int64)
	GetAttempts() int64
	SetAttempts(int64)
	GetFunction() TaskFunction
	GetArguments() []interface{}
	GetStatus() int64
	SetStatus(int64)
	GetLastError() interface{}
	SetLastError(interface{})
	GetFinishedAt() *time.Time
	SetFinishedTime(time.Time)
	GetTimeout() time.Duration
}

type TaskFunction func(int64, chan bool, ...interface{}) (int64, time.Duration)

type Task struct {
	mutex sync.RWMutex

	name     string
	duration time.Duration
	repeats  int64
	fn       TaskFunction
	args     []interface{}

	id         string
	status     int64
	attempts   int64
	createdAt  time.Time
	finishedAt *time.Time
	timeout    time.Duration
	lastError  interface{}
}

func NewTask(n string, d time.Duration, t time.Duration, r int64, f TaskFunction, a ...interface{}) *Task {
	return &Task{
		name:     n,
		duration: d,
		timeout:  t,
		repeats:  r,
		fn:       f,
		args:     a,

		id:        uuid.New(),
		status:    TaskStatusWait,
		attempts:  0,
		createdAt: workers.Clock.Now(),
	}
}

func (m *Task) GetId() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.id
}

func (m *Task) GetName() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.name
}

func (m *Task) GetDuration() time.Duration {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.duration
}

func (m *Task) SetDuration(d time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.duration = d
}

func (m *Task) GetRepeats() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.repeats
}

func (m *Task) SetRepeats(r int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.repeats = r
}

func (m *Task) GetAttempts() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.attempts
}

func (m *Task) SetAttempts(a int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.attempts = a
}

func (m *Task) GetFunction() TaskFunction {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.fn
}

func (m *Task) GetArguments() []interface{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.args
}

func (m *Task) GetStatus() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.status
}

func (m *Task) SetStatus(s int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.status = s
}

func (m *Task) GetLastError() interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.lastError
}

func (m *Task) SetLastError(e interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.lastError = e
}

func (m *Task) GetFinishedAt() *time.Time {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// TODO: copy value

	return m.finishedAt
}

func (m *Task) SetFinishedTime(t time.Time) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.finishedAt = &t
}

func (m *Task) GetTimeout() time.Duration {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.timeout
}
