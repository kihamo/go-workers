package collection

import (
	"github.com/kihamo/go-workers/task"
)

type Tasks struct {
	Heap
}

func NewTasks() *Tasks {
	heap := &Tasks{}
	heap.items = make([]HeapItem, 0)
	heap.positions = map[string]int{}

	return heap
}

func (c *Tasks) GetItems() []task.Tasker {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	heapItems := c.Heap.GetItems()
	items := make([]task.Tasker, len(heapItems))

	for key, value := range heapItems {
		items[key] = value.(task.Tasker)
	}

	return items
}
