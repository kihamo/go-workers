package worker

import (
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
	clock time.Time
}

func (s *WorkerSuite) SetupTest() {
	s.clock = time.Date(2016, 6, 5, 4, 3, 2, 1, time.UTC)
	workers.Clock = fakeclock.NewFakeClock(s.clock)
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

func (s *WorkerSuite) Test_NewInstance_ReturnsStatusIsWait() {
	done := make(chan Worker)
	w := NewWorker(done)

	assert.Equal(s.T(), w.GetStatus(), WorkerStatusWait)
}

func (s *WorkerSuite) Test_GetId_ReturnsId() {
	done := make(chan Worker)
	w := NewWorker(done)

	assert.IsType(s.T(), "", w.GetId())
	assert.NotEmpty(s.T(), w.GetId())
}

func (s *WorkerSuite) Test_GetCreateAt_ReturnsCreateAtTime() {
	done := make(chan Worker)
	w := NewWorker(done)

	assert.Equal(s.T(), w.GetCreatedAt(), w.createdAt)
	assert.IsType(s.T(), time.Time{}, w.GetCreatedAt())
}

func (s *WorkerSuite) Test_SetTask_Success() {
	t := task.NewTask(s.job)
	done := make(chan Worker)
	w := NewWorker(done)
	w.setTask(t)

	assert.Equal(s.T(), w.GetTask(), t)
}

func (s *WorkerSuite) Test_IsNotRunning_ReturnStatusIsWait() {
	done := make(chan Worker)
	w := NewWorker(done)

	assert.Equal(s.T(), w.GetStatus(), WorkerStatusWait)
	w.Kill()
}

func (s *WorkerSuite) Test_IsRunning_ReturnStatusIsProcess() {
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()

	time.Sleep(time.Second)

	assert.Equal(s.T(), w.GetStatus(), WorkerStatusProcess)
	w.Kill()
}

func (s *WorkerSuite) Test_IsRunningAndSendTask_ReturnStatusIsBusy() {
	t := task.NewTask(s.jobSleepSixSeconds)
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	w.SendTask(t)

	time.Sleep(time.Second)

	assert.Equal(s.T(), w.GetStatus(), WorkerStatusBusy)
	w.Kill()
}

func (s *WorkerSuite) Test_Reset_ReturnEmptyTask() {
	t := task.NewTask(s.job)
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	w.SendTask(t)

	for {
		select {
		case <-done:
			assert.NotNil(s.T(), w.GetTask())
			w.Reset()
			assert.Nil(s.T(), w.GetTask())
			return
		}
	}
}

func (s *WorkerSuite) Test_Kill_ReturnStatusIsWait() {
	t := task.NewTask(s.jobSleepSixSeconds)
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	w.SendTask(t)
	sendKillSignal := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendKillSignal && w.GetStatus() == WorkerStatusBusy {
				w.Kill()
				sendKillSignal = true
			}
		case <-done:
			assert.Equal(s.T(), w.GetStatus(), WorkerStatusWait)
			return
		}
	}
}

func (s *WorkerSuite) Test_WithTaskReturnsPanic_SetTaskStatusIsFail() {
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	t := task.NewTask(s.jobReturnsPanic)
	sendTask := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				w.SendTask(t)
				sendTask = true
			}

		case <-done:
			assert.Equal(s.T(), t.GetStatus(), task.TaskStatusFail)
			return
		}
	}
}

func (s *WorkerSuite) Test_WithTaskWithTimeoutAndReturnsPanic_SetTaskStatusIsFail() {
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	t := task.NewTask(s.jobReturnsPanic)
	t.SetTimeout(time.Second * 10)
	sendTask := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				w.SendTask(t)
				sendTask = true
			}

		case <-done:
			assert.Equal(s.T(), t.GetStatus(), task.TaskStatusFail)
			return
		}
	}
}

func (s *WorkerSuite) Test_WithTask_SetTaskStatusIsKillIfWorkerKill() {
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	t := task.NewTask(s.jobSleepSixSeconds)
	sendTask := false
	sendKillSignal := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				w.SendTask(t)
				sendTask = true
				continue
			}

			if !sendKillSignal {
				w.Kill()
			}

		case <-done:
			assert.Equal(s.T(), t.GetStatus(), task.TaskStatusKill)
			return
		}
	}
}

func (s *WorkerSuite) Test_WithTaskWithTimeout_SetTaskStatusIsKillIfWorkerKill() {
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	t := task.NewTask(s.jobSleepSixSeconds)
	t.SetTimeout(time.Second * 10)
	sendTask := false
	sendKillSignal := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				w.SendTask(t)
				sendTask = true
				continue
			}

			if !sendKillSignal {
				w.Kill()
			}

		case <-done:
			assert.Equal(s.T(), t.GetStatus(), task.TaskStatusKill)
			return
		}
	}
}

func (s *WorkerSuite) Test_WithTaskWithTimeout_SetTaskStatusIsFailTimeout() {
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	t := task.NewTask(s.jobSleepSixSeconds)
	t.SetTimeout(time.Second)
	sendTask := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				w.SendTask(t)
				sendTask = true
			}

		case <-done:
			assert.Equal(s.T(), t.GetStatus(), task.TaskStatusFailByTimeout)
			return
		}
	}
}

func (s *WorkerSuite) Test_WithTask_SetTaskStatusIsSuccess() {
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	t := task.NewTask(s.job)
	sendTask := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				w.SendTask(t)
				sendTask = true
			}

		case <-done:
			assert.Equal(s.T(), t.GetStatus(), task.TaskStatusSuccess)
			return
		}
	}
}

func (s *WorkerSuite) Test_WithTaskWithTimeout_SetTaskStatusIsSuccess() {
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	t := task.NewTask(s.jobSleepSixSeconds)
	t.SetTimeout(time.Second * 10)
	sendTask := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				w.SendTask(t)
				sendTask = true
			}

		case <-done:
			assert.Equal(s.T(), t.GetStatus(), task.TaskStatusSuccess)
			return
		}
	}
}
