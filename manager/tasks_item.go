package manager

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/kihamo/go-workers"
)

type TasksManagerItem struct {
	workers.ManagerItemBase
	mutex sync.RWMutex

	attempts       int64
	task           workers.Task
	allowStartAt   unsafe.Pointer
	firstStartedAt unsafe.Pointer
	lastStartedAt  unsafe.Pointer

	index  int64
	cancel context.CancelFunc
}

func NewTasksManagerItem(task workers.Task, status workers.TaskStatus) *TasksManagerItem {
	item := &TasksManagerItem{
		task:  task,
		index: -1,
	}

	allowStartAt := time.Now()
	startedAt := task.StartedAt()
	if startedAt != nil && startedAt.After(allowStartAt) {
		allowStartAt = *startedAt
	}

	item.SetAllowStartAt(allowStartAt)
	item.SetStatus(status)

	return item
}

func (t *TasksManagerItem) Task() workers.Task {
	return t.task
}

func (t *TasksManagerItem) Id() string {
	return t.task.Id()
}

func (t *TasksManagerItem) Metadata() workers.Metadata {
	return workers.Metadata{
		workers.TaskMetadataStatus:         t.Status(),
		workers.TaskMetadataAttempts:       t.Attempts(),
		workers.TaskMetadataAllowStartAt:   t.AllowStartAt(),
		workers.TaskMetadataFirstStartedAt: t.FirstStartedAt(),
		workers.TaskMetadataLastStartedAt:  t.LastStartedAt(),
	}
}

func (t *TasksManagerItem) Attempts() int64 {
	return atomic.LoadInt64(&t.attempts)
}

func (t *TasksManagerItem) SetAttempts(attempt int64) {
	atomic.StoreInt64(&t.attempts, attempt)
}

func (t *TasksManagerItem) AllowStartAt() *time.Time {
	p := atomic.LoadPointer(&t.allowStartAt)
	return (*time.Time)(p)
}

func (t *TasksManagerItem) SetAllowStartAt(allowStartedAt time.Time) {
	atomic.StorePointer(&t.allowStartAt, unsafe.Pointer(&allowStartedAt))
}

func (t *TasksManagerItem) IsAllowedStart() bool {
	now := time.Now()
	allowStartAt := t.AllowStartAt()

	return allowStartAt.Before(now) || allowStartAt.Equal(now)
}

func (t *TasksManagerItem) FirstStartedAt() *time.Time {
	p := atomic.LoadPointer(&t.firstStartedAt)
	return (*time.Time)(p)
}

func (t *TasksManagerItem) SetFirstStartedAt(firstStartedAt time.Time) {
	atomic.StorePointer(&t.firstStartedAt, unsafe.Pointer(&firstStartedAt))
}

func (t *TasksManagerItem) LastStartedAt() *time.Time {
	p := atomic.LoadPointer(&t.lastStartedAt)
	return (*time.Time)(p)
}

func (t *TasksManagerItem) SetLastStartedAt(lastStartedAt time.Time) {
	atomic.StorePointer(&t.lastStartedAt, unsafe.Pointer(&lastStartedAt))
}

func (t *TasksManagerItem) IsWait() bool {
	return t.IsStatus(workers.TaskStatusWait) || t.IsStatus(workers.TaskStatusRepeatWait)
}

func (t *TasksManagerItem) IsLocked() bool {
	return t.ManagerItemBase.IsLocked() || !t.IsAllowedStart()
}

func (t *TasksManagerItem) Index() int {
	return int(atomic.LoadInt64(&t.index))
}

func (t *TasksManagerItem) setIndex(index int) {
	atomic.StoreInt64(&t.index, int64(index))
}

func (t *TasksManagerItem) Status() workers.Status {
	return workers.TaskStatus(t.StatusInt64())
}

func (t *TasksManagerItem) SetCancel(cancel context.CancelFunc) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.cancel = cancel
}

func (t *TasksManagerItem) Cancel() {
	t.mutex.RLock()
	cancel := t.cancel
	t.mutex.RUnlock()

	if cancel != nil {
		cancel()
	}
}
