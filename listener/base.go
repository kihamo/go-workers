package listener

import (
	"fmt"
	"sync/atomic"

	"github.com/google/uuid"
)

type BaseListener struct {
	id   string
	name atomic.Value
}

func (t *BaseListener) Init() {
	t.id = uuid.New().String()
}

func (t *BaseListener) Id() string {
	return t.id
}

func (t *BaseListener) Name() string {
	var name string

	if value := t.name.Load(); value != nil {
		name = value.(string)
	}

	return name
}

func (t *BaseListener) SetName(name string) {
	t.name.Store(name)
}

func (t *BaseListener) String() string {
	return "Listener #" + t.Id()
}

func (t *BaseListener) GoString() string {
	return fmt.Sprintf("%s %#p", t.String(), t)
}
