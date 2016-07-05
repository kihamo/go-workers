package worker

import (
	"errors"
	"testing"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
	"github.com/kihamo/go-workers/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type WorkerSuite struct {
	suite.Suite
	clockTime time.Time
}

func TestWorkerSuite(t *testing.T) {
	suite.Run(t, new(WorkerSuite))
}

func (s *WorkerSuite) SetupSuite() {
	s.clockTime = time.Date(2016, 6, 5, 4, 3, 2, 1, time.UTC)
}

func (s *WorkerSuite) getTaskWithSleepJob() task.Tasker {
	c := fakeclock.NewFakeClock(s.clockTime)

	return task.NewTaskWithClock(c, func(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration, error) {
		c.Sleep(time.Second * 6)
		return 1, time.Second, nil
	})
}

func (s *WorkerSuite) job(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration, error) {
	return 1, time.Second, nil
}

func (s *WorkerSuite) jobReturnsPanic(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration, error) {
	panic("Panic!!!")
}

func (s *WorkerSuite) Test_CreateNewInstance_ReturnsStatusIsWait() {
	w := NewWorkman(make(chan Worker))

	assert.Equal(s.T(), w.GetStatus(), WorkerStatusWait)
}

func (s *WorkerSuite) Test_GetId_ReturnsId() {
	w := NewWorkman(make(chan Worker))

	assert.IsType(s.T(), "", w.GetId())
	assert.NotEmpty(s.T(), w.GetId())
}

func (s *WorkerSuite) Test_GetCreateAt_ReturnsCreateAtTime() {
	w := NewWorkmanWithClock(make(chan Worker), fakeclock.NewFakeClock(s.clockTime))

	assert.Equal(s.T(), w.GetCreatedAt(), s.clockTime)
	assert.IsType(s.T(), time.Time{}, w.GetCreatedAt())
}

func (s *WorkerSuite) Test_SetTask_Success() {
	w := NewWorkman(make(chan Worker))

	t := task.NewTask(s.job)
	w.setTask(t)

	assert.Equal(s.T(), w.GetTask(), t)
}

func (s *WorkerSuite) Test_FirstRun_ReturnEmptyError() {
	var err error

	w := NewWorkman(make(chan Worker))

	go func() {
		err = w.Run()
	}()
	for w.GetStatus() != WorkerStatusProcess {
	}

	assert.Nil(s.T(), err)
	w.Kill()
}

func (s *WorkerSuite) Test_TwiceRun_ReturnErrorForSecondRun() {
	w := NewWorkman(make(chan Worker))
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	err := w.Run()

	assert.Equal(s.T(), err, errors.New("Worker is running"))
	w.Kill()
}

func (s *WorkerSuite) Test_IsNotRunning_ReturnStatusIsWait() {
	w := NewWorkman(make(chan Worker))

	assert.Equal(s.T(), w.GetStatus(), WorkerStatusWait)
}

func (s *WorkerSuite) Test_IsRunning_ReturnStatusIsProcess() {
	w := NewWorkman(make(chan Worker))
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	assert.Equal(s.T(), w.GetStatus(), WorkerStatusProcess)
	w.Kill()
}

func (s *WorkerSuite) Test_IsRunningAndSendTask_ReturnStatusIsBusy() {
	w := NewWorkman(make(chan Worker))
	t := s.getTaskWithSleepJob()
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	w.SendTask(t)
	for w.GetStatus() != WorkerStatusBusy {
	}

	assert.Equal(s.T(), w.GetStatus(), WorkerStatusBusy)
	w.Kill()
}

func (s *WorkerSuite) Test_Reset_ReturnEmptyTask() {
	w := NewWorkman(make(chan Worker))
	t := task.NewTask(s.job)
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	w.SendTask(t)
	for w.GetStatus() != WorkerStatusWait {
	}

	assert.NotNil(s.T(), w.GetTask())
	w.Reset()
	assert.Nil(s.T(), w.GetTask())

	w.Kill()
}

func (s *WorkerSuite) Test_IsRunningAndKill_ReturnEmptyError() {
	w := NewWorkman(make(chan Worker))
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	err := w.Kill()

	assert.Nil(s.T(), err)
}

func (s *WorkerSuite) Test_CreateNewInstanceAndKill_ReturnError() {
	w := NewWorkman(make(chan Worker))
	err := w.Kill()

	assert.Equal(s.T(), err, errors.New("Worker isn't running"))
}

func (s *WorkerSuite) Test_IsRunningTwiceKill_ReturnErrorForSecondKill() {
	w := NewWorkman(make(chan Worker))
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}
	w.Kill()
	for w.GetStatus() != WorkerStatusWait {
	}

	err := w.Kill()

	assert.Equal(s.T(), err, errors.New("Worker isn't running"))
}

func (s *WorkerSuite) Test_Kill_ReturnStatusIsWait() {
	w := NewWorkman(make(chan Worker))
	t := s.getTaskWithSleepJob()
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}
	w.SendTask(t)
	for w.GetStatus() != WorkerStatusBusy {
	}
	w.Kill()
	for w.GetStatus() != WorkerStatusWait {
	}

	assert.Equal(s.T(), w.GetStatus(), WorkerStatusWait)
}

func (s *WorkerSuite) Test_WithTaskReturnsPanic_SetTaskStatusIsFail() {
	w := NewWorkman(make(chan Worker))
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	t := task.NewTask(s.jobReturnsPanic)
	w.SendTask(t)
	for w.GetStatus() != WorkerStatusWait {
	}

	assert.Equal(s.T(), t.GetStatus(), task.TaskStatusFail)

	w.Kill()
}

func (s *WorkerSuite) Test_WithTaskReturnsPanic_SetLastErrorIsString() {
	w := NewWorkman(make(chan Worker))
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	t := task.NewTask(s.jobReturnsPanic)
	w.SendTask(t)
	for w.GetStatus() != WorkerStatusWait {
	}

	assert.Equal(s.T(), t.GetLastError(), "Panic!!!")

	w.Kill()
}

func (s *WorkerSuite) Test_WithTaskWithTimeoutAndReturnsPanic_SetTaskStatusIsFail() {
	w := NewWorkman(make(chan Worker))
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	t := task.NewTask(s.jobReturnsPanic)
	t.SetTimeout(time.Second * 10)
	w.SendTask(t)
	for w.GetStatus() != WorkerStatusWait {
	}

	assert.Equal(s.T(), t.GetStatus(), task.TaskStatusFail)

	w.Kill()
}

func (s *WorkerSuite) Test_WithTask_SetTaskStatusIsKillIfWorkerKill() {
	w := NewWorkman(make(chan Worker))
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	t := s.getTaskWithSleepJob()
	w.SendTask(t)
	for w.GetStatus() != WorkerStatusBusy {
	}

	w.Kill()
	for w.GetStatus() != WorkerStatusWait {
	}

	assert.Equal(s.T(), t.GetStatus(), task.TaskStatusKill)
}

func (s *WorkerSuite) Test_WithTaskWithTimeout_SetTaskStatusIsKillIfWorkerKill() {
	w := NewWorkman(make(chan Worker))
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	t := s.getTaskWithSleepJob()
	t.SetTimeout(time.Second * 10)
	w.SendTask(t)
	for w.GetStatus() != WorkerStatusBusy {
	}

	w.Kill()
	for w.GetStatus() != WorkerStatusWait {
	}

	assert.Equal(s.T(), t.GetStatus(), task.TaskStatusKill)
}

func (s *WorkerSuite) Test_WithTaskWithTimeout_SetTaskStatusIsFailTimeout() {
	w := NewWorkmanWithClock(make(chan Worker), fakeclock.NewFakeClock(s.clockTime))
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	t := s.getTaskWithSleepJob()
	t.SetTimeout(time.Second)
	w.SendTask(t)

	w.GetClock().(*fakeclock.FakeClock).WaitForWatcherAndIncrement(time.Second)

	for w.GetStatus() != WorkerStatusWait {
	}

	assert.Equal(s.T(), t.GetStatus(), task.TaskStatusFailByTimeout)

	w.Kill()
}

func (s *WorkerSuite) Test_WithTask_SetTaskStatusIsSuccess() {
	w := NewWorkman(make(chan Worker))
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	t := task.NewTask(s.job)
	w.SendTask(t)
	for w.GetStatus() != WorkerStatusWait {
	}

	assert.Equal(s.T(), t.GetStatus(), task.TaskStatusSuccess)

	w.Kill()
}

func (s *WorkerSuite) Test_WithTaskWithTimeout_SetTaskStatusIsSuccess() {
	w := NewWorkmanWithClock(make(chan Worker), fakeclock.NewFakeClock(s.clockTime))
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	c := fakeclock.NewFakeClock(s.clockTime)
	t := task.NewTaskWithClock(c, func(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration, error) {
		c.Sleep(time.Second * 6)
		return 1, time.Second, nil
	})
	t.SetTimeout(time.Second * 10)
	w.SendTask(t)
	for w.GetStatus() != WorkerStatusBusy {
	}

	c.WaitForWatcherAndIncrement(time.Second * 6)

	for t.GetStatus() != task.TaskStatusSuccess {
	}

	assert.Equal(s.T(), t.GetStatus(), task.TaskStatusSuccess)

	w.Kill()
}

func (s *WorkerSuite) Test_IsRunningAndAddTask_SetNowForTaskStartedAtAfterTaskProcess() {
	w := NewWorkmanWithClock(make(chan Worker), fakeclock.NewFakeClock(s.clockTime))
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	t := s.getTaskWithSleepJob()
	w.SendTask(t)
	for t.GetStatus() != task.TaskStatusProcess {
	}

	assert.Equal(s.T(), t.GetStartedAt().String(), w.GetClock().Now().String())

	w.Kill()
}

func (s *WorkerSuite) Test_IsRunningAndAddTask_NotSetStartedAtBeforeTaskProcess() {
	w := NewWorkmanWithClock(make(chan Worker), fakeclock.NewFakeClock(s.clockTime))
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	t := s.getTaskWithSleepJob()
	w.SendTask(t)

	assert.Nil(s.T(), t.GetStartedAt())

	w.Kill()
}

func (s *WorkerSuite) Test_IsRunningAndAddTaskWithNotEmptyLastError_SetNilLastError() {
	w := NewWorkman(make(chan Worker))
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	t := s.getTaskWithSleepJob()
	t.SetLastError(errors.New("Panic"))
	w.SendTask(t)
	for t.GetStatus() != task.TaskStatusProcess {
	}

	assert.Nil(s.T(), t.GetLastError())

	w.Kill()
}

func (s *WorkerSuite) Test_IsRunningAndAddTaskWithTwoAttempts_SetOneAttempts() {
	w := NewWorkman(make(chan Worker))
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	t := s.getTaskWithSleepJob()
	t.SetAttempts(2)
	w.SendTask(t)
	for t.GetStatus() != task.TaskStatusProcess {
	}

	assert.Equal(s.T(), t.GetAttempts(), int64(1))

	w.Kill()
}

func (s *WorkerSuite) Test_IsRunningAndAddTaskWithWaitStatusAndTwoAttempts_SetThreeAttempts() {
	w := NewWorkman(make(chan Worker))
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	t := s.getTaskWithSleepJob()
	t.SetAttempts(2)
	t.SetStatus(task.TaskStatusRepeatWait)
	w.SendTask(t)
	for t.GetStatus() != task.TaskStatusProcess {
	}

	assert.Equal(s.T(), t.GetAttempts(), int64(3))

	w.Kill()
}

func (s *WorkerSuite) Test_IsRunningAndAddTask_SetFinishedAt() {
	w := NewWorkmanWithClock(make(chan Worker), fakeclock.NewFakeClock(s.clockTime))
	go w.Run()
	for w.GetStatus() != WorkerStatusProcess {
	}

	finishedAt := w.GetClock().Now().Add(time.Second * 6)

	c := fakeclock.NewFakeClock(s.clockTime)
	t := task.NewTaskWithClock(c, func(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration, error) {
		c.Sleep(time.Second * 6)
		return 1, time.Second, nil
	})
	w.SendTask(t)
	for t.GetStatus() != task.TaskStatusProcess {
	}

	w.GetClock().(*fakeclock.FakeClock).IncrementBySeconds(6)
	c.WaitForWatcherAndIncrement(time.Second * 6)

	for t.GetStatus() != task.TaskStatusSuccess {
	}

	assert.Equal(s.T(), t.GetFinishedAt().String(), finishedAt.String())

	w.Kill()
}
