package manager

import (
	"sync"
)

type tasksQueue struct {
	mutex sync.RWMutex
	list  []*TasksManagerItem
}

func newTasksQueue() *tasksQueue {
	return &tasksQueue{
		list: make([]*TasksManagerItem, 0),
	}
}

func (q *tasksQueue) Len() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	return len(q.list)
}

func (q *tasksQueue) Less(i, j int) bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	if q.list[i].IsWait() != q.list[j].IsWait() {
		return q.list[i].IsWait()
	}

	if q.list[i].Task().Priority() == q.list[j].Task().Priority() {
		return q.list[i].AllowStartAt().Before(q.list[j].AllowStartAt())
	}

	return q.list[i].Task().Priority() < q.list[j].Task().Priority()
}

func (q *tasksQueue) Swap(i, j int) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	n := len(q.list)

	if i >= 0 && i < n && j >= 0 && j < n {
		q.list[i], q.list[j] = q.list[j], q.list[i]
		q.list[i].setIndex(i)
		q.list[j].setIndex(j)
	}
}

func (q *tasksQueue) Push(x interface{}) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	item := x.(*TasksManagerItem)
	item.setIndex(len(q.list))

	q.list = append(q.list, item)
}

func (q *tasksQueue) Pop() interface{} {
	if q.Len() == 0 {
		return nil
	}

	q.mutex.Lock()
	defer q.mutex.Unlock()

	old := q.list
	n := len(old) - 1

	item := old[n]
	item.setIndex(-1)

	q.list = old[:n]

	return item
}

func (q *tasksQueue) All() []*TasksManagerItem {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	tmp := make([]*TasksManagerItem, len(q.list))
	copy(tmp, q.list)

	return tmp
}
