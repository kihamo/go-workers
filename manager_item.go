package workers

import (
	"sync"
	"sync/atomic"
)

const (
	LockFalse = int64(iota)
	LockTrue
)

type ManagerItem interface {
	sync.Locker
	StatusItem

	IsLocked() bool

	Id() string
	Metadata() Metadata
}

type ManagerItemBase struct {
	StatusItemBase

	lock int64
}

func (i *ManagerItemBase) Lock() {
	atomic.StoreInt64(&i.lock, LockTrue)
}

func (i *ManagerItemBase) Unlock() {
	atomic.StoreInt64(&i.lock, LockFalse)
}

func (i *ManagerItemBase) IsLocked() bool {
	return atomic.LoadInt64(&i.lock) == LockTrue
}

func (i *ManagerItemBase) Metadata() Metadata {
	return nil
}
