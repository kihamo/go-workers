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

	attempts     int64
	task         workers.Task
	allowStartAt time.Time
	startedAt    unsafe.Pointer
	finishedAt   unsafe.Pointer

	index  int64
	cancel context.CancelFunc
}

func NewTasksManagerItem(task workers.Task, status workers.TaskStatus) *TasksManagerItem {
	item := &TasksManagerItem{
		task:         task,
		allowStartAt: time.Now(),
		index:        -1,
	}

	item.SetStatus(status)

	d := task.Duration()
	if d > 0 {
		item.allowStartAt.Add(d)
	}

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
		workers.TaskMetadataStatus:       t.Status(),
		workers.TaskMetadataAttempts:     t.Attempts(),
		workers.TaskMetadataAllowStartAt: t.AllowStartAt(),
		workers.TaskMetadataStartedAt:    t.StartedAt(),
		workers.TaskMetadataFinishedAt:   t.FinishedAt(),
	}
}

func (t *TasksManagerItem) Attempts() int64 {
	return atomic.LoadInt64(&t.attempts)
}

func (t *TasksManagerItem) SetAttempts(attempt int64) {
	atomic.StoreInt64(&t.attempts, attempt)
}

func (t *TasksManagerItem) AllowStartAt() time.Time {
	return t.allowStartAt
}

func (t *TasksManagerItem) IsAllowedStart() bool {
	now := time.Now()
	return t.allowStartAt.Before(now) || t.allowStartAt.Equal(now)
}

func (t *TasksManagerItem) StartedAt() *time.Time {
	p := atomic.LoadPointer(&t.startedAt)
	return (*time.Time)(p)
}

func (t *TasksManagerItem) SetStartedAt(startedAt time.Time) {
	atomic.StorePointer(&t.startedAt, unsafe.Pointer(&startedAt))
}

func (t *TasksManagerItem) FinishedAt() *time.Time {
	p := atomic.LoadPointer(&t.finishedAt)
	return (*time.Time)(p)
}

func (t *TasksManagerItem) SetFinishedAt(finishedAt time.Time) {
	atomic.StorePointer(&t.finishedAt, unsafe.Pointer(&finishedAt))
}

func (t *TasksManagerItem) IsWait() bool {
	return t.IsStatus(workers.TaskStatusWait) || t.IsStatus(workers.TaskStatusRepeatWait)
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
