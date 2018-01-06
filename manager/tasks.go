package manager

import (
	"container/heap"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/kihamo/go-workers"
)

type TasksManager struct {
	mutex          sync.Mutex
	unlockedCounts uint64
	queue          *tasksQueue
}

func NewTasksManager() *TasksManager {
	return &TasksManager{
		queue:          newTasksQueue(),
		unlockedCounts: 0,
	}
}

func (m *TasksManager) Push(task workers.ManagerItem) error {
	if task == nil {
		return errors.New("TasksManagerItem can't be nil")
	}

	task.Unlock()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	t := task.(*TasksManagerItem)

	if t.Index() < 0 {
		heap.Push(m.queue, t)
	}

	atomic.AddUint64(&m.unlockedCounts, 1)

	return nil
}

func (m *TasksManager) Pull() workers.ManagerItem {
	if m.UnlockedCount() < 1 {
		return nil
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	var ret workers.ManagerItem
	forReturn := []workers.ManagerItem{}

	for item := heap.Pop(m.queue); item != nil; item = heap.Pop(m.queue) {
		mItem := item.(workers.ManagerItem)
		forReturn = append(forReturn, mItem)

		if !mItem.IsLocked() {
			mItem.Lock()
			atomic.AddUint64(&m.unlockedCounts, ^uint64(0))
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
	m.mutex.Lock()
	defer m.mutex.Unlock()

	t := item.(*TasksManagerItem)

	i := t.Index()
	if i >= 0 && i < m.queue.Len() {
		heap.Remove(m.queue, i)

		if !t.IsLocked() {
			// TODO: check is exists
			atomic.AddUint64(&m.unlockedCounts, ^uint64(0))
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

func (m *TasksManager) UnlockedCount() uint64 {
	return atomic.LoadUint64(&m.unlockedCounts)
}
