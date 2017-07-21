package dispatcher

import (
	"container/heap"
	"sync"
	"sync/atomic"

	"github.com/kihamo/go-workers/task"
)

const hashLength = 36

type tasksQueue []tasksQueueElement

type tasksQueueElement struct {
	index      int
	sort       int
	priority   int64
	lastStatus int64
	hash       [hashLength]byte
}

type tasksItem struct {
	t task.Tasker
	e *tasksQueueElement
}

func (c tasksQueue) Len() int {
	return len(c)
}

func (c tasksQueue) Less(i, j int) bool {
	if c[i].priority == c[j].priority {
		return c[i].sort < c[j].sort
	}

	return c[i].priority < c[j].priority
}

func (c tasksQueue) Swap(i, j int) {
	n := len(c)

	if i >= 0 && i < n && j >= 0 && j < n {
		c[i], c[j] = c[j], c[i]
		c[i].index, c[j].index = i, j
	}
}

func (c *tasksQueue) Push(x interface{}) {
	element := x.(tasksQueueElement)
	element.index = len(*c)
	element.sort = element.index

	*c = append(*c, element)
}

func (c *tasksQueue) Pop() interface{} {
	old := *c
	n := len(old) - 1

	element := old[n]
	element.index = -1

	*c = old[:n]
	return element
}

type Tasks struct {
	mutex       sync.RWMutex
	queue       *tasksQueue
	itemsByHash map[[hashLength]byte]tasksItem
	length      int64
	wait        int64
}

func NewTasks() *Tasks {
	return &Tasks{
		queue:       &tasksQueue{},
		itemsByHash: make(map[[hashLength]byte]tasksItem, 0),
		length:      0,
		wait:        0,
	}
}

func (q *Tasks) Len() int {
	return int(atomic.LoadInt64(&q.length))
}

func (q *Tasks) Add(t task.Tasker) {
	element := tasksQueueElement{
		priority:   t.GetPriority(),
		lastStatus: t.GetStatus(),
		hash:       q.getHash(t.GetId()),
	}

	item := tasksItem{
		t: t,
		e: &element,
	}

	q.mutex.Lock()
	q.itemsByHash[element.hash] = item
	heap.Push(q.queue, element)
	q.mutex.Unlock()

	if element.lastStatus == task.TaskStatusWait || element.lastStatus == task.TaskStatusRepeatWait {
		atomic.AddInt64(&q.wait, 1)
	}

	atomic.AddInt64(&q.length, 1)
}

func (q *Tasks) GetWait() (t task.Tasker) {
	if !q.HasWait() {
		return nil
	}

	q.mutex.Lock()
	defer q.mutex.Unlock()

	i := 0
	for e := heap.Pop(q.queue); i < len(q.itemsByHash); i++ {
		element := e.(tasksQueueElement)

		if element.lastStatus != task.TaskStatusWait && element.lastStatus != task.TaskStatusRepeatWait {
			heap.Push(q.queue, e)
			continue
		}

		atomic.AddInt64(&q.wait, -1)
		atomic.AddInt64(&q.length, -1)

		t = q.itemsByHash[element.hash].t
		delete(q.itemsByHash, element.hash)

		return t
	}

	return nil
}

func (q *Tasks) Remove(t task.Tasker) {
	q.RemoveById(t.GetId())
}

func (q *Tasks) RemoveById(id string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	hash := q.getHash(id)
	if item, ok := q.itemsByHash[hash]; ok {
		element := heap.Remove(q.queue, item.e.index).(tasksQueueElement)

		if element.lastStatus == task.TaskStatusWait || element.lastStatus == task.TaskStatusRepeatWait {
			atomic.AddInt64(&q.wait, -1)
		}

		atomic.AddInt64(&q.length, -1)

		delete(q.itemsByHash, hash)
	}
}

func (q *Tasks) HasWait() bool {
	return atomic.LoadInt64(&q.wait) > 0
}

func (q *Tasks) GetItems() []task.Tasker {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	items := make([]task.Tasker, 0, len(q.itemsByHash))
	for _, t := range q.itemsByHash {
		items = append(items, t.t)
	}

	return items
}

func (q *Tasks) getHash(id string) (hash [hashLength]byte) {
	copy(hash[:], id)
	return hash
}
