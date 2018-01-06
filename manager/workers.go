package manager

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/kihamo/go-workers"
)

type WorkersManager struct {
	mutex          sync.RWMutex
	unlockedCounts uint64
	queue          []workers.ManagerItem
	workers        map[string]workers.ManagerItem
}

func NewWorkersManager() *WorkersManager {
	c := &WorkersManager{
		queue:          []workers.ManagerItem{},
		workers:        map[string]workers.ManagerItem{},
		unlockedCounts: 0,
	}

	return c
}

func (m *WorkersManager) Push(worker workers.ManagerItem) error {
	if worker == nil {
		return errors.New("Worker can't be nil")
	}

	m.mutex.RLock()
	exists, ok := m.workers[worker.Id()]
	m.mutex.RUnlock()

	if ok && exists != worker {
		return fmt.Errorf("Worker with ID %s already exists", worker.Id())
	}

	worker.Unlock()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !ok {
		m.workers[worker.Id()] = worker
	}

	m.queue = append(m.queue, worker)
	atomic.AddUint64(&m.unlockedCounts, 1)

	return nil
}

func (m *WorkersManager) Pull() (worker workers.ManagerItem) {
	if m.UnlockedCount() < 1 {
		return nil
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	for len(m.queue) > 0 {
		worker, m.queue = m.queue[0], m.queue[1:]

		if _, ok := m.workers[worker.Id()]; !ok || worker.IsLocked() {
			continue
		}

		worker.Lock()
		atomic.AddUint64(&m.unlockedCounts, ^uint64(0))

		return worker
	}

	return nil
}

func (m *WorkersManager) Remove(item workers.ManagerItem) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if w, ok := m.workers[item.Id()]; ok {
		delete(m.workers, item.Id())

		if !w.IsLocked() {
			atomic.AddUint64(&m.unlockedCounts, ^uint64(0))
		} else {
			w.Unlock()
		}
	}
}

func (m *WorkersManager) GetById(id string) workers.ManagerItem {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if w, ok := m.workers[id]; ok {
		return w
	}

	return nil
}

func (m *WorkersManager) GetAll() []workers.ManagerItem {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	collection := make([]workers.ManagerItem, 0, len(m.workers))

	for _, w := range m.workers {
		collection = append(collection, w)
	}

	return collection
}

func (m *WorkersManager) UnlockedCount() uint64 {
	return atomic.LoadUint64(&m.unlockedCounts)
}
