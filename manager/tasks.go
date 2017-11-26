package manager

import (
	"container/heap"
	"errors"
	"sync/atomic"

	"github.com/kihamo/go-workers"
)

type TasksManager struct {
	queue      *tasksQueue
	waitCounts int64
}

func NewTasksManager() *TasksManager {
	return &TasksManager{
		queue:      newTasksQueue(),
		waitCounts: 0,
	}
}

func (m *TasksManager) Push(task workers.ManagerItem) error {
	if task == nil {
		return errors.New("TasksManagerItem can't be nil")
	}

	t := task.(*TasksManagerItem)

	heap.Push(m.queue, t)
	if t.IsWait() {
		atomic.AddInt64(&m.waitCounts, 1)
	}

	return nil
}

func (m *TasksManager) Pull() workers.ManagerItem {
	if !m.Check() {
		return nil
	}

	for item := heap.Pop(m.queue); item != nil; item = heap.Pop(m.queue) {
		mItem := item.(workers.ManagerItem)

		if item.(*TasksManagerItem).IsWait() {
			mItem.Lock()
			atomic.AddInt64(&m.waitCounts, -1)

			heap.Push(m.queue, item)
			return mItem
		}

		heap.Push(m.queue, item)
	}

	return nil
}

func (m *TasksManager) Remove(item workers.ManagerItem) {
	t := item.(*TasksManagerItem)

	i := t.Index()
	if i > 0 && i < m.queue.Len() {
		// TODO: Fix or Remove ???
		heap.Fix(m.queue, i)

		if !t.IsLocked() {
			// TODO: check is exists
			atomic.AddInt64(&m.waitCounts, -1)
		} else {
			t.Unlock()
		}
	}

	t.setIndex(-1)
}

func (m *TasksManager) GetAll() []workers.ManagerItem {
	all := m.queue.All()
	collection := make([]workers.ManagerItem, 0, len(all))

	for _, t := range all {
		collection = append(collection, t)
	}

	return collection
}

func (m *TasksManager) Check() bool {
	return atomic.LoadInt64(&m.waitCounts) > 0
}
