package dispatcher

import (
	"sync"

	"github.com/kihamo/go-workers/task"
)

type listenerTasks struct {
	mutex sync.Mutex
	items []task.Tasker
}

func newListenerTasks() *listenerTasks {
	return &listenerTasks{
		items: make([]task.Tasker, 0),
	}
}

func (l *listenerTasks) add(item task.Tasker) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.items = append(l.items, item)
}

func (l *listenerTasks) shift() (item task.Tasker) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if len(l.items) == 0 {
		return nil
	}

	item, l.items = l.items[0], l.items[1:]
	return item
}
