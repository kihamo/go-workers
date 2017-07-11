package dispatcher

import (
	"sync"
	"time"

	"github.com/kihamo/go-workers/task"
)

type ListenerItem interface {
	Listener

	GetCreated() time.Time
	GetUpdated() time.Time
}

type listenerItem struct {
	mutex    sync.RWMutex
	listener Listener
	created  time.Time
	updated  time.Time
}

func NewListenerItem(l Listener) *listenerItem {
	return &listenerItem{
		listener: l,
		created:  time.Now(),
		updated:  time.Now(),
	}
}

func (l *listenerItem) update() {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.updated = time.Now()
}

func (l *listenerItem) GetCreated() time.Time {
	return l.created
}

func (l *listenerItem) GetUpdated() time.Time {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	return l.updated
}

func (l *listenerItem) GetName() string {
	return l.listener.GetName()
}

func (l *listenerItem) NotifyTaskDone(t task.Tasker) {
	l.update()

	l.listener.NotifyTaskDone(t)
}

func (l *listenerItem) NotifyTaskDoneTimeout(t task.Tasker) {
	l.update()

	l.listener.NotifyTaskDoneTimeout(t)
}
