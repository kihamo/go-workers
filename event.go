package workers

import (
	"github.com/kihamo/go-workers/event"
)

var (
	EventAll                     = event.NewBaseEvent("All")
	EventDispatcherStatusChanged = event.NewBaseEvent("DispatcherStatusChanged")
	EventWorkerAdd               = event.NewBaseEvent("WorkerAdd")
	EventWorkerRemove            = event.NewBaseEvent("WorkerRemove")
	EventWorkerExecuteStart      = event.NewBaseEvent("WorkerExecuteStart")
	EventWorkerExecuteStop       = event.NewBaseEvent("WorkerExecuteStop")
	EventWorkerStatusChanged     = event.NewBaseEvent("WorkerStatusChanged")
	EventTaskAdd                 = event.NewBaseEvent("TaskAdd")
	EventTaskRemove              = event.NewBaseEvent("TaskRemove")
	EventTaskExecuteStart        = event.NewBaseEvent("TaskExecuteStart")
	EventTaskExecuteStop         = event.NewBaseEvent("TaskExecuteStop")
	EventTaskStatusChanged       = event.NewBaseEvent("TaskStatusChanged")
	EventListenerAdd             = event.NewBaseEvent("ListenerAdd")
	EventListenerRemove          = event.NewBaseEvent("ListenerRemove")
)

type Event interface {
	Id() string
	Name() string
}
