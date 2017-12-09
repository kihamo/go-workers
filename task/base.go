package task

import (
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type BaseTask struct {
	id        string
	name      atomic.Value
	priority  int64
	repeats   int64
	duration  int64
	timeout   int64
	createdAt time.Time
}

func (t *BaseTask) Init() {
	t.id = uuid.New().String()
	t.createdAt = time.Now()
}

func (t *BaseTask) Cancel() error {
	// FIXME:

	return nil
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

func (t *BaseTask) Duration() time.Duration {
	return time.Duration(atomic.LoadInt64(&t.duration))
}

func (t *BaseTask) SetDuration(duration time.Duration) {
	atomic.StoreInt64(&t.duration, int64(duration))
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

func (t *BaseTask) String() string {
	return "Task #" + t.Id()
}
