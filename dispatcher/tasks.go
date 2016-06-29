package dispatcher

import (
	"container/list"
	"sync"

	"github.com/kihamo/go-workers/task"
)

type tasksItem struct {
	element    *list.Element
	lastStatus int64
}

type Tasks struct {
	mutex     sync.RWMutex
	container *list.List

	wait  int
	items map[string]*tasksItem
}

func NewTasks() *Tasks {
	return &Tasks{
		container: list.New(),
		wait:      0,
		items:     make(map[string]*tasksItem),
	}
}

func (q *Tasks) Len() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	return q.container.Len()
}

func (q *Tasks) Add(t task.Tasker) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	status := t.GetStatus()
	if status == task.TaskStatusWait || status == task.TaskStatusRepeatWait {
		q.wait++
	}

	q.items[t.GetId()] = &tasksItem{
		element:    q.container.PushBack(t),
		lastStatus: status,
	}
}

func (q *Tasks) GetWait() (t task.Tasker) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if q.wait == 0 {
		return nil
	}

	if e := q.container.Front(); e != nil {
		t = q.container.Remove(e).(task.Tasker)

		status := q.items[t.GetId()].lastStatus
		if status == task.TaskStatusWait || status == task.TaskStatusRepeatWait {
			q.wait--
		}

		delete(q.items, t.GetId())
	}

	return t
}

func (q *Tasks) Remove(t task.Tasker) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if item, ok := q.items[t.GetId()]; ok {
		q.container.Remove(item.element)

		if item.lastStatus == task.TaskStatusWait || item.lastStatus == task.TaskStatusRepeatWait {
			q.wait--
		}

		delete(q.items, t.GetId())
	}
}

func (q *Tasks) HasWait() bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	return q.wait > 0
}

func (q *Tasks) GetWaitsCount() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	return q.wait
}
