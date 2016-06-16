package workers

import (
	"sync"
	"time"

	"github.com/pborman/uuid"
)

const (
	taskStatusWait = int64(iota)
	taskStatusProcess
	taskStatusSuccess
	taskStatusFail
	taskStatusFailByTimeOut
	taskStatusRepeatWait
)

type TaskFunction func(int64, ...interface{}) (int64, time.Duration)

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

func NewTask(name string, duration time.Duration, repeats int64, fn TaskFunction, args ...interface{}) *Task {
	return &Task{
		name:     name,
		duration: duration,
		repeats:  repeats,
		fn:       fn,
		args:     args,

		id:        uuid.New(),
		status:    taskStatusWait,
		attempts:  0,
		createdAt: time.Now(),
	}
}

func (t *Task) GetId() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.id
}

func (t *Task) GetName() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.name
}

func (t *Task) GetDuration() time.Duration {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.duration
}

func (t *Task) SetDuration(duration time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.duration = duration
}

func (t *Task) GetRepeats() int64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.repeats
}

func (t *Task) SetRepeats(repeats int64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.repeats = repeats
}

func (t *Task) GetAttempts() int64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.attempts
}

func (t *Task) setAttempts(attempts int64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.attempts = attempts
}

func (t *Task) GetFunction() TaskFunction {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.fn
}

func (t *Task) GetArguments() []interface{} {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.args
}

func (t *Task) GetStatus() int64 {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.status
}

func (t *Task) setStatus(status int64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.status = status
}

func (t *Task) GetLastError() interface{} {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.lastError
}

func (t *Task) setLastError(err interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.lastError = err
}

func (t *Task) GetFinishedAt() *time.Time {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	// TODO: copy value

	return t.finishedAt
}

func (t *Task) setFinishedTime(time time.Time) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.finishedAt = &time
}
