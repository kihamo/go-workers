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

func (suite *WorkerSuite) Test_WorkerNewInstance_StatusIsWait() {
	done := make(chan *Worker)
	worker := NewWorker(done)

	assert.Equal(suite.T(), worker.GetStatus(), WorkerStatusWait)
}

func (suite *WorkerSuite) Test_WorkerGetId_ReturnId() {

	done := make(chan *Worker)
	worker := NewWorker(done)

	assert.Equal(suite.T(), worker.GetId(), worker.id)
	assert.IsType(suite.T(), "", worker.GetId())
	assert.NotEmpty(suite.T(), worker.GetId())
}

func (suite *WorkerSuite) Test_WorkerGetCreateAt_ReturnCreateAtTime() {
	done := make(chan *Worker)
	worker := NewWorker(done)

	assert.Equal(suite.T(), worker.GetCreatedAt(), worker.createdAt)
	assert.IsType(suite.T(), time.Time{}, worker.GetCreatedAt())
}

func (suite *WorkerSuite) Test_WorkerSetStatus_ReturnChangeStatus() {
	done := make(chan *Worker)
	worker := NewWorker(done)
	worker.setStatus(WorkerStatusBusy)

	assert.Equal(suite.T(), worker.GetStatus(), WorkerStatusBusy)
}

func (suite *WorkerSuite) Test_WorkerKill_ReturnStatusIsWait() {
	task := NewTask("jobSimpleSleepFiveSeconds", time.Second, 0, 1, jobSimpleSleepSixSeconds)

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
	task := NewTask("job", time.Second, time.Second, 1, jobSimple)

	worker.setTask(task)

	assert.Equal(suite.T(), worker.GetTask(), task)
}
