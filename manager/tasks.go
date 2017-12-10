package manager

import (
	"container/heap"
	"errors"
	"sync/atomic"

	"github.com/kihamo/go-workers"
)

type TasksManager struct {
	queue      *tasksQueue
	waitCounts uint64
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

	if t.Index() < 0 {
		heap.Push(m.queue, t)
	}

	if t.IsWait() {
		atomic.AddUint64(&m.waitCounts, 1)
	}

	return nil
}

func (m *TasksManager) Pull() workers.ManagerItem {
	if m.WaitingCount() < 1 {
		return nil
	}

	var ret workers.ManagerItem
	forReturn := []workers.ManagerItem{}

	for item := heap.Pop(m.queue); item != nil; item = heap.Pop(m.queue) {
		mItem := item.(workers.ManagerItem)
		forReturn = append(forReturn, mItem)

		if item.(*TasksManagerItem).IsWait() {
			mItem.Lock()
			atomic.AddUint64(&m.waitCounts, ^uint64(0))
			ret = mItem

			break
		}
	}

	for _, item := range forReturn {
		heap.Push(m.queue, item)
	}

	return ret
}

func (m *TasksManager) Remove(item workers.ManagerItem) {
	t := item.(*TasksManagerItem)

	i := t.Index()
	if i >= 0 && i < m.queue.Len() {
		heap.Remove(m.queue, i)

		if !t.IsLocked() {
			// TODO: check is exists
			atomic.AddUint64(&m.waitCounts, ^uint64(0))
		} else {
			t.Unlock()
		}
	}
}

func (m *TasksManager) GetById(id string) workers.ManagerItem {
	for _, t := range m.queue.All() {
		if t.Id() == id {
			return t
		}
	}

	return nil
}

func (m *TasksManager) GetAll() []workers.ManagerItem {
	all := m.queue.All()
	collection := make([]workers.ManagerItem, 0, len(all))

	for _, t := range all {
		collection = append(collection, t)
	}

	return collection
}

func (m *TasksManager) WaitingCount() uint64 {
	return atomic.LoadUint64(&m.waitCounts)
}
