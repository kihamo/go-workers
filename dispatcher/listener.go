package dispatcher

import (
	"github.com/kihamo/go-workers/task"
)

type Listener interface {
	NotifyTaskDone(task.Tasker)
}

type DefaultListener struct {
	TaskDone chan task.Tasker
}

func NewDefaultListener(b int) *DefaultListener {
	return &DefaultListener{
		TaskDone: make(chan task.Tasker, b),
	}
}

func (l *DefaultListener) NotifyTaskDone(t task.Tasker) {
	l.TaskDone <- t
}
