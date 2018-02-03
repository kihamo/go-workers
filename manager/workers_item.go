package manager

import (
	"context"
	"sync"

	"github.com/kihamo/go-workers"
)

type WorkersManagerItem struct {
	workers.ManagerItemBase
	mutex sync.RWMutex

	worker workers.Worker
	task   workers.Task
	cancel context.CancelFunc
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
		workers.WorkerMetadataTask:   w.Task(),
		workers.WorkerMetadataLocked: w.IsLocked(),
	}
}

func (w *WorkersManagerItem) Status() workers.Status {
	return workers.WorkerStatus(w.StatusInt64())
}

func (w *WorkersManagerItem) Task() workers.Task {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.task
}

func (w *WorkersManagerItem) SetTask(task workers.Task) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.task = task
}

func (w *WorkersManagerItem) SetCancel(cancel context.CancelFunc) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.cancel = cancel
}

func (w *WorkersManagerItem) Cancel() {
	w.mutex.RLock()
	cancel := w.cancel
	w.mutex.RUnlock()

	if cancel != nil {
		cancel()
	}
}
