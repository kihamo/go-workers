package dispatcher

import (
	"errors"
	"testing"
	"time"

	"github.com/kihamo/go-workers"
	"github.com/kihamo/go-workers/task"
	"github.com/kihamo/go-workers/worker"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DispatcherSuite struct {
	suite.Suite

	clock     *fakeclock.FakeClock
	clockTime time.Time

	dispatcher *Dispatcher
}

var (
	jobFuncInNonExportVariable = func(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
		return 1, time.Second
	}
	jobFuncInExportVariable = func(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
		return 1, time.Second
	}
)

func jobFunc(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
	return 1, time.Second
}

func TestDispatcherSuite(t *testing.T) {
	suite.Run(t, new(DispatcherSuite))
}

func (s *DispatcherSuite) SetupTest() {
	s.dispatcher = NewDispatcher()

	s.clockTime = time.Date(2016, 6, 5, 4, 3, 2, 1, time.UTC)

	workers.Clock = fakeclock.NewFakeClock(s.clockTime)
	s.clock = fakeclock.NewFakeClock(s.clockTime)
}

func (s *DispatcherSuite) TearDownTest() {
	s.dispatcher.Kill()
}

func (s *DispatcherSuite) jobSleepSixSeconds(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
	s.clock.Sleep(time.Second * 6)
	return 1, time.Second
}

func (s *DispatcherSuite) Test_FirstRun_ReturnEmptyError() {
	var err error

	go func() {
		err = s.dispatcher.Run()
	}()

	assert.Nil(s.T(), err)
}

func (s *DispatcherSuite) Test_TwiceRun_ReturnErrorForSecondRun() {
	go s.dispatcher.Run()
	for s.dispatcher.GetStatus() != DispatcherStatusProcess {
	}

	err := s.dispatcher.Run()

	assert.Equal(s.T(), err, errors.New("Dispatcher is running"))
}

func (s *DispatcherSuite) Test_CreateNewInstance_ReturnsStatusWait() {
	assert.Equal(s.T(), s.dispatcher.GetStatus(), DispatcherStatusWait)
}

func (s *DispatcherSuite) Test_Run_ReturnsStatusProcess() {
	go s.dispatcher.Run()
	for s.dispatcher.GetStatus() != DispatcherStatusProcess {
	}

	assert.Equal(s.T(), s.dispatcher.GetStatus(), DispatcherStatusProcess)
	s.dispatcher.Kill()
}

func (s *DispatcherSuite) Test_IsRunningAndKill_ReturnEmptyError() {
	go s.dispatcher.Run()
	for s.dispatcher.GetStatus() != DispatcherStatusProcess {
	}

	err := s.dispatcher.Kill()

	assert.Nil(s.T(), err)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndKill_ReturnError() {
	err := s.dispatcher.Kill()

	assert.Equal(s.T(), err, errors.New("Dispatcher isn't running"))
}

func (s *DispatcherSuite) Test_IsRunningTwiceKill_ReturnErrorForSecondKill() {
	go s.dispatcher.Run()
	for s.dispatcher.GetStatus() != DispatcherStatusProcess {
	}

	s.dispatcher.Kill()
	for s.dispatcher.GetStatus() != DispatcherStatusWait {
	}

	err := s.dispatcher.Kill()

	assert.Equal(s.T(), err, errors.New("Dispatcher isn't running"))
}

func (s *DispatcherSuite) Test_Kill_ReturnsStatusWait() {
	go s.dispatcher.Run()
	for s.dispatcher.GetStatus() != DispatcherStatusProcess {
	}

	s.dispatcher.Kill()
	for s.dispatcher.GetStatus() != DispatcherStatusWait {
	}

	assert.Equal(s.T(), s.dispatcher.GetStatus(), DispatcherStatusWait)
	s.dispatcher.Kill()
}

func (s *DispatcherSuite) Test_CreateNewInstance_ReturnsZeroSizeOfWorkersList() {
	assert.Equal(s.T(), s.dispatcher.GetWorkers().Len(), 0)
}

func (s *DispatcherSuite) Test_AddOneWorker_ReturnsOneSizeOfWorkersList() {
	s.dispatcher.AddWorker()

	assert.Equal(s.T(), s.dispatcher.GetWorkers().Len(), 1)
	s.dispatcher.Kill()
}

func (s *DispatcherSuite) Test_CreateNewInstance_ReturnsZeroSizeOfTasksList() {
	assert.Equal(s.T(), s.dispatcher.GetTasks().Len(), 0)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTask_ReturnZeroSizeOfTasksList() {
	s.dispatcher.AddTaskByFunc(s.jobSleepSixSeconds)

	assert.Equal(s.T(), s.dispatcher.GetTasks().Len(), 0)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskAndRun_ReturnsOneSizeOfTasksList() {
	w := s.dispatcher.AddWorker()
	s.dispatcher.AddTaskByFunc(s.jobSleepSixSeconds)
	go s.dispatcher.Run()
	for w.GetStatus() != worker.WorkerStatusBusy {
	}

	assert.Equal(s.T(), s.dispatcher.GetTasks().Len(), 1)
}

func (s *DispatcherSuite) Test_IsRunningAndAddTask_ReturnOneSizeOfTasksList() {
	go s.dispatcher.Run()
	for s.dispatcher.GetStatus() != DispatcherStatusProcess {
	}

	s.dispatcher.AddWorker()
	s.dispatcher.AddTaskByFunc(s.jobSleepSixSeconds)

	assert.Equal(s.T(), s.dispatcher.GetTasks().Len(), 1)
}

func (s *DispatcherSuite) Test_IsRunningAndAddTaskWithDuration_ReturnsZeroSizeOfTasksListBeforeExpirationDuration() {
	go s.dispatcher.Run()
	for s.dispatcher.GetStatus() != DispatcherStatusProcess {
	}

	t := task.NewTask(s.jobSleepSixSeconds)
	t.SetDuration(time.Second * 2)
	s.dispatcher.AddTask(t)

	assert.Equal(s.T(), s.dispatcher.GetTasks().Len(), 0)
}

func (s *DispatcherSuite) Test_IsRunningAndAddTaskWithDuration_ReturnsOneSizeOfTasksListAfterExpirationDuration() {
	clock := workers.Clock.(*fakeclock.FakeClock)

	go s.dispatcher.Run()
	for s.dispatcher.GetStatus() != DispatcherStatusProcess {
	}

	w := s.dispatcher.AddWorker()
	for w.GetStatus() != worker.WorkerStatusWait {
	}

	t := task.NewTask(s.jobSleepSixSeconds)
	t.SetDuration(time.Second * 2)
	s.dispatcher.AddTask(t)

	assert.Equal(s.T(), s.dispatcher.GetTasks().Len(), 0)
	clock.IncrementBySeconds(2)
	for t.GetStatus() != task.TaskStatusProcess {
	}

	assert.Equal(s.T(), s.dispatcher.GetTasks().Len(), 1)
}

func (s *DispatcherSuite) Test_CreateNewInstance_ReturnsZeroSizeOfWaitTasksList() {
	assert.Equal(s.T(), s.dispatcher.GetWaitTasks().Len(), 0)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTask_ReturnsOneSizeOfWaitTasksList() {
	s.dispatcher.AddTaskByFunc(s.jobSleepSixSeconds)

	assert.Equal(s.T(), s.dispatcher.GetWaitTasks().Len(), 1)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskAndRun_ReturnsZeroSizeOfWaitTasksList() {
	w := s.dispatcher.AddWorker()
	s.dispatcher.AddTaskByFunc(s.jobSleepSixSeconds)
	go s.dispatcher.Run()
	for w.GetStatus() != worker.WorkerStatusBusy {
	}

	assert.Equal(s.T(), s.dispatcher.GetWaitTasks().Len(), 0)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskByInnerFunc_ReturnsTask() {
	t := s.dispatcher.AddTaskByFunc(s.jobSleepSixSeconds)

	assert.IsType(s.T(), &task.Task{}, t)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskByInnerFunc_ReturnsTaskWithAutoGenerateName() {
	t := s.dispatcher.AddTaskByFunc(s.jobSleepSixSeconds)

	assert.Equal(s.T(), "github.com/kihamo/go-workers/dispatcher.jobSleepSixSeconds", t.GetName())
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskByFunc_ReturnsTaskWithAutoGenerateName() {
	t := s.dispatcher.AddTaskByFunc(jobFunc)

	assert.Equal(s.T(), "github.com/kihamo/go-workers/dispatcher.jobFunc", t.GetName())
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskByAnonymousFunc_ReturnsTaskWithAutoGenerateName() {
	t := s.dispatcher.AddTaskByFunc(func(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
		return 1, time.Second
	})

	assert.Equal(s.T(), "github.com/kihamo/go-workers/dispatcher.func", t.GetName())
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskByAnonymousFuncFromNonExportVariable_ReturnsTaskWithAutoGenerateName() {
	t := s.dispatcher.AddTaskByFunc(jobFuncInNonExportVariable)

	assert.Equal(s.T(), "github.com/kihamo/go-workers/dispatcher.func", t.GetName())
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskByAnonymousFuncFromExportVariable_ReturnsTaskWithAutoGenerateName() {
	t := s.dispatcher.AddTaskByFunc(jobFuncInExportVariable)

	assert.Equal(s.T(), "github.com/kihamo/go-workers/dispatcher.func", t.GetName())
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddNamedTaskByInnerFunc_ReturnsTask() {
	t := s.dispatcher.AddNamedTaskByFunc("task.test", s.jobSleepSixSeconds)

	assert.IsType(s.T(), &task.Task{}, t)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddNameTaskByInnerFunc_ReturnsTaskWithAutoGenerateName() {
	t := s.dispatcher.AddNamedTaskByFunc("task.test", s.jobSleepSixSeconds)

	assert.Equal(s.T(), "task.test", t.GetName())
}
