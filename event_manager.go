package workers

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
)

func (i EventId) Int64() int64 {
	if i < 0 || i >= EventId(len(_EventId_index)-1) {
		return -1
	}

	return int64(i)
}

type EventsManager interface {
	Listeners() map[EventId][]Listener
	Attach(EventId, Listener)
	DeAttach(EventId, Listener)
	Trigger(EventId, ...interface{})
	AsyncTrigger(EventId, ...interface{})
}
