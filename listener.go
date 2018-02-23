package workers

import (
	"context"
	"time"
)

type Listener interface {
	Run(context.Context, Event, time.Time, ...interface{})
	Id() string
	Name() string
}

type ListenerWithEvents interface {
	Listener

	Events() []Event
}
