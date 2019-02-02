package workers

import (
	"sync"
	"sync/atomic"
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

	lock uint32
}

func (i *ManagerItemBase) Lock() {
	atomic.StoreUint32(&i.lock, 1)
}

func (i *ManagerItemBase) Unlock() {
	atomic.StoreUint32(&i.lock, 0)
}

func (i *ManagerItemBase) IsLocked() bool {
	return atomic.LoadUint32(&i.lock) == 1
}

func (i *ManagerItemBase) Metadata() Metadata {
	return nil
}
