package dispatcher

import (
	"sync"
)

type ListenerList struct {
	mutex sync.RWMutex
	items []Listener
}

func NewListenerList() *ListenerList {
	return &ListenerList{
		items: make([]Listener, 0),
	}
}

func (l *ListenerList) GetAll() []Listener {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	return l.items[:]
}

func (l *ListenerList) Add(item Listener) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.items = append(l.items, item)
}

func (l *ListenerList) Remove(item Listener) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for i := len(l.items) - 1; i >= 0; i-- {
		if l.items[i] == item {
			l.items = append(l.items[:i], l.items[i+1:]...)
		}
	}
}
