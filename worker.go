package workers

import (
	"context"
	"time"
)

type WorkerStatus int64

const (
	WorkerStatusUndefined WorkerStatus = iota
	WorkerStatusWait
	WorkerStatusProcess
	WorkerStatusCancel
)

func (i WorkerStatus) Int64() int64 {
	if i < 0 || i >= WorkerStatus(len(_WorkerStatus_index)-1) {
		return -1
	}

	return int64(i)
}

type Worker interface {
	Cancel() error
	RunTask(context.Context, Task) (interface{}, error)
	Id() string
	CreatedAt() time.Time
}
