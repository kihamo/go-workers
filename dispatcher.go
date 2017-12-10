package workers

import (
	"context"
)

type DispatcherStatus int64

const (
	DispatcherStatusUndefined DispatcherStatus = iota
	DispatcherStatusWait
	DispatcherStatusProcess
	DispatcherStatusCancel
)

func (i DispatcherStatus) Int64() int64 {
	if i < 0 || i >= DispatcherStatus(len(_DispatcherStatus_index)-1) {
		return -1
	}

	return int64(i)
}

type Dispatcher interface {
	Context() context.Context
	Run() error
	Cancel() error
	Status() DispatcherStatus
	Metadata() Metadata

	AddWorker(Worker) error
	RemoveWorker(Worker)
	GetWorkers() []Worker
	GetWorkerMetadata(string) Metadata

	AddTask(Task) error
	RemoveTask(Task)
	GetTasks() []Task
	GetTaskMetadata(string) Metadata

	AddListener(EventId, Listener)
	RemoveListener(EventId, Listener)
	GetListeners() map[EventId][]Listener
}
