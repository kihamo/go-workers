package dispatcher

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/kihamo/go-workers"
	"github.com/kihamo/go-workers/manager"
)

type SimpleDispatcherResult struct {
	workerItem *manager.WorkersManagerItem
	taskItem   *manager.TasksManagerItem
	result     interface{}
	err        error
	cancel     bool
}

type SimpleDispatcher struct {
	wg sync.WaitGroup
	workers.StatusItemBase

	ctx       context.Context
	ctxCancel context.CancelFunc

	workers   workers.Manager
	tasks     workers.Manager
	listeners *manager.ListenersManager

	allowExecuteTasks       chan struct{}
	tickerAllowExecuteTasks *workers.Ticker
	results                 chan SimpleDispatcherResult
}

func NewSimpleDispatcher() *SimpleDispatcher {
	return NewSimpleDispatcherWithContext(context.Background())
}

func NewSimpleDispatcherWithContext(ctx context.Context) *SimpleDispatcher {
	d := &SimpleDispatcher{
		workers:                 manager.NewWorkersManager(),
		tasks:                   manager.NewTasksManager(),
		listeners:               manager.NewListenersManager(),
		allowExecuteTasks:       make(chan struct{}, 1),
		tickerAllowExecuteTasks: workers.NewTicker(time.Second),
		results:                 make(chan SimpleDispatcherResult),
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

	d.listeners.AsyncTrigger(workers.EventIdWorkerAdd, worker)
	d.notifyAllowExecuteTasks()
	return nil
}

func (d *SimpleDispatcher) RemoveWorker(worker workers.Worker) {
	item := d.workers.GetById(worker.Id())
	if item != nil {
		d.setStatusWorker(item, workers.WorkerStatusCancel)

		workerItem := item.(*manager.WorkersManagerItem)
		workerItem.Cancel()
		workerItem.SetTask(nil)

		d.workers.Remove(item)
		d.listeners.AsyncTrigger(workers.EventIdWorkerRemove, workerItem.Worker())
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

	d.listeners.AsyncTrigger(workers.EventIdTaskAdd, task)
	d.notifyAllowExecuteTasks()
	return nil
}

func (d *SimpleDispatcher) RemoveTask(task workers.Task) {
	item := d.tasks.GetById(task.Id())
	if item != nil {
		d.setStatusTask(item, workers.TaskStatusCancel)

		taskItem := item.(*manager.TasksManagerItem)
		taskItem.Cancel()

		d.tasks.Remove(item)
		d.listeners.AsyncTrigger(workers.EventIdTaskRemove, taskItem.Task())
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

func (d *SimpleDispatcher) AddListener(eventId workers.EventId, listener workers.Listener) error {
	d.listeners.Attach(eventId, listener)
	d.listeners.AsyncTrigger(workers.EventIdListenerAdd, eventId, listener)

	return nil
}

func (d *SimpleDispatcher) RemoveListener(eventId workers.EventId, listener workers.Listener) {
	d.listeners.DeAttach(eventId, listener)
	d.listeners.AsyncTrigger(workers.EventIdListenerRemove, eventId, listener)
}

func (d *SimpleDispatcher) GetListenerMetadata(id string) workers.Metadata {
	if item := d.listeners.GetById(id); item != nil {
		return item.Metadata()
	}

	return nil
}

func (d *SimpleDispatcher) GetListeners() []workers.Listener {
	return d.listeners.Listeners()
}

func (d *SimpleDispatcher) doResultCollector() {
	d.wg.Add(1)

	for {
		select {
		case result := <-d.results:
			result.taskItem.SetCancel(nil)
			result.workerItem.SetCancel(nil)

			if d.IsStatus(workers.DispatcherStatusCancel) {
				continue
			}

			result.workerItem.SetTask(nil)

			if !result.cancel || !result.workerItem.IsStatus(workers.WorkerStatusCancel) {
				d.setStatusWorker(result.workerItem, workers.WorkerStatusWait)
				d.workers.Push(result.workerItem)
			}

			if !result.cancel || !result.taskItem.IsStatus(workers.TaskStatusCancel) {
				if result.err != nil {
					d.setStatusTask(result.taskItem, workers.TaskStatusFail)
				} else {
					d.setStatusTask(result.taskItem, workers.TaskStatusSuccess)
				}

				if repeats := result.taskItem.Task().Repeats(); repeats < 0 || result.taskItem.Attempts() < repeats {
					repeatInterval := result.taskItem.Task().RepeatInterval()
					if repeatInterval > 0 {
						result.taskItem.SetAllowStartAt(time.Now().Add(repeatInterval))
					}

					d.setStatusTask(result.taskItem, workers.TaskStatusRepeatWait)
					d.tasks.Push(result.taskItem)
				} else {
					d.tasks.Remove(result.taskItem)
				}
			}

			d.listeners.AsyncTrigger(workers.EventIdTaskExecuteStop, result.taskItem.Task(), result.workerItem.Worker(), result.result, result.err)
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

					d.listeners.AsyncTrigger(workers.EventIdTaskExecuteStart, castTask.Task(), castWorker.Worker())
					go d.doRunTask(castWorker, castTask)
				}
				// TODO: log else
			}

		case <-d.tickerAllowExecuteTasks.C():
			d.notifyAllowExecuteTasks()

		case <-d.ctx.Done():
			d.tickerAllowExecuteTasks.Stop()
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

	now := time.Now()
	if taskItem.Attempts() == 1 {
		taskItem.SetFirstStartedAt(now)
	}
	taskItem.SetLastStartedAt(now)

	ctx := workers.NewContextWithAttempt(d.ctx, taskItem.Attempts())

	var ctxCancel context.CancelFunc

	timeout := task.Timeout()
	if timeout > 0 {
		ctx, ctxCancel = context.WithTimeout(ctx, timeout)
	} else {
		ctx, ctxCancel = context.WithCancel(ctx)
	}

	defer ctxCancel()
	taskItem.SetCancel(ctxCancel)
	workerItem.SetCancel(ctxCancel)

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
			cancel:     ctx.Err() == context.Canceled,
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
	d.listeners.AsyncTrigger(workers.EventIdDispatcherStatusChanged, d, status, last)
}

func (d *SimpleDispatcher) setStatusWorker(worker workers.ManagerItem, status workers.Status) {
	last := worker.Status()
	worker.SetStatus(status)
	d.listeners.AsyncTrigger(workers.EventIdWorkerStatusChanged, worker.(*manager.WorkersManagerItem).Worker(), status, last)
}

func (d *SimpleDispatcher) setStatusTask(task workers.ManagerItem, status workers.Status) {
	last := task.Status()
	task.SetStatus(status)
	d.listeners.AsyncTrigger(workers.EventIdTaskStatusChanged, task.(*manager.TasksManagerItem).Task(), status, last)
}

func (d *SimpleDispatcher) SetTickerExecuteTasksDuration(t time.Duration) {
	d.tickerAllowExecuteTasks.SetDuration(t)
}
