package workers

type Workers struct {
	Heap
}

func NewWorkers() *Workers {
	heap := &Workers{}
	heap.items = make([]HeapItem, 0)
	heap.positions = map[string]int{}

	return heap
}

func (h *Workers) GetItems() []*Worker {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	heapItems := h.Heap.GetItems()
	items := make([]*Worker, len(heapItems))

	for key, value := range heapItems {
		items[key] = value.(*Worker)
	}

	return items
}
