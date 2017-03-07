package dispatcher

import (
	"github.com/kihamo/go-workers/task"
)

type Listener interface {
	NotifyTaskDone(task.Tasker)
	GetTaskDoneChannel() <-chan task.Tasker
}

type DefaultListener struct {
	taskDone chan task.Tasker
}

func NewDefaultListener() *DefaultListener {
	return &DefaultListener{
		taskDone: make(chan task.Tasker),
	}
}

func (l *DefaultListener) NotifyTaskDone(t task.Tasker) {
	l.taskDone <- t
}

func (l *DefaultListener) GetTaskDoneChannel() <-chan task.Tasker {
	return l.taskDone
}
