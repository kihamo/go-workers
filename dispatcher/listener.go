package dispatcher

import (
	"fmt"
	"log"
	"reflect"

	"github.com/kihamo/go-workers/task"
)

type Listener interface {
	GetName() string
	NotifyTaskDone(task.Tasker) error
	NotifyTaskDoneTimeout(task.Tasker) error
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

func (l *DefaultListener) NotifyTaskDone(t task.Tasker) error {
	l.taskDone <- t

	return nil
}

func (l *DefaultListener) NotifyTaskDoneTimeout(t task.Tasker) error {
	log.Printf("Cancel send event to listener \"%s\" by timeout for task \"%s\"", l.GetName(), t.GetName())
	<-l.taskDone

	return nil
}

func (l *DefaultListener) GetTaskDoneChannel() <-chan task.Tasker {
	return l.taskDone
}
