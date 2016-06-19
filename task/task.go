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
	GetFunction() TaskFunction
	GetArguments() []interface{}
	GetId() string
	GetName() string
	SetName(string)
	GetDuration() time.Duration
	SetDuration(time.Duration)
	GetRepeats() int64
	SetRepeats(int64)
	GetAttempts() int64
	SetAttempts(int64)
	GetStatus() int64
	SetStatus(int64)
	GetLastError() interface{}
	SetLastError(interface{})
	GetFinishedAt() *time.Time
	SetFinishedAt(time.Time)
	GetTimeout() time.Duration
	SetTimeout(time.Duration)
}

type TaskFunction func(int64, chan bool, ...interface{}) (int64, time.Duration)

type Task struct {
	mutex sync.RWMutex

	fn         TaskFunction
	args       []interface{}
	id         string
	name       string
	duration   time.Duration
	repeats    int64
	attempts   int64
	status     int64
	lastError  interface{}
	finishedAt *time.Time
	timeout    time.Duration
	createdAt  time.Time
}

func NewTask(f TaskFunction, a ...interface{}) *Task {
	return &Task{
		fn:        f,
		args:      a,
		id:        uuid.New(),
		duration:  0,
		repeats:   1,
		attempts:  0,
		status:    TaskStatusWait,
		timeout:   0,
		createdAt: workers.Clock.Now(),
	}
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

func (m *Task) GetId() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.id
}

func (m *Task) GetName() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.name == "" {
		return m.id
	}

	return m.name
}

func (m *Task) SetName(n string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.name = n
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

func (m *Task) SetFinishedAt(t time.Time) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.finishedAt = &t
}

func (m *Task) GetTimeout() time.Duration {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.timeout
}

func (m *Task) SetTimeout(t time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.timeout = t
}
