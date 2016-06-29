package dispatcher

import (
	"container/list"
	"sync"

	"github.com/kihamo/go-workers/worker"
)

type workersItem struct {
	element    *list.Element
	lastStatus int64
}

type Workers struct {
	mutex     sync.RWMutex
	container *list.List

	wait  int
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
		q.wait++
	} else {
		e = q.container.PushBack(w)
	}

	q.items[w.GetId()] = &workersItem{
		element:    e,
		lastStatus: w.GetStatus(),
	}
}

func (q *Workers) GetWait() (w worker.Worker) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.wait == 0 {
		return nil
	}

	for e := q.container.Front(); e != nil; e = e.Next() {
		w = e.Value.(worker.Worker)

		if w.GetStatus() == worker.WorkerStatusWait {
			q.container.Remove(e)

			if q.items[w.GetId()].lastStatus == worker.WorkerStatusWait {
				q.wait--
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
				q.wait++
			}
		} else {
			q.container.MoveToBack(item.element)

			if item.lastStatus == worker.WorkerStatusWait {
				q.wait--
			}
		}

		item.lastStatus = w.GetStatus()
	}
}

func (q *Workers) HasWait() bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	return q.wait > 0
}

func (q *Workers) GetWaitsCount() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	return q.wait
}
