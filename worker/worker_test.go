package worker

import (
	"errors"
	"testing"
	"time"

	"github.com/kihamo/go-workers"
	"github.com/kihamo/go-workers/task"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type WorkerSuite struct {
	suite.Suite

	clock  time.Time
	worker *Workman
	done   chan Worker
}

func (s *WorkerSuite) SetupTest() {
	s.clock = time.Date(2016, 6, 5, 4, 3, 2, 1, time.UTC)
	workers.Clock = fakeclock.NewFakeClock(s.clock)

	s.done = make(chan Worker)
	s.worker = NewWorker(s.done)
}

func (s *WorkerSuite) TearDownTest() {
	s.worker.Kill()
}

func TestWorkerSuite(t *testing.T) {
	suite.Run(t, new(WorkerSuite))
}

func (s *WorkerSuite) job(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
	return 1, time.Second
}

func (s *WorkerSuite) jobSleepSixSeconds(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
	time.Sleep(time.Second * 6)
	return 1, time.Second
}

func (s *WorkerSuite) jobReturnsPanic(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
	panic("Panic!!!")
}

func (s *WorkerSuite) Test_CreateNewInstance_ReturnsStatusIsWait() {
	assert.Equal(s.T(), s.worker.GetStatus(), WorkerStatusWait)
}

func (s *WorkerSuite) Test_GetId_ReturnsId() {
	assert.IsType(s.T(), "", s.worker.GetId())
	assert.NotEmpty(s.T(), s.worker.GetId())
}

func (s *WorkerSuite) Test_GetCreateAt_ReturnsCreateAtTime() {
	assert.Equal(s.T(), s.worker.GetCreatedAt(), s.clock)
	assert.IsType(s.T(), time.Time{}, s.worker.GetCreatedAt())
}

func (s *WorkerSuite) Test_SetTask_Success() {
	t := task.NewTask(s.job)
	s.worker.setTask(t)

	assert.Equal(s.T(), s.worker.GetTask(), t)
}

func (s *WorkerSuite) Test_FirstRun_ReturnEmptyError() {
	var err error

	go func() {
		err = s.worker.Run()
	}()
	for s.worker.GetStatus() != WorkerStatusProcess {
	}

	assert.Nil(s.T(), err)
}

func (s *WorkerSuite) Test_TwiceRun_ReturnErrorForSecondRun() {
	go s.worker.Run()
	for s.worker.GetStatus() != WorkerStatusProcess {
	}

	err := s.worker.Run()

	assert.Equal(s.T(), err, errors.New("Worker is running"))
}

func (s *WorkerSuite) Test_IsNotRunning_ReturnStatusIsWait() {
	assert.Equal(s.T(), s.worker.GetStatus(), WorkerStatusWait)
}

func (s *WorkerSuite) Test_IsRunning_ReturnStatusIsProcess() {
	go s.worker.Run()
	for s.worker.GetStatus() != WorkerStatusProcess {
	}

	assert.Equal(s.T(), s.worker.GetStatus(), WorkerStatusProcess)
}

func (s *WorkerSuite) Test_IsRunningAndSendTask_ReturnStatusIsBusy() {
	t := task.NewTask(s.jobSleepSixSeconds)
	go s.worker.Run()
	for s.worker.GetStatus() != WorkerStatusProcess {
	}

	s.worker.SendTask(t)
	for s.worker.GetStatus() != WorkerStatusBusy {
	}

	assert.Equal(s.T(), s.worker.GetStatus(), WorkerStatusBusy)
}

func (s *WorkerSuite) Test_Reset_ReturnEmptyTask() {
	t := task.NewTask(s.job)
	go s.worker.Run()
	for s.worker.GetStatus() != WorkerStatusProcess {
	}

	s.worker.SendTask(t)
	for s.worker.GetStatus() != WorkerStatusWait {
	}

	assert.NotNil(s.T(), s.worker.GetTask())
	s.worker.Reset()
	assert.Nil(s.T(), s.worker.GetTask())
}

func (s *WorkerSuite) Test_IsRunningAndKill_ReturnEmptyError() {
	go s.worker.Run()
	for s.worker.GetStatus() != WorkerStatusProcess {
	}

	err := s.worker.Kill()

	assert.Nil(s.T(), err)
}

func (s *WorkerSuite) Test_CreateNewInstanceAndKill_ReturnError() {
	err := s.worker.Kill()

	assert.Equal(s.T(), err, errors.New("Worker isn't running"))
}

func (s *WorkerSuite) Test_IsRunningTwiceKill_ReturnErrorForSecondKill() {
	go s.worker.Run()
	for s.worker.GetStatus() != WorkerStatusProcess {
	}
	s.worker.Kill()
	for s.worker.GetStatus() != WorkerStatusWait {
	}

	err := s.worker.Kill()

	assert.Equal(s.T(), err, errors.New("Worker isn't running"))
}

func (s *WorkerSuite) Test_Kill_ReturnStatusIsWait() {
	t := task.NewTask(s.jobSleepSixSeconds)
	go s.worker.Run()
	for s.worker.GetStatus() != WorkerStatusProcess {
	}
	s.worker.SendTask(t)
	for s.worker.GetStatus() != WorkerStatusBusy {
	}
	s.worker.Kill()
	for s.worker.GetStatus() != WorkerStatusWait {
	}

	assert.Equal(s.T(), s.worker.GetStatus(), WorkerStatusWait)
}

func (s *WorkerSuite) Test_WithTaskReturnsPanic_SetTaskStatusIsFail() {
	go s.worker.Run()
	for s.worker.GetStatus() != WorkerStatusProcess {
	}

	t := task.NewTask(s.jobReturnsPanic)
	s.worker.SendTask(t)
	for s.worker.GetStatus() != WorkerStatusWait {
	}

	assert.Equal(s.T(), t.GetStatus(), task.TaskStatusFail)
}

func (s *WorkerSuite) Test_WithTaskWithTimeoutAndReturnsPanic_SetTaskStatusIsFail() {
	go s.worker.Run()
	for s.worker.GetStatus() != WorkerStatusProcess {
	}

	t := task.NewTask(s.jobReturnsPanic)
	t.SetTimeout(time.Second * 10)
	s.worker.SendTask(t)
	for s.worker.GetStatus() != WorkerStatusWait {
	}

	assert.Equal(s.T(), t.GetStatus(), task.TaskStatusFail)
}

func (s *WorkerSuite) Test_WithTask_SetTaskStatusIsKillIfWorkerKill() {
	go s.worker.Run()
	for s.worker.GetStatus() != WorkerStatusProcess {
	}

	t := task.NewTask(s.jobSleepSixSeconds)
	s.worker.SendTask(t)
	for s.worker.GetStatus() != WorkerStatusBusy {
	}

	s.worker.Kill()
	for s.worker.GetStatus() != WorkerStatusWait {
	}

	assert.Equal(s.T(), t.GetStatus(), task.TaskStatusKill)
}

func (s *WorkerSuite) Test_WithTaskWithTimeout_SetTaskStatusIsKillIfWorkerKill() {
	go s.worker.Run()
	for s.worker.GetStatus() != WorkerStatusProcess {
	}

	t := task.NewTask(s.jobSleepSixSeconds)
	t.SetTimeout(time.Second * 10)
	s.worker.SendTask(t)
	for s.worker.GetStatus() != WorkerStatusBusy {
	}

	s.worker.Kill()
	for s.worker.GetStatus() != WorkerStatusWait {
	}

	assert.Equal(s.T(), t.GetStatus(), task.TaskStatusKill)
}

func (s *WorkerSuite) Test_WithTaskWithTimeout_SetTaskStatusIsFailTimeout() {
	go s.worker.Run()
	for s.worker.GetStatus() != WorkerStatusProcess {
	}

	t := task.NewTask(s.jobSleepSixSeconds)
	t.SetTimeout(time.Second)
	s.worker.SendTask(t)
	for s.worker.GetStatus() != WorkerStatusWait {
	}

	assert.Equal(s.T(), t.GetStatus(), task.TaskStatusFailByTimeout)
}

func (s *WorkerSuite) Test_WithTask_SetTaskStatusIsSuccess() {
	go s.worker.Run()
	for s.worker.GetStatus() != WorkerStatusProcess {
	}

	t := task.NewTask(s.job)
	s.worker.SendTask(t)
	for s.worker.GetStatus() != WorkerStatusWait {
	}

	assert.Equal(s.T(), t.GetStatus(), task.TaskStatusSuccess)
}

func (s *WorkerSuite) Test_WithTaskWithTimeout_SetTaskStatusIsSuccess() {
	go s.worker.Run()
	for s.worker.GetStatus() != WorkerStatusProcess {
	}

	t := task.NewTask(s.jobSleepSixSeconds)
	t.SetTimeout(time.Second * 10)
	s.worker.SendTask(t)
	for s.worker.GetStatus() != WorkerStatusWait {
	}

	assert.Equal(s.T(), t.GetStatus(), task.TaskStatusSuccess)
}
