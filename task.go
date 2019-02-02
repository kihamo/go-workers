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
	if i < 0 || i >= TaskStatus(len(_TaskStatusIndex)-1) {
		return -1
	}

	return int64(i)
}

type Task interface {
	Run(context.Context) (interface{}, error)
	Id() string
	Name() string
	// приоритет выполения, чем меньше значение тем приоритетнее в очереди задач
	Priority() int64
	Repeats() int64
	// интервал между повторениями
	RepeatInterval() time.Duration
	// таймаут на выполнения задачи
	Timeout() time.Duration
	CreatedAt() time.Time
	// для отложенного запуска
	StartedAt() *time.Time
}
