package dispatcher

import (
	"fmt"
	"reflect"

	"github.com/kihamo/go-workers/task"
)

type Listener interface {
	GetName() string
	NotifyTaskDone(task.Tasker)
	GetTaskDoneChannel() <-chan task.Tasker
}

type DefaultListener struct {
	name     string
	taskDone chan task.Tasker
}

func NewDefaultListener(n string) *DefaultListener {
	l := &DefaultListener{
		name:     n,
		taskDone: make(chan task.Tasker),
	}

	if l.name == "" {
		t := reflect.TypeOf(*l)
		l.name = fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
	}

	return l
}

func (l *DefaultListener) GetName() string {
	return l.name
}

func (l *DefaultListener) NotifyTaskDone(t task.Tasker) {
	l.taskDone <- t
}

func (l *DefaultListener) GetTaskDoneChannel() <-chan task.Tasker {
	return l.taskDone
}
