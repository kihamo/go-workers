package manager

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/kihamo/go-workers"
)

type WorkersManager struct {
	mutex      sync.RWMutex
	queue      []workers.ManagerItem
	workers    map[string]workers.ManagerItem
	freeCounts uint64
}

func NewWorkersManager() *WorkersManager {
	c := &WorkersManager{
		queue:      []workers.ManagerItem{},
		workers:    map[string]workers.ManagerItem{},
		freeCounts: 0,
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

	if !ok {
		m.workers[worker.Id()] = worker
	}

	m.queue = append(m.queue, worker)
	atomic.AddUint64(&m.freeCounts, 1)

	m.mutex.Unlock()
	return nil
}

func (m *WorkersManager) Pull() (worker workers.ManagerItem) {
	if m.WaitingCount() < 1 {
		return worker
	}

	m.mutex.Lock()
	forReturn := []workers.ManagerItem{}

	defer func() {
		for _, w := range forReturn {
			m.queue = append(m.queue, w)
		}

		m.mutex.Unlock()
	}()

	for len(m.queue) > 0 {
		worker, m.queue = m.queue[0], m.queue[1:]

		if _, ok := m.workers[worker.Id()]; ok {
			if !worker.IsLocked() {
				worker.Lock()
				atomic.AddUint64(&m.freeCounts, ^uint64(0))

				return worker
			}

			forReturn = append(forReturn, worker)
		}
	}

	return nil
}

func (m *WorkersManager) Remove(item workers.ManagerItem) {
	m.mutex.Lock()

	if w, ok := m.workers[item.Id()]; ok {
		delete(m.workers, item.Id())

		if !w.IsLocked() {
			atomic.AddUint64(&m.freeCounts, ^uint64(0))
		} else {
			w.Unlock()
		}
	}

	m.mutex.Unlock()
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

func (m *WorkersManager) WaitingCount() uint64 {
	return atomic.LoadUint64(&m.freeCounts)
}
