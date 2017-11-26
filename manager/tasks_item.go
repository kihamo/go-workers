package manager

import (
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/kihamo/go-workers"
)

type TasksManagerItem struct {
	workers.ManagerItemBase

	task         workers.Task
	attempts     int64
	allowStartAt time.Time
	startedAt    unsafe.Pointer
	finishedAt   unsafe.Pointer

	index int64
}

func NewTasksManagerItem(task workers.Task, status workers.TaskStatus) *TasksManagerItem {
	t := &TasksManagerItem{
		task:         task,
		allowStartAt: time.Now(),
	}

	t.SetStatus(status)

	d := task.Duration()
	if d > 0 {
		t.allowStartAt.Add(d)
	}

	return t
}

func (t *TasksManagerItem) Task() workers.Task {
	return t.task
}

func (t *TasksManagerItem) Id() string {
	return t.task.Id()
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
