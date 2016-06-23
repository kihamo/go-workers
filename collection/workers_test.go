package collection

import (
	"container/heap"
	"testing"

	"github.com/kihamo/go-workers/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type WorkersSuite struct {
	suite.Suite

	collection *Workers
}

func TestWorkersSuite(t *testing.T) {
	suite.Run(t, new(WorkersSuite))
}

func (s *WorkersSuite) SetupTest() {
	s.collection = NewWorkers()
}

func (s *WorkersSuite) Test_CreateNewInstance_LenReturnsZero() {
	assert.Equal(s.T(), s.collection.Len(), 0)
}

func (s *WorkersSuite) Test_CreateNewHeap_LenReturnsZero() {
	heap.Init(s.collection)

	assert.Equal(s.T(), s.collection.Len(), 0)
}

func (s *WorkersSuite) Test_CreateNewInstanceAndAddOneItem_LenReturnsOne() {
	s.collection.Push(worker.NewWorkman(make(chan worker.Worker)))

	assert.Equal(s.T(), s.collection.Len(), 1)
}

func (s *WorkersSuite) Test_CreateNewHeapAndAddOneItem_LenReturnsOne() {
	heap.Init(s.collection)
	heap.Push(s.collection, worker.NewWorkman(make(chan worker.Worker)))

	assert.Equal(s.T(), s.collection.Len(), 1)
}

func (s *WorkersSuite) Test_CreateNewInstance_GetItemsReturnsEmptyList() {
	assert.Len(s.T(), s.collection.GetItems(), 0)
}

func (s *WorkersSuite) Test_CreateNewHeap_GetItemsReturnsEmptyList() {
	heap.Init(s.collection)

	assert.Len(s.T(), s.collection.GetItems(), 0)
}

func (s *WorkersSuite) Test_CreateNewInstanceAndAddOneItem_GetItemsReturnsOneSizeOfList() {
	s.collection.Push(worker.NewWorkman(make(chan worker.Worker)))

	assert.Len(s.T(), s.collection.GetItems(), 1)
}

func (s *WorkersSuite) Test_CreateNewHeapAndAddOneItem_GetItemsReturnsOneSizeOfList() {
	heap.Init(s.collection)
	heap.Push(s.collection, worker.NewWorkman(make(chan worker.Worker)))

	assert.Len(s.T(), s.collection.GetItems(), 1)
}

func (s *WorkersSuite) Test_CreateNewInstanceAndAddOneItem_PopReturnsTasker() {
	s.collection.Push(worker.NewWorkman(make(chan worker.Worker)))

	assert.IsType(s.T(), &worker.Workman{}, s.collection.Pop())
}

func (s *WorkersSuite) Test_CreateNewHeapAndAddOneItem_PopReturnsTasker() {
	heap.Init(s.collection)
	heap.Push(s.collection, worker.NewWorkman(make(chan worker.Worker)))

	assert.IsType(s.T(), &worker.Workman{}, heap.Pop(s.collection))
}
