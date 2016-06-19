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

func jobSimple(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
	return 1, time.Second
}

func jobSimpleSleepSixSeconds(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
	time.Sleep(time.Second * 6)
	return 1, time.Second
}

func jobSimpleReturnsPanic(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
	panic("Panic!!!")
}

func jobSimpleWithTimeoutAndReturnsPanic(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
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

func (suite *WorkerSuite) Test_WorkerKill_ReturnStatusIsWait() {
	task := NewTask("job", 0, 0, 1, jobSimpleSleepSixSeconds)
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
				continue
			}

			assert.Equal(suite.T(), worker.GetStatus(), WorkerStatusWait)
			return
		}
	}

}

func (suite *WorkerSuite) Test_WorkerSetTask_Success() {
	done := make(chan *Worker)
	worker := NewWorker(done)
	task := NewTask("job", 0, time.Second, 1, jobSimple)

	worker.setTask(task)

	assert.Equal(suite.T(), worker.GetTask(), task)
}

func (suite *WorkerSuite) Test_WorkerWithTaskReturnsPanic_SetTaskStatusIsFail() {
	done := make(chan *Worker)
	worker := NewWorker(done)
	go worker.run()
	task := NewTask("job", 0, 0, 1, jobSimpleReturnsPanic)
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
	task := NewTask("job", 0, 0, 0, jobSimpleWithTimeoutAndReturnsPanic)
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

func (suite *WorkerSuite) Test_WorkerWithTaskWithTimeout_SetTaskStatusIsFailTimeout() {
	done := make(chan *Worker)
	worker := NewWorker(done)
	go worker.run()
	task := NewTask("job", 0, time.Second*1, 1, jobSimpleSleepSixSeconds)
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
	task := NewTask("job", 0, 0, 1, jobSimple)
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
	task := NewTask("job", 0, time.Second*10, 1, jobSimpleSleepSixSeconds)
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
