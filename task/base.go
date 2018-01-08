package task

import (
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/google/uuid"
)

type BaseTask struct {
	priority       int64
	repeats        int64
	repeatInterval int64
	timeout        int64
	id             string
	name           atomic.Value
	createdAt      time.Time
	startedAt      unsafe.Pointer
}

func (t *BaseTask) Init() {
	t.id = uuid.New().String()
	t.createdAt = time.Now()
}

func (t *BaseTask) Id() string {
	return t.id
}

func (t *BaseTask) Name() string {
	var name string

	if value := t.name.Load(); value != nil {
		name = value.(string)
	}

	return name
}

func (t *BaseTask) SetName(name string) {
	t.name.Store(name)
}

func (t *BaseTask) Priority() int64 {
	return atomic.LoadInt64(&t.priority)
}

func (t *BaseTask) SetPriority(priority int64) {
	atomic.StoreInt64(&t.priority, priority)
}

func (t *BaseTask) Repeats() int64 {
	return atomic.LoadInt64(&t.repeats)
}

func (t *BaseTask) SetRepeats(repeats int64) {
	atomic.StoreInt64(&t.repeats, repeats)
}

func (t *BaseTask) RepeatInterval() time.Duration {
	return time.Duration(atomic.LoadInt64(&t.repeatInterval))
}

func (t *BaseTask) SetRepeatInterval(interval time.Duration) {
	atomic.StoreInt64(&t.repeatInterval, int64(interval))
}

func (t *BaseTask) Timeout() time.Duration {
	return time.Duration(atomic.LoadInt64(&t.timeout))
}

func (t *BaseTask) SetTimeout(duration time.Duration) {
	atomic.StoreInt64(&t.timeout, int64(duration))
}

func (t *BaseTask) CreatedAt() time.Time {
	return t.createdAt
}

func (t *BaseTask) StartedAt() *time.Time {
	p := atomic.LoadPointer(&t.startedAt)
	return (*time.Time)(p)
}

func (t *BaseTask) SetStartedAt(startedAt time.Time) {
	atomic.StorePointer(&t.startedAt, unsafe.Pointer(&startedAt))
}

func (t *BaseTask) String() string {
	return "Task #" + t.Id()
}

func (t *BaseTask) GoString() string {
	return fmt.Sprintf("%s %#p", t.String(), t)
}
