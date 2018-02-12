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
	eventIds    []workers.EventId
	listener    workers.Listener
	id          string
	firstFireAt unsafe.Pointer
	lastFireAt  unsafe.Pointer
}

func NewListenersManagerItem(eventId workers.EventId, listener workers.Listener) *ListenersManagerItem {
	item := &ListenersManagerItem{
		id:       uuid.New(),
		eventIds: []workers.EventId{},
		listener: listener,
	}
	item.AddEventId(eventId)

	return item
}

func (l *ListenersManagerItem) Id() string {
	return l.listener.Id()
}

func (l *ListenersManagerItem) EventIds() []workers.EventId {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	tmp := make([]workers.EventId, len(l.eventIds))
	copy(tmp, l.eventIds)

	return tmp
}

func (l *ListenersManagerItem) EventIdIsAllowed(eventId workers.EventId) bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if l.eventAll {
		return true
	}

	for _, id := range l.eventIds {
		if id == eventId {
			return true
		}
	}

	return false
}

func (l *ListenersManagerItem) AddEventId(eventId workers.EventId) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if eventId == workers.EventIdAll {
		l.eventAll = true
	}

	for _, id := range l.eventIds {
		if id == eventId {
			return
		}
	}

	l.eventIds = append(l.eventIds, eventId)
}

func (l *ListenersManagerItem) RemoveEventId(eventId workers.EventId) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if eventId == workers.EventIdAll {
		l.eventAll = false
	}

	for i := len(l.eventIds) - 1; i >= 0; i-- {
		if l.eventIds[i] == eventId {
			l.eventIds = append(l.eventIds[:i], l.eventIds[i+1:]...)
			break
		}
	}
}

func (l *ListenersManagerItem) Listener() workers.Listener {
	return l.listener
}

func (l *ListenersManagerItem) Fire(ctx context.Context, eventId workers.EventId, t time.Time, args ...interface{}) {
	if !l.EventIdIsAllowed(eventId) {
		return
	}

	now := time.Now()

	atomic.AddInt64(&l.fires, 1)
	atomic.StorePointer(&l.lastFireAt, unsafe.Pointer(&now))

	if l.FirstFireAt() == nil {
		atomic.StorePointer(&l.firstFireAt, unsafe.Pointer(&now))
	}

	l.listener.Run(ctx, eventId, t, args...)
}

func (l *ListenersManagerItem) Metadata() workers.Metadata {
	return workers.Metadata{
		workers.ListenerMetadataFires:        l.Fires(),
		workers.ListenerMetadataFirstFiredAt: l.FirstFireAt(),
		workers.ListenerMetadataLastFireAt:   l.LastFireAt(),
		workers.ListenerMetadataEventIds:     l.EventIds(),
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
