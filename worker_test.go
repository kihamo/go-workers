package workers

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

func jobSimple(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
	return 1, time.Second
}

func Test_WorkerNewInstance_StatusIsWait(t *testing.T) {
	worker := NewWorker()

	assert.Equal(t, worker.GetStatus(), WorkerStatusWait)
}

func Test_WorkerSetTask_Success(t *testing.T) {
	worker := NewWorker()
	task := NewTask("job", time.Second, time.Second, 1, jobSimple)

	worker.setTask(task)

	assert.Equal(t, worker.GetTask(), task)
}
