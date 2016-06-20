package collection

import (
	"sync"
)

type HeapItem interface {
	GetId() string
	GetStatus() int64
}

type Heap struct {
	mutex sync.RWMutex

	items     []HeapItem
	positions map[string]int
}

func (h *Heap) Len() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return len(h.items)
}

func (h *Heap) Less(i, j int) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.items[i].GetStatus() < h.items[j].GetStatus()
}

func (h *Heap) Swap(i, j int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	n := len(h.items)

	if i >= 0 && i < n && j >= 0 && j < n {
		h.items[i], h.items[j] = h.items[j], h.items[i]
	}
}

func (h *Heap) Push(x interface{}) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	item := x.(HeapItem)
	h.positions[item.GetId()] = len(h.items)

	h.items = append(h.items, item)
}

func (h *Heap) Pop() interface{} {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	n := len(h.items)

	item := h.items[n-1]
	delete(h.positions, item.GetId())

	h.items = h.items[0 : n-1]

	return item
}

func (h *Heap) GetItems() []HeapItem {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return h.items
}

func (h *Heap) GetIndexById(i string) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if v, ok := h.positions[i]; ok {
		return v
	}

	return -1
}

func (h *Heap) Remove(p int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if p > len(h.positions)-1 {
		return
	}

	item := h.items[p]

	h.items = append(h.items[:p], h.items[p+1:]...)
	delete(h.positions, item.GetId())
}

func (h *Heap) RemoveById(i string) {
	position := h.GetIndexById(i)
	if position != -1 {
		h.Remove(position)
	}
}
