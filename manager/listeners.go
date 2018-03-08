package manager

import (
	"context"
	"sync"
	"time"

	"github.com/kihamo/go-workers"
)

type ListenersManager struct {
	mutex     sync.RWMutex
	events    map[workers.Event][]*ListenersManagerItem
	listeners map[string]*ListenersManagerItem
}

func NewListenersManager() *ListenersManager {
	return &ListenersManager{
		events:    map[workers.Event][]*ListenersManagerItem{},
		listeners: map[string]*ListenersManagerItem{},
	}
}

func (m *ListenersManager) AddListener(listener workers.ListenerWithEvents) {
	events := listener.Events()

	if len(events) > 0 {
		for _, event := range events {
			m.Attach(event, listener)
		}
	}
}

func (m *ListenersManager) RemoveListener(listener workers.ListenerWithEvents) {
	events := listener.Events()

	if len(events) > 0 {
		for _, event := range events {
			m.DeAttach(event, listener)
		}
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

func (m *ListenersManager) GetEventById(id string) workers.Event {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for event := range m.events {
		if event.Id() == id {
			return event
		}
	}

	return nil
}

func (m *ListenersManager) Attach(event workers.Event, listener workers.Listener) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	item, ok := m.listeners[listener.Id()]
	if !ok {
		item = NewListenersManagerItem(event, listener)
	}
	item.AddEvent(event)

	if _, ok := m.events[event]; !ok {
		m.events[event] = []*ListenersManagerItem{item}
	} else {
		m.events[event] = append(m.events[event], item)
	}

	m.listeners[listener.Id()] = item
}

func (m *ListenersManager) DeAttach(event workers.Event, listener workers.Listener) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.events[event]; !ok {
		return
	}

	item, ok := m.listeners[listener.Id()]
	if !ok {
		return
	}

	for i := len(m.events[event]) - 1; i >= 0; i-- {
		if m.events[event][i].Listener() == listener {
			m.events[event] = append(m.events[event][:i], m.events[event][i+1:]...)
		}
	}

	item.RemoveEvent(event)
	if len(item.Events()) == 0 {
		delete(m.listeners, listener.Id())
	}
}

func (m *ListenersManager) Trigger(event workers.Event, args ...interface{}) {
	listeners := m.listenersForEvent(event)
	if len(listeners) == 0 {
		return
	}

	now := time.Now()
	ctx := context.TODO()

	for _, item := range listeners {
		item.Fire(ctx, event, now, args...)
	}
}

func (m *ListenersManager) AsyncTrigger(event workers.Event, args ...interface{}) {
	listeners := m.listenersForEvent(event)

	if len(listeners) == 0 {
		return
	}

	now := time.Now()
	ctx := context.TODO()

	for _, item := range listeners {
		go func(ctx context.Context, i *ListenersManagerItem) {
			i.Fire(ctx, event, now, args...)
		}(ctx, item)
	}
}

func (m *ListenersManager) listenersForEvent(event workers.Event) []*ListenersManagerItem {
	m.mutex.RLock()
	listeners := make([]*ListenersManagerItem, 0, len(m.listeners))
	listenersByEvent, okByEvent := m.events[event]
	m.mutex.RUnlock()

	if okByEvent {
		listeners = listenersByEvent
	}

	if event == workers.EventAll {
		return listeners
	}

	m.mutex.RLock()
	listenersAll, okAll := m.events[workers.EventAll]
	m.mutex.RUnlock()

	if okAll {
		for _, listener := range listenersAll {
			listeners = append(listeners, listener)
		}
	}

	return listeners
}
