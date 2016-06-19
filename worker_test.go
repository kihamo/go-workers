package workers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type WorkerSuite struct {
	suite.Suite
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
	done := make(chan *Worker)
	worker := NewWorker(done)

	assert.Equal(suite.T(), worker.GetStatus(), WorkerStatusWait)
}

func (suite *WorkerSuite) Test_WorkerGetId_ReturnsId() {

	done := make(chan *Worker)
	worker := NewWorker(done)

	assert.Equal(suite.T(), worker.GetId(), worker.id)
	assert.IsType(suite.T(), "", worker.GetId())
	assert.NotEmpty(suite.T(), worker.GetId())
}

func (suite *WorkerSuite) Test_WorkerGetCreateAt_ReturnsCreateAtTime() {
	done := make(chan *Worker)
	worker := NewWorker(done)

	assert.Equal(suite.T(), worker.GetCreatedAt(), worker.createdAt)
	assert.IsType(suite.T(), time.Time{}, worker.GetCreatedAt())
}

func (suite *WorkerSuite) Test_WorkerSetStatus_ReturnsChangeStatus() {
	done := make(chan *Worker)
	worker := NewWorker(done)
	worker.setStatus(WorkerStatusBusy)

	assert.Equal(suite.T(), worker.GetStatus(), WorkerStatusBusy)
}

func (suite *WorkerSuite) Test_WorkerSetTask_Success() {
	task := NewTask("job", 0, time.Second, 1, workerJob)
	done := make(chan *Worker)
	worker := NewWorker(done)
	worker.setTask(task)

	assert.Equal(suite.T(), worker.GetTask(), task)
}

func (suite *WorkerSuite) Test_WorkerWithTask_ReturnStatusIsProcess() {
	task := NewTask("job", 0, 0, 1, workerJobSleepSixSeconds)
	done := make(chan *Worker)
	worker := NewWorker(done)
	go worker.run()
	worker.sendTask(task)

	time.Sleep(time.Second)

	assert.Equal(suite.T(), worker.GetStatus(), WorkerStatusProcess)
	worker.Kill()
}

func (suite *WorkerSuite) Test_WorkerKill_ReturnStatusIsWait() {
	task := NewTask("job", 0, 0, 1, workerJobSleepSixSeconds)
	done := make(chan *Worker)
	worker := NewWorker(done)
	go worker.run()
	worker.sendTask(task)
	sendKillSignal := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendKillSignal && worker.GetStatus() == WorkerStatusBusy {
				worker.Kill()
				sendKillSignal = true
			}
		case <-done:
			assert.Equal(suite.T(), worker.GetStatus(), WorkerStatusWait)
			return
		}
	}
}

func (suite *WorkerSuite) Test_WorkerWithTaskReturnsPanic_SetTaskStatusIsFail() {
	done := make(chan *Worker)
	worker := NewWorker(done)
	go worker.run()
	task := NewTask("job", 0, 0, 1, workerJobReturnsPanic)
	sendTask := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				worker.sendTask(task)
				sendTask = true
			}

		case <-done:
			assert.Equal(suite.T(), task.GetStatus(), TaskStatusFail)
			return
		}
	}
}

func (suite *WorkerSuite) Test_WorkerWithTaskWithTimeoutAndReturnsPanic_SetTaskStatusIsFail() {
	done := make(chan *Worker)
	worker := NewWorker(done)
	go worker.run()
	task := NewTask("job", 0, time.Second*10, 0, workerJobReturnsPanic)
	sendTask := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				worker.sendTask(task)
				sendTask = true
			}

		case <-done:
			assert.Equal(suite.T(), task.GetStatus(), TaskStatusFail)
			return
		}
	}
}

func (suite *WorkerSuite) Test_WorkerWithTask_SetTaskStatusIsKillIfWorkerKill() {
	done := make(chan *Worker)
	worker := NewWorker(done)
	go worker.run()
	task := NewTask("job", 0, 0, 0, workerJobSleepSixSeconds)
	sendTask := false
	sendKillSignal := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				worker.sendTask(task)
				sendTask = true
				continue
			}

			if !sendKillSignal {
				worker.Kill()
			}

		case <-done:
			assert.Equal(suite.T(), task.GetStatus(), TaskStatusKill)
			return
		}
	}
}

func (suite *WorkerSuite) Test_WorkerWithTaskWithTimeout_SetTaskStatusIsKillIfWorkerKill() {
	done := make(chan *Worker)
	worker := NewWorker(done)
	go worker.run()
	task := NewTask("job", 0, time.Second*10, 0, workerJobSleepSixSeconds)
	sendTask := false
	sendKillSignal := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				worker.sendTask(task)
				sendTask = true
				continue
			}

			if !sendKillSignal {
				worker.Kill()
			}

		case <-done:
			assert.Equal(suite.T(), task.GetStatus(), TaskStatusKill)
			return
		}
	}
}

func (suite *WorkerSuite) Test_WorkerWithTaskWithTimeout_SetTaskStatusIsFailTimeout() {
	done := make(chan *Worker)
	worker := NewWorker(done)
	go worker.run()
	task := NewTask("job", 0, time.Second*1, 1, workerJobSleepSixSeconds)
	sendTask := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				worker.sendTask(task)
				sendTask = true
			}

		case <-done:
			assert.Equal(suite.T(), task.GetStatus(), TaskStatusFailByTimeout)
			return
		}
	}
}

func (suite *WorkerSuite) Test_WorkerWithTask_SetTaskStatusIsSuccess() {
	done := make(chan *Worker)
	worker := NewWorker(done)
	go worker.run()
	task := NewTask("job", 0, 0, 1, workerJob)
	sendTask := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				worker.sendTask(task)
				sendTask = true
			}

		case <-done:
			assert.Equal(suite.T(), task.GetStatus(), TaskStatusSuccess)
			return
		}
	}
}

func (suite *WorkerSuite) Test_WorkerWithTaskWithTimeout_SetTaskStatusIsSuccess() {
	done := make(chan *Worker)
	worker := NewWorker(done)
	go worker.run()
	task := NewTask("job", 0, time.Second*10, 1, workerJobSleepSixSeconds)
	sendTask := false

	for {
		select {
		case <-time.After(time.Second):
			if !sendTask {
				worker.sendTask(task)
				sendTask = true
			}

		case <-done:
			assert.Equal(suite.T(), task.GetStatus(), TaskStatusSuccess)
			return
		}
	}
}
