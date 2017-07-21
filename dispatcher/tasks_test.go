package dispatcher

import (
	"testing"
	"time"

	"github.com/kihamo/go-workers/task"
	"github.com/stretchr/testify/suite"
)

type TasksSuite struct {
	suite.Suite
}

func TestTasksSuite(t *testing.T) {
	suite.Run(t, new(TasksSuite))
}

func (s *TasksSuite) jobNothing(_ int64, _ chan struct{}, _ ...interface{}) (int64, time.Duration, interface{}, error) {
	return 0, 0, nil, nil
}

func (s *TasksSuite) Test_CreateNewInstance_ReturnsZeroLenOfQueue() {
	q := NewTasks()

	s.Zero(q.Len())
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddTask_ReturnsOneLenOfQueue() {
	q := NewTasks()

	q.Add(task.NewTask(s.jobNothing))

	s.Equal(q.Len(), 1)
}

func (s *TasksSuite) Test_CreateNewInstance_ReturnsNilForGetWait() {
	q := NewTasks()

	s.Nil(q.GetWait())
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddWaitTask_ReturnsNotNilForGetWait() {
	q := NewTasks()
	t := task.NewTask(s.jobNothing)
	t.SetStatus(task.TaskStatusWait)

	q.Add(t)

	s.NotNil(q.GetWait())
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddProcessTask_ReturnsNilForGetWait() {
	q := NewTasks()
	t := task.NewTask(s.jobNothing)
	t.SetStatus(task.TaskStatusProcess)

	q.Add(t)

	s.Nil(q.GetWait())
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddSuccessTask_ReturnsNilForGetWait() {
	q := NewTasks()
	t := task.NewTask(s.jobNothing)
	t.SetStatus(task.TaskStatusSuccess)

	q.Add(t)

	s.Nil(q.GetWait())
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddFailTask_ReturnsNilForGetWait() {
	q := NewTasks()
	t := task.NewTask(s.jobNothing)
	t.SetStatus(task.TaskStatusFail)

	q.Add(t)

	s.Nil(q.GetWait())
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddFailByTimeoutTask_ReturnsNilForGetWait() {
	q := NewTasks()
	t := task.NewTask(s.jobNothing)
	t.SetStatus(task.TaskStatusFailByTimeout)

	q.Add(t)

	s.Nil(q.GetWait())
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddKillTask_ReturnsNilForGetWait() {
	q := NewTasks()
	t := task.NewTask(s.jobNothing)
	t.SetStatus(task.TaskStatusKill)

	q.Add(t)

	s.Nil(q.GetWait())
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddRepeatWaitTask_ReturnsNotNilForGetWait() {
	q := NewTasks()
	t := task.NewTask(s.jobNothing)
	t.SetStatus(task.TaskStatusRepeatWait)

	q.Add(t)

	s.NotNil(q.GetWait())
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddThreeTasksWithOnePriority_ReturnsCorrectPriorityList() {
	q := NewTasks()
	t1 := task.NewTask(s.jobNothing)
	t1.SetPriority(1)
	t2 := task.NewTask(s.jobNothing)
	t2.SetPriority(1)
	t3 := task.NewTask(s.jobNothing)
	t3.SetPriority(1)

	q.Add(t1)
	q.Add(t2)
	q.Add(t3)

	s.Equal(q.GetWait().GetId(), t1.GetId())
	s.Equal(q.GetWait().GetId(), t2.GetId())
	s.Equal(q.GetWait().GetId(), t3.GetId())
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddThreeTasksWithDifferentPriority_ReturnsCorrectPriorityList() {
	q := NewTasks()
	t1 := task.NewTask(s.jobNothing)
	t1.SetPriority(3)
	t2 := task.NewTask(s.jobNothing)
	t2.SetPriority(1)
	t3 := task.NewTask(s.jobNothing)
	t3.SetPriority(2)

	q.Add(t1)
	q.Add(t2)
	q.Add(t3)

	s.Equal(q.GetWait().GetId(), t2.GetId())
	s.Equal(q.GetWait().GetId(), t3.GetId())
	s.Equal(q.GetWait().GetId(), t1.GetId())
}

func (s *TasksSuite) Test_CreateNewInstance_ReturnsFalseForHasWait() {
	q := NewTasks()

	s.False(q.HasWait())
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddWaitTask_ReturnsTrueForHasWait() {
	q := NewTasks()

	q.Add(task.NewTask(s.jobNothing))

	s.True(q.HasWait())
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddProcessTask_ReturnsFalseForHasWait() {
	q := NewTasks()
	t := task.NewTask(s.jobNothing)
	t.SetStatus(task.TaskStatusProcess)

	q.Add(t)

	s.False(q.HasWait())
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddSuccessTask_ReturnsFalseForHasWait() {
	q := NewTasks()
	t := task.NewTask(s.jobNothing)
	t.SetStatus(task.TaskStatusSuccess)

	q.Add(t)

	s.False(q.HasWait())
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddFailTask_ReturnsFalseForHasWait() {
	q := NewTasks()
	t := task.NewTask(s.jobNothing)
	t.SetStatus(task.TaskStatusFail)

	q.Add(t)

	s.False(q.HasWait())
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddFailByTimeoutTask_ReturnsFalseForHasWait() {
	q := NewTasks()
	t := task.NewTask(s.jobNothing)
	t.SetStatus(task.TaskStatusFailByTimeout)

	q.Add(t)

	s.False(q.HasWait())
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddKillTask_ReturnsFalseForHasWait() {
	q := NewTasks()
	t := task.NewTask(s.jobNothing)
	t.SetStatus(task.TaskStatusKill)

	q.Add(t)

	s.False(q.HasWait())
}

func (s *TasksSuite) Test_CreateNewInstanceAndAddRepeatWaitTask_ReturnsFalseForHasWait() {
	q := NewTasks()
	t := task.NewTask(s.jobNothing)
	t.SetStatus(task.TaskStatusRepeatWait)

	q.Add(t)

	s.True(q.HasWait())
}

func BenchmarkAdd(b *testing.B) {
	c := NewTasks()

	f := func(_ int64, _ chan struct{}, _ ...interface{}) (int64, time.Duration, interface{}, error) {
		return 0, 0, nil, nil
	}

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Add(task.NewTask(f))
		}
	})
}

func BenchmarkAddAndMix(b *testing.B) {
	c := NewTasks()

	f := func(_ int64, _ chan struct{}, _ ...interface{}) (int64, time.Duration, interface{}, error) {
		return 0, 0, nil, nil
	}

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			t := task.NewTask(f)
			c.Add(t)
			c.Len()
			c.Remove(t)

			c.Add(t)
			c.HasWait()
			c.GetWait()
		}
	})
}
