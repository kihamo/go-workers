package collection

import (
	"container/heap"
	"testing"
	"time"

	"github.com/kihamo/go-workers/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TasksSuite struct {
	suite.Suite

	collection *Tasks
}

func (s *TasksSuite) SetupTest() {
	s.collection = NewTasks()
}

func TestTasksSuite(t *testing.T) {
	suite.Run(t, new(TasksSuite))
}

func (s *TasksSuite) job(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
	return 1, time.Second
}

func (s *TasksSuite) Test_CreateNewInstance_LenReturnsZero() {
	assert.Equal(s.T(), s.collection.Len(), 0)
}

func (s *TasksSuite) Test_CreateNewHeap_LenReturnsZero() {
	heap.Init(s.collection)

	assert.Equal(s.T(), s.collection.Len(), 0)
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddOneItem_LenReturnsOne() {
	s.collection.Push(task.NewTask(s.job))

	assert.Equal(s.T(), s.collection.Len(), 1)
}

func (s *TasksSuite) Test_CreateNewHeapAndAddOneItem_LenReturnsOne() {
	heap.Init(s.collection)
	heap.Push(s.collection, task.NewTask(s.job))

	assert.Equal(s.T(), s.collection.Len(), 1)
}

func (s *TasksSuite) Test_CreateNewInstance_GetItemsReturnsEmptyList() {
	assert.Len(s.T(), s.collection.GetItems(), 0)
}

func (s *TasksSuite) Test_CreateNewHeap_GetItemsReturnsEmptyList() {
	heap.Init(s.collection)

	assert.Len(s.T(), s.collection.GetItems(), 0)
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddOneItem_GetItemsReturnsOneSizeOfList() {
	s.collection.Push(task.NewTask(s.job))

	assert.Len(s.T(), s.collection.GetItems(), 1)
}

func (s *TasksSuite) Test_CreateNewHeapAndAddOneItem_GetItemsReturnsOneSizeOfList() {
	heap.Init(s.collection)
	heap.Push(s.collection, task.NewTask(s.job))

	assert.Len(s.T(), s.collection.GetItems(), 1)
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddOneItem_PopReturnsTasker() {
	s.collection.Push(task.NewTask(s.job))

	assert.IsType(s.T(), &task.Task{}, s.collection.Pop())
}

func (s *TasksSuite) Test_CreateNewHeapAndAddOneItem_PopReturnsTasker() {
	heap.Init(s.collection)
	heap.Push(s.collection, task.NewTask(s.job))

	assert.IsType(s.T(), &task.Task{}, heap.Pop(s.collection))
}
