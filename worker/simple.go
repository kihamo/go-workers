package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/kihamo/go-workers"
	"github.com/pborman/uuid"
)

type SimpleWorker struct {
	id        string
	createdAt time.Time
}

func NewSimpleWorker() *SimpleWorker {
	return &SimpleWorker{
		id:        uuid.New(),
		createdAt: time.Now(),
	}
}

func (w *SimpleWorker) RunTask(ctx context.Context, task workers.Task) (interface{}, error) {
	return task.Run(ctx)
}

func (w *SimpleWorker) Id() string {
	return w.id
}

func (w *SimpleWorker) CreatedAt() time.Time {
	return w.createdAt
}

func (w *SimpleWorker) String() string {
	return "Worker #" + w.Id()
}

func (w *SimpleWorker) GoString() string {
	return fmt.Sprintf("%s %#p", w.String(), w)
}
