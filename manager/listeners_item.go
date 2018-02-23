package manager

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/kihamo/go-workers"
	"github.com/pborman/uuid"
)

type ListenersManagerItem struct {
	mutex       sync.RWMutex
	fires       int64
	eventAll    bool
	events      []workers.Event
	listener    workers.Listener
	id          string
	firstFireAt unsafe.Pointer
	lastFireAt  unsafe.Pointer
}

func NewListenersManagerItem(event workers.Event, listener workers.Listener) *ListenersManagerItem {
	item := &ListenersManagerItem{
		id:       uuid.New(),
		events:   []workers.Event{},
		listener: listener,
	}
	item.AddEvent(event)

	return item
}

func (l *ListenersManagerItem) Id() string {
	return l.listener.Id()
}

func (l *ListenersManagerItem) Events() []workers.Event {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	tmp := make([]workers.Event, len(l.events))
	copy(tmp, l.events)

	return tmp
}

func (l *ListenersManagerItem) EventIsAllowed(event workers.Event) bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if l.eventAll {
		return true
	}

	for _, id := range l.events {
		if id == event {
			return true
		}
	}

	return false
}

func (l *ListenersManagerItem) AddEvent(event workers.Event) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if event == workers.EventAll {
		l.eventAll = true
	}

	for _, id := range l.events {
		if id == event {
			return
		}
	}

	l.events = append(l.events, event)
}

func (l *ListenersManagerItem) RemoveEvent(event workers.Event) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if event == workers.EventAll {
		l.eventAll = false
	}

	for i := len(l.events) - 1; i >= 0; i-- {
		if l.events[i] == event {
			l.events = append(l.events[:i], l.events[i+1:]...)
			break
		}
	}
}

func (l *ListenersManagerItem) Listener() workers.Listener {
	return l.listener
}

func (l *ListenersManagerItem) Fire(ctx context.Context, event workers.Event, t time.Time, args ...interface{}) {
	if !l.EventIsAllowed(event) {
		return
	}

	now := time.Now()

	atomic.AddInt64(&l.fires, 1)
	atomic.StorePointer(&l.lastFireAt, unsafe.Pointer(&now))

	if l.FirstFireAt() == nil {
		atomic.StorePointer(&l.firstFireAt, unsafe.Pointer(&now))
	}

	l.listener.Run(ctx, event, t, args...)
}

func (l *ListenersManagerItem) Metadata() workers.Metadata {
	return workers.Metadata{
		workers.ListenerMetadataFires:        l.Fires(),
		workers.ListenerMetadataFirstFiredAt: l.FirstFireAt(),
		workers.ListenerMetadataLastFireAt:   l.LastFireAt(),
		workers.ListenerMetadataEvents:       l.Events(),
	}
}

func (l *ListenersManagerItem) Fires() int64 {
	return atomic.LoadInt64(&l.fires)
}

func (l *ListenersManagerItem) FirstFireAt() *time.Time {
	p := atomic.LoadPointer(&l.firstFireAt)
	return (*time.Time)(p)
}

func (l *ListenersManagerItem) LastFireAt() *time.Time {
	p := atomic.LoadPointer(&l.lastFireAt)
	return (*time.Time)(p)
}
