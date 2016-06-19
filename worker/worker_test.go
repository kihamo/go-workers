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

func workerJob(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
	return 1, time.Second
}

func workerJobSleepSixSeconds(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
	time.Sleep(time.Second * 6)
	return 1, time.Second
}

func workerJobReturnsPanic(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
	panic("Panic!!!")
}

func (suite *WorkerSuite) Test_WorkerNewInstance_ReturnsStatusIsWait() {
	done := make(chan Worker)
	w := NewWorker(done)

	assert.Equal(suite.T(), w.GetStatus(), WorkerStatusWait)
}

func (suite *WorkerSuite) Test_WorkerGetId_ReturnsId() {
	done := make(chan Worker)
	w := NewWorker(done)

	assert.Equal(suite.T(), w.GetId(), w.id)
	assert.IsType(suite.T(), "", w.GetId())
	assert.NotEmpty(suite.T(), w.GetId())
}

func (suite *WorkerSuite) Test_WorkerGetCreateAt_ReturnsCreateAtTime() {
	done := make(chan Worker)
	w := NewWorker(done)

	assert.Equal(suite.T(), w.GetCreatedAt(), w.createdAt)
	assert.IsType(suite.T(), time.Time{}, w.GetCreatedAt())
}

func (suite *WorkerSuite) Test_WorkerSetTask_Success() {
	t := task.NewTask("job", 0, time.Second, 1, workerJob)
	done := make(chan Worker)
	w := NewWorker(done)
	w.setTask(t)

	assert.Equal(suite.T(), w.GetTask(), t)
}

func (suite *WorkerSuite) Test_WorkerIsNotRunning_ReturnStatusIsWait() {
	done := make(chan Worker)
	w := NewWorker(done)

	assert.Equal(suite.T(), w.GetStatus(), WorkerStatusWait)
	w.Kill()
}

func (suite *WorkerSuite) Test_WorkerIsRunning_ReturnStatusIsProcess() {
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()

	time.Sleep(time.Second)

	assert.Equal(suite.T(), w.GetStatus(), WorkerStatusProcess)
	w.Kill()
}

func (suite *WorkerSuite) Test_WorkerIsRunningAndSendTask_ReturnStatusIsBusy() {
	t := task.NewTask("job", 0, 0, 1, workerJobSleepSixSeconds)
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	w.SendTask(t)

	time.Sleep(time.Second)

	assert.Equal(suite.T(), w.GetStatus(), WorkerStatusBusy)
	w.Kill()
}

func (suite *WorkerSuite) Test_WorkerReset_ReturnEmptyTask() {
	t := task.NewTask("job", 0, 0, 1, workerJob)
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	w.SendTask(t)

	for {
		select {
		case <-done:
			assert.NotNil(suite.T(), w.GetTask())
			w.Reset()
			assert.Nil(suite.T(), w.GetTask())
			return
		}
	}
}

func (suite *WorkerSuite) Test_WorkerKill_ReturnStatusIsWait() {
	t := task.NewTask("job", 0, 0, 1, workerJobSleepSixSeconds)
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
			assert.Equal(suite.T(), w.GetStatus(), WorkerStatusWait)
			return
		}
	}
}

func (suite *WorkerSuite) Test_WorkerWithTaskReturnsPanic_SetTaskStatusIsFail() {
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	t := task.NewTask("job", 0, 0, 1, workerJobReturnsPanic)
	sendTask := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				w.SendTask(t)
				sendTask = true
			}

		case <-done:
			assert.Equal(suite.T(), t.GetStatus(), task.TaskStatusFail)
			return
		}
	}
}

func (suite *WorkerSuite) Test_WorkerWithTaskWithTimeoutAndReturnsPanic_SetTaskStatusIsFail() {
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	t := task.NewTask("job", 0, time.Second*10, 0, workerJobReturnsPanic)
	sendTask := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				w.SendTask(t)
				sendTask = true
			}

		case <-done:
			assert.Equal(suite.T(), t.GetStatus(), task.TaskStatusFail)
			return
		}
	}
}

func (suite *WorkerSuite) Test_WorkerWithTask_SetTaskStatusIsKillIfWorkerKill() {
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	t := task.NewTask("job", 0, 0, 0, workerJobSleepSixSeconds)
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
			assert.Equal(suite.T(), t.GetStatus(), task.TaskStatusKill)
			return
		}
	}
}

func (suite *WorkerSuite) Test_WorkerWithTaskWithTimeout_SetTaskStatusIsKillIfWorkerKill() {
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	t := task.NewTask("job", 0, time.Second*10, 0, workerJobSleepSixSeconds)
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
			assert.Equal(suite.T(), t.GetStatus(), task.TaskStatusKill)
			return
		}
	}
}

func (suite *WorkerSuite) Test_WorkerWithTaskWithTimeout_SetTaskStatusIsFailTimeout() {
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	t := task.NewTask("job", 0, time.Second*1, 1, workerJobSleepSixSeconds)
	sendTask := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				w.SendTask(t)
				sendTask = true
			}

		case <-done:
			assert.Equal(suite.T(), t.GetStatus(), task.TaskStatusFailByTimeout)
			return
		}
	}
}

func (suite *WorkerSuite) Test_WorkerWithTask_SetTaskStatusIsSuccess() {
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	t := task.NewTask("job", 0, 0, 1, workerJob)
	sendTask := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				w.SendTask(t)
				sendTask = true
			}

		case <-done:
			assert.Equal(suite.T(), t.GetStatus(), task.TaskStatusSuccess)
			return
		}
	}
}

func (suite *WorkerSuite) Test_WorkerWithTaskWithTimeout_SetTaskStatusIsSuccess() {
	done := make(chan Worker)
	w := NewWorker(done)
	go w.Run()
	t := task.NewTask("job", 0, time.Second*10, 1, workerJobSleepSixSeconds)
	sendTask := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				w.SendTask(t)
				sendTask = true
			}

		case <-done:
			assert.Equal(suite.T(), t.GetStatus(), task.TaskStatusSuccess)
			return
		}
	}
}
