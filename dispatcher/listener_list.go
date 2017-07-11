package dispatcher

import (
	"sync"
)

type ListenerList struct {
	mutex sync.RWMutex
	items []*listenerItem
}

func NewListenerList() *ListenerList {
	return &ListenerList{
		items: make([]*listenerItem, 0),
	}
}

func (l *ListenerList) GetAll() []Listener {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	items := make([]Listener, len(l.items), len(l.items))

	for i, l := range l.items[:] {
		items[i] = l
	}

	return items
}

func (l *ListenerList) Add(listener Listener) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.items = append(l.items, NewListenerItem(listener))
}

func (l *ListenerList) Remove(listener Listener) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for i := len(l.items) - 1; i >= 0; i-- {
		if l.items[i].listener == listener {
			l.items = append(l.items[:i], l.items[i+1:]...)
		}
	}
}
