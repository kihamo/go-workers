package dispatcher

import (
	"context"
	"errors"
	"sync"

	"github.com/kihamo/go-workers"
	"github.com/kihamo/go-workers/manager"
)

type SimpleDispatcherResult struct {
	workerItem *manager.WorkersManagerItem
	taskItem   *manager.TasksManagerItem
	result     interface{}
	err        error
}

type SimpleDispatcher struct {
	wg sync.WaitGroup
	workers.StatusItemBase

	ctx       context.Context
	ctxCancel context.CancelFunc

	workers workers.Manager
	tasks   workers.Manager
	events  workers.EventsManager

	allowExecuteTasks chan struct{}
	results           chan SimpleDispatcherResult
}

func NewSimpleDispatcher() *SimpleDispatcher {
	return NewSimpleDispatcherWithContext(context.Background())
}

func NewSimpleDispatcherWithContext(ctx context.Context) *SimpleDispatcher {
	d := &SimpleDispatcher{
		workers:           manager.NewWorkersManager(),
		tasks:             manager.NewTasksManager(),
		events:            manager.NewEventsManager(),
		allowExecuteTasks: make(chan struct{}, 1),
		results:           make(chan SimpleDispatcherResult),
	}

	d.setStatusDispatcher(workers.DispatcherStatusWait)

	d.ctx, d.ctxCancel = context.WithCancel(ctx)
	return d
}

func (d *SimpleDispatcher) Context() context.Context {
	return d.ctx
}

func (d *SimpleDispatcher) Run() error {
	if !d.IsStatus(workers.DispatcherStatusWait) {
		return errors.New("Dispatcher is running")
	}

	d.setStatusDispatcher(workers.DispatcherStatusProcess)

	go d.doResultCollector()
	go d.doDispatch()
	d.notifyAllowExecuteTasks()

	<-d.ctx.Done()
	d.setStatusDispatcher(workers.DispatcherStatusCancel)

	for _, w := range d.workers.GetAll() {
		d.setStatusWorker(w, workers.WorkerStatusCancel)
	}

	for _, t := range d.tasks.GetAll() {
		d.setStatusTask(t, workers.WorkerStatusCancel)
	}

	d.wg.Wait()
	d.setStatusDispatcher(workers.DispatcherStatusWait)

	return nil
}

func (d *SimpleDispatcher) Cancel() error {
	d.ctxCancel()
	return d.ctx.Err()
}

func (d *SimpleDispatcher) Metadata() workers.Metadata {
	return workers.Metadata{
		workers.DispatcherMetadataWorkersUnlocked: d.workers.UnlockedCount(),
		workers.DispatcherMetadataTasksUnlocked:   d.tasks.UnlockedCount(),
	}
}

func (d *SimpleDispatcher) Status() workers.DispatcherStatus {
	return workers.DispatcherStatus(d.StatusInt64())
}

func (d *SimpleDispatcher) SetStatus(status workers.Status) {
	panic("Change status nof allowed")
}

func (d *SimpleDispatcher) AddWorker(worker workers.Worker) error {
	err := d.workers.Push(manager.NewWorkersManagerItem(worker, workers.WorkerStatusWait))
	if err != nil {
		return err
	}

	d.events.AsyncTrigger(workers.EventIdWorkerAdd, worker)
	d.notifyAllowExecuteTasks()
	return nil
}

func (d *SimpleDispatcher) RemoveWorker(worker workers.Worker) {
	item := d.workers.GetById(worker.Id())
	if item != nil {
		cancel := item.IsStatus(workers.WorkerStatusProcess)
		d.setStatusWorker(item, workers.WorkerStatusCancel)

		if cancel {
			worker.Cancel()
		}

		d.workers.Remove(item)
		d.events.AsyncTrigger(workers.EventIdWorkerRemove, item.(*manager.WorkersManagerItem).Worker())
	}

	return
}

func (d *SimpleDispatcher) GetWorkerMetadata(id string) workers.Metadata {
	if item := d.workers.GetById(id); item != nil {
		return item.Metadata()
	}

	return nil
}

func (d *SimpleDispatcher) GetWorkers() []workers.Worker {
	all := d.workers.GetAll()
	collection := make([]workers.Worker, 0, len(all))

	for _, item := range all {
		collection = append(collection, item.(*manager.WorkersManagerItem).Worker())
	}

	return collection
}

func (d *SimpleDispatcher) AddTask(task workers.Task) error {
	err := d.tasks.Push(manager.NewTasksManagerItem(task, workers.TaskStatusWait))
	if err != nil {
		return err
	}

	d.events.AsyncTrigger(workers.EventIdTaskAdd, task)
	d.notifyAllowExecuteTasks()
	return nil
}

func (d *SimpleDispatcher) RemoveTask(task workers.Task) {
	item := d.tasks.GetById(task.Id())
	if item != nil {
		cancel := item.IsStatus(workers.TaskStatusProcess) || item.IsStatus(workers.TaskStatusRepeatWait)
		d.setStatusTask(item, workers.TaskStatusCancel)

		if cancel {
			task.Cancel()
		}

		d.tasks.Remove(item)
		d.events.AsyncTrigger(workers.EventIdTaskRemove, item.(*manager.TasksManagerItem).Task())
	}

	return
}

func (d *SimpleDispatcher) GetTaskMetadata(id string) workers.Metadata {
	if item := d.tasks.GetById(id); item != nil {
		return item.Metadata()
	}

	return nil
}

func (d *SimpleDispatcher) GetTasks() []workers.Task {
	all := d.tasks.GetAll()
	collection := make([]workers.Task, 0, len(all))

	for _, item := range all {
		collection = append(collection, item.(*manager.TasksManagerItem).Task())
	}

	return collection
}

func (d *SimpleDispatcher) AddListener(id workers.EventId, listener workers.Listener) {
	d.events.Attach(id, listener)
	d.events.AsyncTrigger(workers.EventIdListenerAdd, id, listener)
}

func (d *SimpleDispatcher) RemoveListener(id workers.EventId, listener workers.Listener) {
	d.events.DeAttach(id, listener)
	d.events.AsyncTrigger(workers.EventIdListenerRemove, id, listener)
}

func (d *SimpleDispatcher) GetListeners() map[workers.EventId][]workers.Listener {
	return d.events.Listeners()
}

func (d *SimpleDispatcher) doResultCollector() {
	d.wg.Add(1)

	for {
		select {
		case result := <-d.results:
			if d.IsStatus(workers.DispatcherStatusCancel) {
				continue
			}

			result.workerItem.SetTask(nil)
			d.setStatusWorker(result.workerItem, workers.WorkerStatusWait)
			d.workers.Push(result.workerItem)

			if result.err != nil {
				d.setStatusTask(result.taskItem, workers.TaskStatusFail)
			} else {
				d.setStatusTask(result.taskItem, workers.TaskStatusSuccess)
			}

			if repeats := result.taskItem.Task().Repeats(); repeats < 0 || result.taskItem.Attempts() < repeats {
				d.setStatusTask(result.taskItem, workers.TaskStatusRepeatWait)
				d.tasks.Push(result.taskItem)
			} else {
				d.tasks.Remove(result.taskItem)
			}

			d.events.AsyncTrigger(workers.EventIdTaskExecuteStop, result.taskItem.Task(), result.workerItem.Worker(), result.result, result.err)
			d.notifyAllowExecuteTasks()

		case <-d.ctx.Done():
			d.wg.Done()
			return
		}
	}
}

func (d *SimpleDispatcher) doDispatch() {
	d.wg.Add(1)

	for {
		select {
		case <-d.allowExecuteTasks:
			if d.IsStatus(workers.DispatcherStatusCancel) {
				continue
			}

			for d.tasks.UnlockedCount() > 0 && d.workers.UnlockedCount() > 0 {
				pullWorker := d.workers.Pull()
				pullTask := d.tasks.Pull()

				if pullWorker != nil && pullTask != nil {
					castWorker := pullWorker.(*manager.WorkersManagerItem)
					castTask := pullTask.(*manager.TasksManagerItem)

					d.events.AsyncTrigger(workers.EventIdTaskExecuteStart, castTask.Task(), castWorker.Worker())
					go d.doRunTask(castWorker, castTask)
				}
				// TODO: log else
			}

		case <-d.ctx.Done():
			d.wg.Done()
			return
		}
	}
}

func (d *SimpleDispatcher) doRunTask(workerItem *manager.WorkersManagerItem, taskItem *manager.TasksManagerItem) {
	d.wg.Add(1)
	defer d.wg.Done()

	task := taskItem.Task()

	workerItem.SetTask(task)
	d.setStatusWorker(workerItem, workers.WorkerStatusProcess)

	taskItem.SetAttempts(taskItem.Attempts() + 1)
	d.setStatusTask(taskItem, workers.TaskStatusProcess)

	ctx := workers.NewContextWithAttempt(d.ctx, taskItem.Attempts())

	var ctxCancel context.CancelFunc

	timeout := task.Timeout()
	if timeout > 0 {
		ctx, ctxCancel = context.WithTimeout(ctx, timeout)
	} else {
		ctx, ctxCancel = context.WithCancel(ctx)
	}

	defer ctxCancel()

	done := make(chan SimpleDispatcherResult, 1)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				done <- SimpleDispatcherResult{
					workerItem: workerItem,
					taskItem:   taskItem,
					result:     err,
					err:        errors.New("Panic recovered"),
				}
			}
		}()

		result, err := workerItem.Worker().RunTask(ctx, task)

		done <- SimpleDispatcherResult{
			workerItem: workerItem,
			taskItem:   taskItem,
			result:     result,
			err:        err,
		}
	}()

	select {
	case <-ctx.Done():
		// TODO:
		//<-done

		d.results <- SimpleDispatcherResult{
			workerItem: workerItem,
			taskItem:   taskItem,
			err:        ctx.Err(),
		}

	case r := <-done:
		d.results <- r
	}
}

func (d *SimpleDispatcher) notifyAllowExecuteTasks() {
	if d.IsStatus(workers.DispatcherStatusProcess) && len(d.allowExecuteTasks) == 0 {
		d.allowExecuteTasks <- struct{}{}
	}
}

func (d *SimpleDispatcher) setStatusDispatcher(status workers.Status) {
	last := d.Status()
	d.StatusItemBase.SetStatus(status)
	d.events.AsyncTrigger(workers.EventIdDispatcherStatusChanged, d, status, last)
}

func (d *SimpleDispatcher) setStatusWorker(worker workers.ManagerItem, status workers.Status) {
	last := worker.Status()
	worker.SetStatus(status)
	d.events.AsyncTrigger(workers.EventIdWorkerStatusChanged, worker.(*manager.WorkersManagerItem).Worker(), status, last)
}

func (d *SimpleDispatcher) setStatusTask(task workers.ManagerItem, status workers.Status) {
	last := task.Status()
	task.SetStatus(status)
	d.events.AsyncTrigger(workers.EventIdTaskStatusChanged, task.(*manager.TasksManagerItem).Task(), status, last)
}
