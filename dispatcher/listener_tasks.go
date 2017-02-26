package dispatcher

import (
	"sync"

	"github.com/kihamo/go-workers/task"
)

type ListenerTasks struct {
	mutex sync.RWMutex
	items []task.Tasker
}

func NewListenerTasks() *ListenerTasks {
	return &ListenerTasks{
		items: make([]task.Tasker, 0),
	}
}

func (l *ListenerTasks) GetAll() []task.Tasker {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	return l.items[:]
}

func (l *ListenerTasks) Add(item task.Tasker) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.items = append(l.items, item)
}

func (l *ListenerTasks) Shift() (item task.Tasker) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if len(l.items) == 0 {
		return nil
	}

	item, l.items = l.items[0], l.items[1:]
	return item
}
