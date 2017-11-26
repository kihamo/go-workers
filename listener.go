package workers

type Listener interface {
	GetName() string

	NotifyDispatcherStart() error
	NotifyDispatcherDone() error
	NotifyDispatcherChangeStatus(Task) error
	NotifyDispatcherRegisteredWorker() error
	NotifyDispatcherUnregisteredWorker() error
	NotifyDispatcherRegisteredTask() error
	NotifyDispatcherUnregisteredTask() error

	NotifyWorkerStart(Worker) error
	NotifyWorkerDone(Worker) error
	NotifyWorkerChangeStatus(Worker) error
	NotifyWorkerDoneTask(Worker) error

	NotifyTaskStart() error
	NotifyTaskDone(Task) error
	NotifyTaskDoneByTimeout(Task) error
	NotifyTaskChangeStatus(Task) error
}
