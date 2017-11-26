package manager

import (
	"testing"
	"time"

	"github.com/kihamo/go-workers"
	"github.com/kihamo/go-workers/task"
	"github.com/stretchr/testify/suite"
)

type TasksQueueSuite struct {
	suite.Suite
}

func TestTasksQueueSuite(t *testing.T) {
	suite.Run(t, new(TasksQueueSuite))
}

func (s *TasksQueueSuite) Test_CreateNewInstance_ReturnsZeroLenOfQueue() {
	q := newTasksQueue()

	s.Zero(q.Len())
}

func (s *TasksQueueSuite) Test_CreateNewInstanceAndPushOnce_ReturnsOneLenOfQueue() {
	q := newTasksQueue()

	q.Push(&TasksManagerItem{})

	s.Equal(q.Len(), 1)
}

func (s *TasksQueueSuite) Test_CreateNewInstanceAndPushOnceAndPopOnce_ReturnsZeroLenOfQueue() {
	q := newTasksQueue()
	q.Push(&TasksManagerItem{})

	q.Pop()

	s.Zero(q.Len())
}

func (s *TasksQueueSuite) Test_CreateNewInstanceAndPopOnce_ReturnsZeroLenOfQueue() {
	q := newTasksQueue()

	q.Pop()

	s.Zero(q.Len())
}

func (s *TasksQueueSuite) Test_CreateNewInstanceAndPopOnce_ReturnsNil() {
	q := newTasksQueue()

	t := q.Pop()

	s.Nil(t)
}

func (s *TasksQueueSuite) Test_CreateNewInstanceAndPushOnce_ReturnsNotNil() {
	q := newTasksQueue()
	q.Push(&TasksManagerItem{})

	t := q.Pop()

	s.NotNil(t)
}

func (s *TasksQueueSuite) Test_CreateNewInstanceAndPushOnce_ReturnsEqualItem() {
	q := newTasksQueue()
	source := &TasksManagerItem{}
	q.Push(source)

	target := q.Pop()

	s.Equal(target, source)
}

func (s *TasksQueueSuite) Test_LessTwoTasksWithEqualPriority_ReturnsTrue() {
	q := newTasksQueue()
	t1 := task.NewFunctionTask(func() (interface{}, error) { return nil, nil })
	t1.SetPriority(1)
	i1 := NewTasksManagerItem(t1, workers.TaskStatusWait)
	t2 := task.NewFunctionTask(func() (interface{}, error) { return nil, nil })
	t2.SetPriority(1)
	i2 := NewTasksManagerItem(t2, workers.TaskStatusWait)
	q.Push(i1)
	q.Push(i2)

	result := q.Less(0, 1)

	s.True(result)
}

func (s *TasksQueueSuite) Test_LessTwoTasksWithEqualPriorityAndEqualAllowStartAt_ReturnsFalse() {
	now := time.Now()
	q := newTasksQueue()
	t1 := task.NewFunctionTask(func() (interface{}, error) { return nil, nil })
	t1.SetPriority(1)
	i1 := NewTasksManagerItem(t1, workers.TaskStatusWait)
	i1.allowStartAt = now
	t2 := task.NewFunctionTask(func() (interface{}, error) { return nil, nil })
	t2.SetPriority(1)
	i2 := NewTasksManagerItem(t2, workers.TaskStatusWait)
	i2.allowStartAt = now
	q.Push(i1)
	q.Push(i2)

	result := q.Less(0, 1)

	s.False(result)
}

func (s *TasksQueueSuite) Test_LessTwoTasksWithEqualPriorityAndGtAllowStartAt_ReturnsTrue() {
	now := time.Now()
	q := newTasksQueue()
	t1 := task.NewFunctionTask(func() (interface{}, error) { return nil, nil })
	t1.SetPriority(1)
	i1 := NewTasksManagerItem(t1, workers.TaskStatusWait)
	i1.allowStartAt = now.Add(time.Second)
	t2 := task.NewFunctionTask(func() (interface{}, error) { return nil, nil })
	t2.SetPriority(1)
	i2 := NewTasksManagerItem(t2, workers.TaskStatusWait)
	i2.allowStartAt = now
	q.Push(i1)
	q.Push(i2)

	result := q.Less(0, 1)

	s.False(result)
}

func (s *TasksQueueSuite) Test_LessTwoTasksWithEqualPriorityAndLtAllowStartAt_ReturnsTrue() {
	now := time.Now()
	q := newTasksQueue()
	t1 := task.NewFunctionTask(func() (interface{}, error) { return nil, nil })
	t1.SetPriority(1)
	i1 := NewTasksManagerItem(t1, workers.TaskStatusWait)
	i1.allowStartAt = now
	t2 := task.NewFunctionTask(func() (interface{}, error) { return nil, nil })
	t2.SetPriority(1)
	i2 := NewTasksManagerItem(t2, workers.TaskStatusWait)
	i2.allowStartAt = now.Add(time.Second)
	q.Push(i1)
	q.Push(i2)

	result := q.Less(0, 1)

	s.True(result)
}

func (s *TasksQueueSuite) Test_LessTwoTasksWithLtPriority_ReturnsTrue() {
	q := newTasksQueue()
	t1 := task.NewFunctionTask(func() (interface{}, error) { return nil, nil })
	t1.SetPriority(1)
	i1 := NewTasksManagerItem(t1, workers.TaskStatusWait)
	t2 := task.NewFunctionTask(func() (interface{}, error) { return nil, nil })
	t2.SetPriority(2)
	i2 := NewTasksManagerItem(t2, workers.TaskStatusWait)
	q.Push(i1)
	q.Push(i2)

	result := q.Less(0, 1)

	s.True(result)
}

func (s *TasksQueueSuite) Test_LessTwoTasksWithGtPriority_ReturnsFalse() {
	q := newTasksQueue()
	t1 := task.NewFunctionTask(func() (interface{}, error) { return nil, nil })
	t1.SetPriority(2)
	i1 := NewTasksManagerItem(t1, workers.TaskStatusWait)
	t2 := task.NewFunctionTask(func() (interface{}, error) { return nil, nil })
	t2.SetPriority(1)
	i2 := NewTasksManagerItem(t2, workers.TaskStatusWait)
	q.Push(i1)
	q.Push(i2)

	result := q.Less(0, 1)

	s.False(result)
}

func (s *TasksQueueSuite) Test_SwapTwoTasks_ReturnsSwapped() {
	q := newTasksQueue()
	t1 := task.NewFunctionTask(func() (interface{}, error) { return nil, nil })
	i1 := NewTasksManagerItem(t1, workers.TaskStatusWait)
	t2 := task.NewFunctionTask(func() (interface{}, error) { return nil, nil })
	i2 := NewTasksManagerItem(t2, workers.TaskStatusWait)
	q.Push(i1)
	q.Push(i2)

	q.Swap(0, 1)
	target1 := q.Pop()
	target2 := q.Pop()

	s.Equal(target1, i1)
	s.Equal(target2, i2)
}
