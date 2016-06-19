package collection

import (
	"github.com/kihamo/go-workers/worker"
)

type Workers struct {
	Heap
}

func NewWorkers() *Workers {
	heap := &Workers{}
	heap.items = make([]HeapItem, 0)
	heap.positions = map[string]int{}

	return heap
}

func (c *Workers) GetItems() []worker.Worker {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	heapItems := c.Heap.GetItems()
	items := make([]worker.Worker, len(heapItems))

	for key, value := range heapItems {
		items[key] = value.(worker.Worker)
	}

	return items
}
