package manager

import (
	"github.com/kihamo/go-workers"
)

type WorkersManagerItem struct {
	workers.ManagerItemBase

	worker workers.Worker
}

func NewWorkersManagerItem(worker workers.Worker, status workers.WorkerStatus) *WorkersManagerItem {
	item := &WorkersManagerItem{
		worker: worker,
	}
	item.SetStatus(status)

	return item
}

func (w *WorkersManagerItem) Worker() workers.Worker {
	return w.worker
}

func (w *WorkersManagerItem) Id() string {
	return w.worker.Id()
}

func (w *WorkersManagerItem) Metadata() workers.Metadata {
	return workers.Metadata{
		workers.WorkerMetadataStatus: w.Status(),
	}
}

func (w *WorkersManagerItem) Status() workers.Status {
	return workers.WorkerStatus(w.StatusInt64())
}
