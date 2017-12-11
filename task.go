package workers

import (
	"context"
	"time"
)

type TaskStatus int64

const (
	TaskStatusUndefined TaskStatus = iota
	TaskStatusWait
	TaskStatusProcess
	TaskStatusSuccess
	TaskStatusFail
	TaskStatusRepeatWait
	TaskStatusCancel
)

func (i TaskStatus) Int64() int64 {
	if i < 0 || i >= TaskStatus(len(_TaskStatus_index)-1) {
		return -1
	}

	return int64(i)
}

type Task interface {
	Run(context.Context) (interface{}, error)
	Id() string
	Name() string
	Priority() int64
	Repeats() int64
	Duration() time.Duration
	Timeout() time.Duration
	CreatedAt() time.Time
}
