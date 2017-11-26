package workers

import (
	"context"
)

type DispatcherStatus int64

const (
	DispatcherStatusWait DispatcherStatus = iota
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

	AddWorker(Worker) error
	RemoveWorker(Worker)
	GetWorkers() []Worker

	AddTask(Task) error
	RemoveTask(Task)
	GetTasks() []Task

	/*
		AddListener(Listener)
		RemoveListener(Listener)
		GetListeners() []Listener
	*/
}
