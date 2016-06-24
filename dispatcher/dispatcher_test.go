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

func (s *DispatcherSuite) SetupSuite() {
	s.clockTime = time.Date(2016, 6, 5, 4, 3, 2, 1, time.UTC)

	workers.Clock = fakeclock.NewFakeClock(s.clockTime)
	s.clock = fakeclock.NewFakeClock(s.clockTime)
}

func (s *DispatcherSuite) SetupTest() {
	// reset sleep timer in jobSleepSixSeconds
	if s.clock.WatcherCount() > 0 {
		s.clock.WaitForWatcherAndIncrement(time.Second * 6)
	}
}

func (s *DispatcherSuite) TearDownTest() {
	// reset sleep timer in jobSleepSixSeconds
	if s.clock.WatcherCount() > 0 {
		s.clock.WaitForWatcherAndIncrement(time.Second * 6)
	}
}

func (s *DispatcherSuite) jobSleepSixSeconds(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
	s.clock.Sleep(time.Second * 6)
	return 1, time.Second
}

func (s *DispatcherSuite) func1(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
	s.clock.Sleep(time.Second * 6)
	return 1, time.Second
}

func (s *DispatcherSuite) Test_FirstRun_ReturnEmptyError() {
	var err error

	go func() {
		d := NewDispatcher()
		err = d.Run()
	}()

	assert.Nil(s.T(), err)
}

func (s *DispatcherSuite) Test_TwiceRun_ReturnErrorForSecondRun() {
	d := NewDispatcher()
	go d.Run()
	for d.GetStatus() != DispatcherStatusProcess {
	}

	err := d.Run()

	assert.Equal(s.T(), err, errors.New("Dispatcher is running"))
	d.Kill()
}

func (s *DispatcherSuite) Test_CreateNewInstance_ReturnsStatusWait() {
	d := NewDispatcher()

	assert.Equal(s.T(), d.GetStatus(), DispatcherStatusWait)
}

func (s *DispatcherSuite) Test_Run_ReturnsStatusProcess() {
	d := NewDispatcher()
	go d.Run()
	for d.GetStatus() != DispatcherStatusProcess {
	}

	assert.Equal(s.T(), d.GetStatus(), DispatcherStatusProcess)
	d.Kill()
}

func (s *DispatcherSuite) Test_IsRunningAndKill_ReturnEmptyError() {
	d := NewDispatcher()
	go d.Run()
	for d.GetStatus() != DispatcherStatusProcess {
	}

	err := d.Kill()

	assert.Nil(s.T(), err)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndKill_ReturnError() {
	d := NewDispatcher()
	err := d.Kill()

	assert.Equal(s.T(), err, errors.New("Dispatcher isn't running"))
}

func (s *DispatcherSuite) Test_IsRunningTwiceKill_ReturnErrorForSecondKill() {
	d := NewDispatcher()
	go d.Run()
	for d.GetStatus() != DispatcherStatusProcess {
	}

	d.Kill()
	for d.GetStatus() != DispatcherStatusWait {
	}

	err := d.Kill()

	assert.Equal(s.T(), err, errors.New("Dispatcher isn't running"))
}

func (s *DispatcherSuite) Test_Kill_ReturnsStatusWait() {
	d := NewDispatcher()
	go d.Run()
	for d.GetStatus() != DispatcherStatusProcess {
	}

	d.Kill()
	for d.GetStatus() != DispatcherStatusWait {
	}

	assert.Equal(s.T(), d.GetStatus(), DispatcherStatusWait)
	d.Kill()
}

func (s *DispatcherSuite) Test_CreateNewInstance_ReturnsZeroSizeOfWorkersList() {
	d := NewDispatcher()

	assert.Equal(s.T(), d.GetWorkers().Len(), 0)
}

func (s *DispatcherSuite) Test_AddOneWorker_ReturnsOneSizeOfWorkersList() {
	d := NewDispatcher()
	d.AddWorker()

	assert.Equal(s.T(), d.GetWorkers().Len(), 1)
}

func (s *DispatcherSuite) Test_CreateNewInstance_ReturnsZeroSizeOfTasksList() {
	d := NewDispatcher()

	assert.Equal(s.T(), d.GetTasks().Len(), 0)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTask_ReturnZeroSizeOfTasksList() {
	d := NewDispatcher()
	d.AddTaskByFunc(s.jobSleepSixSeconds)

	assert.Equal(s.T(), d.GetTasks().Len(), 0)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskAndRun_ReturnsOneSizeOfTasksList() {
	d := NewDispatcher()

	w := d.AddWorker()
	d.AddTaskByFunc(s.jobSleepSixSeconds)
	go d.Run()
	for w.GetStatus() != worker.WorkerStatusBusy {
	}

	assert.Equal(s.T(), d.GetTasks().Len(), 1)
}

func (s *DispatcherSuite) Test_IsRunningAndAddTask_ReturnOneSizeOfTasksList() {
	d := NewDispatcher()
	go d.Run()
	for d.GetStatus() != DispatcherStatusProcess {
	}

	d.AddWorker()
	d.AddTaskByFunc(s.jobSleepSixSeconds)

	assert.Equal(s.T(), d.GetTasks().Len(), 1)
	d.Kill()
}

func (s *DispatcherSuite) Test_IsRunningAndAddTaskWithDuration_ReturnsZeroSizeOfTasksListBeforeExpirationDuration() {
	d := NewDispatcher()
	go d.Run()
	for d.GetStatus() != DispatcherStatusProcess {
	}

	t := task.NewTask(s.jobSleepSixSeconds)
	t.SetDuration(time.Second * 2)
	d.AddTask(t)

	assert.Equal(s.T(), d.GetTasks().Len(), 0)
	d.Kill()
}

func (s *DispatcherSuite) Test_IsRunningAndAddTaskWithDuration_ReturnsOneSizeOfTasksListAfterExpirationDuration() {
	clock := workers.Clock.(*fakeclock.FakeClock)

	d := NewDispatcher()
	go d.Run()
	for d.GetStatus() != DispatcherStatusProcess {
	}

	w := d.AddWorker()
	for w.GetStatus() != worker.WorkerStatusWait {
	}

	t := task.NewTask(s.jobSleepSixSeconds)
	t.SetDuration(time.Second * 2)
	d.AddTask(t)

	assert.Equal(s.T(), d.GetTasks().Len(), 0)
	clock.IncrementBySeconds(2)
	for t.GetStatus() != task.TaskStatusProcess {
	}

	assert.Equal(s.T(), d.GetTasks().Len(), 1)
	d.Kill()
}

func (s *DispatcherSuite) Test_CreateNewInstance_ReturnsZeroSizeOfWaitTasksList() {
	d := NewDispatcher()

	assert.Equal(s.T(), d.GetWaitTasks().Len(), 0)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTask_ReturnsOneSizeOfWaitTasksList() {
	d := NewDispatcher()
	d.AddTaskByFunc(s.jobSleepSixSeconds)

	assert.Equal(s.T(), d.GetWaitTasks().Len(), 1)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskAndRun_ReturnsZeroSizeOfWaitTasksList() {
	d := NewDispatcher()
	w := d.AddWorker()
	d.AddTaskByFunc(s.jobSleepSixSeconds)
	go d.Run()
	for w.GetStatus() != worker.WorkerStatusBusy {
	}

	assert.Equal(s.T(), d.GetWaitTasks().Len(), 0)
	d.Kill()
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskByInnerFunc_ReturnsTask() {
	d := NewDispatcher()
	t := d.AddTaskByFunc(s.jobSleepSixSeconds)

	assert.IsType(s.T(), &task.Task{}, t)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskByInnerFunc_ReturnsTaskWithAutoGenerateName() {
	d := NewDispatcher()
	t := d.AddTaskByFunc(s.jobSleepSixSeconds)

	assert.Equal(s.T(), "github.com/kihamo/go-workers/dispatcher.jobSleepSixSeconds", t.GetName())
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskByFunc_ReturnsTaskWithAutoGenerateName() {
	d := NewDispatcher()
	t := d.AddTaskByFunc(jobFunc)

	assert.Equal(s.T(), "github.com/kihamo/go-workers/dispatcher.jobFunc", t.GetName())
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskByAnonymousFunc_ReturnsTaskWithAutoGenerateName() {
	d := NewDispatcher()
	t := d.AddTaskByFunc(func(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration) {
		return 1, time.Second
	})

	assert.Equal(s.T(), "github.com/kihamo/go-workers/dispatcher.func", t.GetName())
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskByAnonymousFuncFromNonExportVariable_ReturnsTaskWithAutoGenerateName() {
	d := NewDispatcher()
	t := d.AddTaskByFunc(jobFuncInNonExportVariable)

	assert.Equal(s.T(), "github.com/kihamo/go-workers/dispatcher.func", t.GetName())
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskByAnonymousFuncFromExportVariable_ReturnsTaskWithAutoGenerateName() {
	d := NewDispatcher()
	t := d.AddTaskByFunc(jobFuncInExportVariable)

	assert.Equal(s.T(), "github.com/kihamo/go-workers/dispatcher.func", t.GetName())
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddNameTaskByInnerFuncWithConflictName_ReturnsTaskWithAutoGenerateName() {
	d := NewDispatcher()
	t := d.AddTaskByFunc(s.func1)

	assert.Equal(s.T(), "github.com/kihamo/go-workers/dispatcher.func1", t.GetName())
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddNamedTaskByInnerFunc_ReturnsTask() {
	d := NewDispatcher()
	t := d.AddNamedTaskByFunc("task.test", s.jobSleepSixSeconds)

	assert.IsType(s.T(), &task.Task{}, t)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddNameTaskByInnerFunc_ReturnsTaskWithAutoGenerateName() {
	d := NewDispatcher()
	t := d.AddNamedTaskByFunc("task.test", s.jobSleepSixSeconds)

	assert.Equal(s.T(), "task.test", t.GetName())
}
