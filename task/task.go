package task

import (
	"reflect"
	"regexp"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"

	"code.cloudfoundry.org/clock"
	"github.com/google/uuid"
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
	GetReturns() interface{}
	SetReturns(interface{})
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

type TaskFunction func(int64, chan struct{}, ...interface{}) (int64, time.Duration, interface{}, error)

type Task struct {
	fn         TaskFunction
	args       []interface{}
	id         string
	name       atomic.Value
	duration   int64
	repeats    int64
	attempts   int64
	status     int64
	priority   int64
	returns    atomic.Value
	lastError  atomic.Value
	timeout    int64
	createdAt  time.Time
	startedAt  unsafe.Pointer
	finishedAt unsafe.Pointer
}

func NewTask(f TaskFunction, a ...interface{}) *Task {
	return NewTaskWithClock(clock.NewClock(), f, a...)
}

func NewTaskWithClock(c clock.Clock, f TaskFunction, a ...interface{}) *Task {
	t := &Task{
		fn:        f,
		args:      a,
		id:        uuid.New().String(),
		duration:  0,
		repeats:   1,
		attempts:  0,
		status:    TaskStatusWait,
		priority:  1,
		timeout:   0,
		createdAt: c.Now().UTC(),
	}

	return t
}

func (m *Task) GetFunction() TaskFunction {
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
	return m.args
}

func (m *Task) GetId() string {
	return m.id
}

func (m *Task) GetName() string {
	var n string

	if d := m.name.Load(); d != nil {
		n = d.(string)
	}

	if n == "" {
		return m.GetFunctionName()
	}

	return n
}

func (m *Task) SetName(n string) {
	m.name.Store(n)
}

func (m *Task) GetDuration() time.Duration {
	return time.Duration(atomic.LoadInt64(&m.duration))
}

func (m *Task) SetDuration(d time.Duration) {
	atomic.StoreInt64(&m.duration, int64(d))
}

func (m *Task) GetRepeats() int64 {
	return atomic.LoadInt64(&m.repeats)
}

func (m *Task) SetRepeats(r int64) {
	atomic.StoreInt64(&m.repeats, r)
}

func (m *Task) GetAttempts() int64 {
	return atomic.LoadInt64(&m.attempts)
}

func (m *Task) SetAttempts(a int64) {
	atomic.StoreInt64(&m.attempts, a)
}

func (m *Task) GetStatus() int64 {
	return atomic.LoadInt64(&m.status)
}

func (m *Task) SetStatus(s int64) {
	atomic.StoreInt64(&m.status, s)
}

func (m *Task) GetPriority() int64 {
	return atomic.LoadInt64(&m.priority)
}

func (m *Task) SetPriority(p int64) {
	atomic.StoreInt64(&m.priority, p)
}

func (m *Task) GetReturns() interface{} {
	return m.returns.Load()
}

func (m *Task) SetReturns(r interface{}) {
	if r == nil {
		m.returns = atomic.Value{}
	} else {
		m.returns.Store(r)
	}
}

func (m *Task) GetLastError() interface{} {
	return m.lastError.Load()
}

func (m *Task) SetLastError(e interface{}) {
	if e == nil {
		m.lastError = atomic.Value{}
	} else {
		m.lastError.Store(e)
	}
}

func (m *Task) GetCreatedAt() time.Time {
	return m.createdAt
}

func (m *Task) GetStartedAt() *time.Time {
	p := atomic.LoadPointer(&m.startedAt)
	return (*time.Time)(p)
}

func (m *Task) SetStartedAt(t time.Time) {
	atomic.StorePointer(&m.startedAt, unsafe.Pointer(&t))
}

func (m *Task) GetFinishedAt() *time.Time {
	p := atomic.LoadPointer(&m.finishedAt)
	return (*time.Time)(p)
}

func (m *Task) SetFinishedAt(t time.Time) {
	atomic.StorePointer(&m.finishedAt, unsafe.Pointer(&t))
}

func (m *Task) GetTimeout() time.Duration {
	return time.Duration(atomic.LoadInt64(&m.timeout))
}

func (m *Task) SetTimeout(t time.Duration) {
	atomic.StoreInt64(&m.timeout, int64(t))
}
