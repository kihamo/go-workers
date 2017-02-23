package dispatcher

import (
	"sync"
)

type listenerList struct {
	mutex sync.RWMutex
	items []Listener
}

func newListenerList() *listenerList {
	return &listenerList{
		items: make([]Listener, 0),
	}
}

func (l *listenerList) getAll() []Listener {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	return l.items[:]
}

func (l *listenerList) add(item Listener) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.items = append(l.items, item)
}

func (l *listenerList) remove(item Listener) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for i := len(l.items) - 1; i >= 0; i-- {
		if l.items[i] == item {
			l.items = append(l.items[:i], l.items[i+1:]...)
		}
	}
}
