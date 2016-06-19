package task

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TaskSuite struct {
	suite.Suite
}

func TestTaskSuite(t *testing.T) {
	suite.Run(t, new(TaskSuite))
}

func (s *TaskSuite) job(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
	return 1, time.Second
}

func (s *TaskSuite) Test_NewInstance_GetIdReturnsId() {
	t := NewTask(s.job)

	assert.IsType(s.T(), "", t.GetId())
	assert.NotEmpty(s.T(), t.GetId())
}

func (s *TaskSuite) Test_NewInstance_GetNameReturnsId() {
	t := NewTask(s.job)

	assert.Equal(s.T(), t.GetName(), t.GetId())
}

func (s *TaskSuite) Test_SetNameTestName_GetNameReturnsTestName() {
	t := NewTask(s.job)

	t.SetName("test-name")

	assert.Equal(s.T(), t.GetName(), "test-name")
}

func (s *TaskSuite) Test_NewInstance_GetDurationReturnsZero() {
	t := NewTask(s.job)

	assert.Equal(s.T(), t.GetDuration(), time.Duration(0))
}

func (s *TaskSuite) Test_SetDurationFiveSecond_GetDurationReturnsFiveSecond() {
	t := NewTask(s.job)

	t.SetDuration(time.Duration(time.Second * 5))

	assert.Equal(s.T(), t.GetDuration(), time.Second*5)
}

func (s *TaskSuite) Test_NewInstance_GetRepeatsReturnsOnce() {
	t := NewTask(s.job)

	assert.Equal(s.T(), t.GetRepeats(), int64(1))
}

func (s *TaskSuite) Test_SetRepeatsTwice_GetRepeatReturnsTwice() {
	t := NewTask(s.job)

	t.SetRepeats(2)

	assert.Equal(s.T(), t.GetRepeats(), int64(2))
}

func (s *TaskSuite) Test_NewInstance_GetAttemptsReturnsZero() {
	t := NewTask(s.job)

	assert.Equal(s.T(), t.GetAttempts(), int64(0))
}

func (s *TaskSuite) Test_SetAttemptsOnce_GetAttemptsReturnsOnce() {
	t := NewTask(s.job)

	t.SetAttempts(1)

	assert.Equal(s.T(), t.GetAttempts(), int64(1))
}

func (s *TaskSuite) Test_NewInstance_GetStatusReturnsStatusWait() {
	t := NewTask(s.job)

	assert.Equal(s.T(), t.GetStatus(), TaskStatusWait)
}

func (s *TaskSuite) Test_SetStatusFail_GetStatusReturnsStatusFail() {
	t := NewTask(s.job)

	t.SetStatus(TaskStatusFail)

	assert.Equal(s.T(), t.GetStatus(), TaskStatusFail)
}

func (s *TaskSuite) Test_NewInstance_GetLastErrorReturnsNil() {
	t := NewTask(s.job)

	assert.Nil(s.T(), t.GetLastError())
}

func (s *TaskSuite) Test_SetLastError_GetLastErrorReturnsError() {
	t := NewTask(s.job)
	err := errors.New("Failed")

	t.SetLastError(err)

	assert.Equal(s.T(), t.GetLastError(), err)
}

func (s *TaskSuite) Test_NewInstance_GetFinishedAtReturnsNil() {
	t := NewTask(s.job)

	assert.Nil(s.T(), t.GetFinishedAt())
}

func (s *TaskSuite) Test_SetFinishedAt_GetFinishedAtReturnsTime() {
	t := NewTask(s.job)
	time := time.Now()

	t.SetFinishedAt(time)

	assert.Equal(s.T(), *t.GetFinishedAt(), time)
}

func (s *TaskSuite) Test_NewInstance_GetTimeoutReturnsZero() {
	t := NewTask(s.job)

	assert.Equal(s.T(), t.GetTimeout(), time.Duration(0))
}

func (s *TaskSuite) Test_SetTimeoutFiveSecond_GetTimeoutReturnsFiveSecond() {
	t := NewTask(s.job)

	t.SetTimeout(time.Duration(time.Second * 5))

	assert.Equal(s.T(), t.GetTimeout(), time.Second*5)
}
