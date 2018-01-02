package manager

import (
	"context"
	"sync"
	"time"

	"github.com/kihamo/go-workers"
)

type ListenersManager struct {
	mutex     sync.RWMutex
	events    map[workers.EventId][]*ListenersManagerItem
	listeners map[string]*ListenersManagerItem
}

func NewListenersManager() *ListenersManager {
	return &ListenersManager{
		events:    map[workers.EventId][]*ListenersManagerItem{},
		listeners: map[string]*ListenersManagerItem{},
	}
}

func (m *ListenersManager) Listeners() []workers.Listener {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	listeners := make([]workers.Listener, 0, len(m.listeners))
	for _, item := range m.listeners {
		listeners = append(listeners, item.Listener())
	}

	return listeners
}

func (m *ListenersManager) GetById(id string) *ListenersManagerItem {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.listeners[id]
}

func (m *ListenersManager) Attach(eventId workers.EventId, listener workers.Listener) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	item, ok := m.listeners[listener.Id()]
	if !ok {
		item = NewListenersManagerItem(eventId, listener)
	}
	item.AddEventId(eventId)

	if _, ok := m.events[eventId]; !ok {
		m.events[eventId] = []*ListenersManagerItem{item}
	} else {
		m.events[eventId] = append(m.events[eventId], item)
	}

	m.listeners[listener.Id()] = item
}

func (m *ListenersManager) DeAttach(eventId workers.EventId, listener workers.Listener) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.events[eventId]; !ok {
		return
	}

	item, ok := m.listeners[listener.Id()]
	if !ok {
		return
	}

	for i := len(m.events[eventId]) - 1; i >= 0; i-- {
		if m.events[eventId][i].Listener() == listener {
			m.events[eventId] = append(m.events[eventId][:i], m.events[eventId][i+1:]...)
		}
	}

	item.RemoveEventId(eventId)
	if len(item.EventIds()) == 0 {
		delete(m.listeners, listener.Id())
	}
}

func (m *ListenersManager) Trigger(eventId workers.EventId, args ...interface{}) {
	m.mutex.RLock()
	listeners, ok := m.events[eventId]
	m.mutex.RUnlock()

	if !ok {
		return
	}

	now := time.Now()
	ctx := context.TODO()

	for _, item := range listeners {
		item.Fire(ctx, eventId, now, args...)
	}
}

func (m *ListenersManager) AsyncTrigger(eventId workers.EventId, args ...interface{}) {
	m.mutex.RLock()
	listeners, ok := m.events[eventId]
	m.mutex.RUnlock()

	if !ok {
		return
	}

	now := time.Now()
	ctx := context.TODO()

	for _, item := range listeners {
		go func(ctx context.Context, i *ListenersManagerItem) {
			item.Fire(ctx, eventId, now, args...)
		}(ctx, item)
	}
}
