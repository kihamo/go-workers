package dispatcher

import (
	"container/list"
	"sync"
	"sync/atomic"

	"github.com/kihamo/go-workers/worker"
)

type workersItem struct {
	element    *list.Element
	lastStatus int64
}

type Workers struct {
	mutex     sync.RWMutex
	container *list.List

	wait  int64
	items map[string]*workersItem
}

func NewWorkers() *Workers {
	return &Workers{
		container: list.New(),
		wait:      0,
		items:     make(map[string]*workersItem),
	}
}

func (q *Workers) Len() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	return q.container.Len()
}

func (q *Workers) Add(w worker.Worker) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	var e *list.Element
	if w.GetStatus() == worker.WorkerStatusWait {
		e = q.container.PushFront(w)
		atomic.AddInt64(&q.wait, 1)
	} else {
		e = q.container.PushBack(w)
	}

	q.items[w.GetId()] = &workersItem{
		element:    e,
		lastStatus: w.GetStatus(),
	}
}

func (q *Workers) Remove(w worker.Worker) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if item, ok := q.items[w.GetId()]; ok {
		q.container.Remove(item.element)

		if item.lastStatus == worker.WorkerStatusWait {
			atomic.AddInt64(&q.wait, -1)
		}

		delete(q.items, w.GetId())
	}
}

func (q *Workers) GetWait() (w worker.Worker) {
	if !q.HasWait() {
		return nil
	}

	q.mutex.Lock()
	defer q.mutex.Unlock()

	for e := q.container.Front(); e != nil; e = e.Next() {
		w = e.Value.(worker.Worker)

		if w.GetStatus() == worker.WorkerStatusWait {
			q.container.Remove(e)

			if q.items[w.GetId()].lastStatus == worker.WorkerStatusWait {
				atomic.AddInt64(&q.wait, -1)
			}

			delete(q.items, w.GetId())
			return w
		}
	}

	return nil
}

func (q *Workers) Update(w worker.Worker) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if item, ok := q.items[w.GetId()]; ok {
		item.element.Value = w

		if w.GetStatus() == worker.WorkerStatusWait {
			q.container.MoveToFront(item.element)

			if item.lastStatus != worker.WorkerStatusWait {
				atomic.AddInt64(&q.wait, 1)
			}
		} else {
			q.container.MoveToBack(item.element)

			if item.lastStatus == worker.WorkerStatusWait {
				atomic.AddInt64(&q.wait, -1)
			}
		}

		item.lastStatus = w.GetStatus()
	}
}

func (q *Workers) HasWait() bool {
	return atomic.LoadInt64(&q.wait) > 0
}

func (q *Workers) GetItems() []worker.Worker {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	items := make([]worker.Worker, q.container.Len())
	i := 0

	for e := q.container.Front(); e != nil; e = e.Next() {
		items[i] = e.Value.(worker.Worker)
		i++
	}

	return items
}
