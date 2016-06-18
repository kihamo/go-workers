package workers

type Tasks struct {
	Heap
}

func NewTasks() *Tasks {
	heap := &Tasks{}
	heap.items = make([]HeapItem, 0)
	heap.positions = map[string]int{}

	return heap
}

func (h *Tasks) GetItems() []*Task {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	heapItems := h.Heap.GetItems()
	items := make([]*Task, len(heapItems))

	for key, value := range heapItems {
		items[key] = value.(*Task)
	}

	return items
}
