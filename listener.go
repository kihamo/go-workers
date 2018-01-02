package workers

import (
	"context"
	"time"
)

type EventId int64

const (
	EventIdDispatcherStatusChanged EventId = iota
	EventIdWorkerAdd
	EventIdWorkerRemove
	EventIdWorkerExecuteStart
	EventIdWorkerExecuteStop
	EventIdWorkerStatusChanged
	EventIdTaskAdd
	EventIdTaskRemove
	EventIdTaskExecuteStart
	EventIdTaskExecuteStop
	EventIdTaskStatusChanged
	EventIdListenerAdd
	EventIdListenerRemove
)

func (i EventId) Int64() int64 {
	if i < 0 || i >= EventId(len(_EventId_index)-1) {
		return -1
	}

	return int64(i)
}

type Listener interface {
	Run(context.Context, EventId, time.Time, ...interface{})
	Id() string
	Name() string
}
