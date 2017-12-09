package manager

import (
	"sync"
	"time"

	"github.com/kihamo/go-workers"
)

type EventsManager struct {
	mutex     sync.RWMutex
	listeners map[workers.EventId][]workers.Listener
}

func NewEventsManager() *EventsManager {
	return &EventsManager{
		listeners: map[workers.EventId][]workers.Listener{},
	}
}

func (m *EventsManager) ListenersByEventId(id workers.EventId) []workers.Listener {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if _, ok := m.listeners[id]; ok {
		listeners := make([]workers.Listener, len(m.listeners[id]))
		copy(listeners, m.listeners[id])
		return listeners
	}

	return nil
}

func (m *EventsManager) Listeners() map[workers.EventId][]workers.Listener {
	m.mutex.RLock()

	listeners := make(map[workers.EventId][]workers.Listener, len(m.listeners))
	for e, l := range m.listeners {
		listeners[e] = l[:]
	}

	m.mutex.RUnlock()

	return listeners
}

func (m *EventsManager) Attach(id workers.EventId, listener workers.Listener) {
	m.mutex.Lock()

	if _, ok := m.listeners[id]; !ok {
		m.listeners[id] = []workers.Listener{}
	}

	m.listeners[id] = append(m.listeners[id], listener)

	m.mutex.Unlock()
}

func (m *EventsManager) DeAttach(id workers.EventId, listener workers.Listener) {
	m.mutex.Lock()

	if _, ok := m.listeners[id]; ok {
		for i := len(m.listeners[id]) - 1; i >= 0; i-- {
			if &(m.listeners[id][i]) == &listener {
				m.listeners[id] = append(m.listeners[id][:i], m.listeners[id][i+1:]...)
			}
		}
	}

	m.mutex.Unlock()
}

func (m *EventsManager) Trigger(id workers.EventId, args ...interface{}) {
	now := time.Now()

	for _, l := range m.ListenersByEventId(id) {
		l(now, args...)
	}
}

func (m *EventsManager) AsyncTrigger(id workers.EventId, args ...interface{}) {
	now := time.Now()

	for _, l := range m.ListenersByEventId(id) {
		go func(f workers.Listener) {
			f(now, args...)
		}(l)
	}
}
