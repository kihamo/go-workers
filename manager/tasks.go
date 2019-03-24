package manager

import (
	"container/heap"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kihamo/go-workers"
)

type TasksManager struct {
	mutex             sync.Mutex
	unlockedCounts    uint64
	queue             *tasksQueue
	tickerRecalculate *workers.Ticker
}

func NewTasksManager() *TasksManager {
	m := &TasksManager{
		queue:             newTasksQueue(),
		unlockedCounts:    0,
		tickerRecalculate: workers.NewTicker(time.Second),
	}

	// TODO: останавливать рутину после остановки диспетчера
	go m.recalculate()

	return m
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

	if !t.IsLocked() {
		atomic.AddUint64(&m.unlockedCounts, 1)
	}

	return nil
}

func (m *TasksManager) Pull() workers.ManagerItem {
	if atomic.LoadUint64(&m.unlockedCounts) < 1 {
		return nil
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	var ret workers.ManagerItem
	forReturn := make([]workers.ManagerItem, 0, m.queue.Len())

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

// пересчитывает количество не заблокированных задач, так как оно меняется произвольно из-за отложенной даты запуска
func (m *TasksManager) recalculate() {
	for {
		<-m.tickerRecalculate.C()
		var unlockedCounts uint64

		for _, t := range m.queue.All() {
			if !t.IsLocked() {
				unlockedCounts++
			}
		}

		atomic.StoreUint64(&m.unlockedCounts, unlockedCounts)
	}
}
