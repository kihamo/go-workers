package dispatcher

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/kihamo/go-workers/task"
)

type ListenerItem interface {
	Listener

	GetCreatedAt() time.Time
	GetLastTaskSuccessAt() *time.Time
	GetLastTaskFailedAt() *time.Time
	GetCountTaskSuccess() uint64
	GetCountTaskFailed() uint64
}

type listenerItem struct {
	mutex    sync.RWMutex
	listener Listener

	createdAt         time.Time
	lastTaskSuccessAt *time.Time
	lastTaskFailedAt  *time.Time

	countTaskSuccess uint64
	countTaskFailed  uint64
}

func NewListenerItem(l Listener) *listenerItem {
	return &listenerItem{
		listener:  l,
		createdAt: time.Now(),
	}
}

func (l *listenerItem) GetCreatedAt() time.Time {
	return l.createdAt
}

func (l *listenerItem) GetLastTaskSuccessAt() *time.Time {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	return l.lastTaskSuccessAt
}

func (l *listenerItem) GetLastTaskFailedAt() *time.Time {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	return l.lastTaskFailedAt
}

func (l *listenerItem) GetCountTaskSuccess() uint64 {
	return atomic.LoadUint64(&l.countTaskSuccess)
}

func (l *listenerItem) GetCountTaskFailed() uint64 {
	return atomic.LoadUint64(&l.countTaskFailed)
}

func (l *listenerItem) successTask() {
	now := time.Now()

	l.mutex.Lock()
	l.lastTaskSuccessAt = &now
	l.mutex.Unlock()

	atomic.AddUint64(&l.countTaskSuccess, 1)
}

func (l *listenerItem) failedTask() {
	now := time.Now()

	l.mutex.Lock()
	l.lastTaskFailedAt = &now
	l.mutex.Unlock()

	atomic.AddUint64(&l.countTaskFailed, 1)
}

func (l *listenerItem) GetName() string {
	return l.listener.GetName()
}

func (l *listenerItem) NotifyTaskDone(t task.Tasker) error {
	err := l.listener.NotifyTaskDone(t)

	if err == nil {
		l.successTask()
	} else {
		l.failedTask()
	}

	return err
}

func (l *listenerItem) NotifyTaskDoneTimeout(t task.Tasker) error {
	err := l.listener.NotifyTaskDoneTimeout(t)

	if err == nil {
		l.successTask()
	} else {
		l.failedTask()
	}

	return err
}
