package workers

import (
	"sync/atomic"
)

type StatusItem interface {
	Status() Status
	SetStatus(Status)
	IsStatus(Status) bool
}

type StatusItemBase struct {
	status int64
}

func (i *StatusItemBase) StatusInt64() int64 {
	return atomic.LoadInt64(&i.status)
}

func (i *StatusItemBase) SetStatus(status Status) {
	atomic.StoreInt64(&i.status, status.Int64())
}

func (i *StatusItemBase) IsStatus(status Status) bool {
	return i.StatusInt64() == status.Int64()
}
