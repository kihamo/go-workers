package task

import (
	"reflect"
	"regexp"
	"runtime"
	"sync"
	"time"

	"github.com/pborman/uuid"
	"github.com/pivotal-golang/clock"
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

type Tasker interface {
	GetFunction() TaskFunction
	GetFunctionName() string
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
	GetPriority() int64
	SetPriority(int64)
	GetLastError() interface{}
	SetLastError(interface{})
	GetTimeout() time.Duration
	SetTimeout(time.Duration)
	GetCreatedAt() time.Time
	GetStartedAt() *time.Time
	SetStartedAt(time.Time)
	GetFinishedAt() *time.Time
	SetFinishedAt(time.Time)
}

type TaskFunction func(int64, chan bool, ...interface{}) (int64, time.Duration, error)

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
	priority   int64
	lastError  interface{}
	timeout    time.Duration
	createdAt  time.Time
	startedAt  *time.Time
	finishedAt *time.Time
}

func NewTask(f TaskFunction, a ...interface{}) *Task {
	return NewTaskWithClock(clock.NewClock(), f, a...)
}

func NewTaskWithClock(c clock.Clock, f TaskFunction, a ...interface{}) *Task {
	return &Task{
		fn:        f,
		args:      a,
		id:        uuid.New(),
		duration:  0,
		repeats:   1,
		attempts:  0,
		status:    TaskStatusWait,
		priority:  1,
		timeout:   0,
		createdAt: c.Now(),
	}
}

func (m *Task) GetFunction() TaskFunction {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.fn
}

func (m *Task) GetFunctionName() string {
	name := runtime.FuncForPC(reflect.ValueOf(m.GetFunction()).Pointer()).Name()

	parts := funcNameRegexp.FindAllStringSubmatch(name, -1)
	if len(parts) > 0 {
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

	return name
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
		return m.GetFunctionName()
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

func (m *Task) GetPriority() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.priority
}

func (m *Task) SetPriority(p int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.priority = p
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

func (m *Task) GetCreatedAt() time.Time {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.createdAt
}

func (m *Task) GetStartedAt() *time.Time {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// TODO: copy value

	return m.startedAt
}

func (m *Task) SetStartedAt(t time.Time) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.startedAt = &t
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
