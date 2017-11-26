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
		allowExecuteTasks: make(chan struct{}, 1),
		results:           make(chan SimpleDispatcherResult),
	}

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

	d.StatusItemBase.SetStatus(workers.DispatcherStatusProcess)

	go d.doResultCollector()
	go d.doDispatch()

	<-d.ctx.Done()
	d.StatusItemBase.SetStatus(workers.DispatcherStatusCancel)

	d.wg.Wait()
	d.StatusItemBase.SetStatus(workers.DispatcherStatusWait)

	return nil
}

func (d *SimpleDispatcher) Cancel() error {
	d.ctxCancel()
	return d.ctx.Err()
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

	d.notifyAllowExecuteTasks()
	return nil
}

func (d *SimpleDispatcher) RemoveWorker(worker workers.Worker) {
	for _, item := range d.workers.GetAll() {
		if item.Id() == worker.Id() {
			d.workers.Remove(item)
		}
	}
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

	d.notifyAllowExecuteTasks()
	return nil
}

func (d *SimpleDispatcher) RemoveTask(task workers.Task) {
	for _, item := range d.tasks.GetAll() {
		if item.Id() == task.Id() {
			d.tasks.Remove(item)
		}
	}

	return
}

func (d *SimpleDispatcher) GetTasks() []workers.Task {
	all := d.tasks.GetAll()
	collection := make([]workers.Task, 0, len(all))

	for _, item := range all {
		collection = append(collection, item.(*manager.TasksManagerItem).Task())
	}

	return collection
}

func (d *SimpleDispatcher) doResultCollector() {
	d.wg.Add(1)

	for {
		select {
		case result := <-d.results:
			d.tasks.Remove(result.taskItem)

			result.workerItem.SetStatus(workers.WorkerStatusWait)
			d.workers.Push(result.workerItem)

			if result.err != nil {
				result.taskItem.SetStatus(workers.TaskStatusFail)
			} else {
				result.taskItem.SetStatus(workers.TaskStatusSuccess)
			}

			if repeats := result.taskItem.Task().Repeats(); repeats == -1 || result.taskItem.Attempts() < repeats {
				result.taskItem.SetStatus(workers.TaskStatusRepeatWait)
				d.tasks.Push(result.taskItem)
			}

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
			if d.tasks.Check() && d.workers.Check() {
				pullWorker := d.workers.Pull()
				pullTask := d.tasks.Pull()

				if pullWorker != nil && pullTask != nil {
					pullWorker.SetStatus(workers.WorkerStatusProcess)
					pullTask.SetStatus(workers.TaskStatusProcess)

					go d.doRunTask(pullWorker.(*manager.WorkersManagerItem), pullTask.(*manager.TasksManagerItem))
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

	taskItem.SetAttempts(taskItem.Attempts() + 1)
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
