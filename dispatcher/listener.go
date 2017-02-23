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

func NewDefaultListener() *DefaultListener {
	return &DefaultListener{
		TaskDone: make(chan task.Tasker),
	}
}

func (l *DefaultListener) NotifyTaskDone(t task.Tasker) {
	l.TaskDone <- t
}
