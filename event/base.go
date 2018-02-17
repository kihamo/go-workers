package event

import (
	"fmt"

	"github.com/pborman/uuid"
)

type BaseEvent struct {
	id   string
	name string
}

func NewBaseEvent(name string) *BaseEvent {
	return &BaseEvent{
		id:   uuid.New(),
		name: name,
	}
}

func (e *BaseEvent) Id() string {
	return e.id
}

func (e *BaseEvent) Name() string {
	return e.name
}

func (e *BaseEvent) String() string {
	return "Event " + e.Name() + " #" + e.Id()
}

func (e *BaseEvent) GoString() string {
	return fmt.Sprintf("%s %#p", e.String(), e)
}
