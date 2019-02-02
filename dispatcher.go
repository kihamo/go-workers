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
	if i < 0 || i >= DispatcherStatus(len(_DispatcherStatusIndex)-1) {
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
	GetWorkerMetadata(string) Metadata
	GetWorkers() []Worker

	AddTask(Task) error
	RemoveTask(Task)
	GetTaskMetadata(string) Metadata
	GetTasks() []Task

	AddListener(Event, Listener) error
	RemoveListener(Event, Listener)
	GetListenerMetadata(string) Metadata
	GetListeners() []Listener
}
