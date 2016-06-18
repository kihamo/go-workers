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

func jobSimple(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
	return 1, time.Second
}

func Test_WorkerNewInstance_StatusIsWait(t *testing.T) {
	suite.Run(t, new(WorkerSuite))

	worker := NewWorker()

	assert.Equal(t, worker.GetStatus(), WorkerStatusWait)
}

func Test_WorkerSetTask_Success(t *testing.T) {
	suite.Run(t, new(WorkerSuite))

	worker := NewWorker()
	task := NewTask("job", time.Second, time.Second, 1, jobSimple)

	worker.setTask(task)

	assert.Equal(t, worker.GetTask(), task)
}
