package dispatcher

import (
	"errors"
	"testing"
	"time"

	"github.com/kihamo/go-workers/task"
	"github.com/kihamo/go-workers/worker"
	"github.com/pivotal-golang/clock/fakeclock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DispatcherSuite struct {
	suite.Suite

	clockTime time.Time
}

var (
	jobFuncInNonExportVariable = func(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration, error) {
		return 1, time.Second, nil
	}
	jobFuncInExportVariable = func(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration, error) {
		return 1, time.Second, nil
	}
)

func jobOuter(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration, error) {
	return 1, time.Second, nil
}

func TestDispatcherSuite(t *testing.T) {
	suite.Run(t, new(DispatcherSuite))
}

func (s *DispatcherSuite) SetupSuite() {
	s.clockTime = time.Date(2016, 6, 5, 4, 3, 2, 1, time.UTC)
}

func (s *DispatcherSuite) func1(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration, error) {
	return 1, time.Second, nil
}

func (s *DispatcherSuite) jobInner(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration, error) {
	return 1, time.Second, nil
}

func (s *DispatcherSuite) Test_FirstRun_ReturnEmptyError() {
	var err error

	d := NewDispatcher()
	go func() {
		err = d.Run()
	}()

	assert.Nil(s.T(), err)
	d.Kill()
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
	d.AddTaskByFunc(s.func1)

	assert.Equal(s.T(), d.GetTasks().Len(), 0)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskAndRun_ReturnsZeroSizeOfTasksListAfterSuccessExecute() {
	d := NewDispatcher()
	w := d.AddWorker()
	c := fakeclock.NewFakeClock(s.clockTime)
	t := task.NewTaskWithClock(c, func(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration, error) {
		c.Sleep(time.Second * 6)
		return 1, time.Second, nil
	})

	d.AddTask(t)
	go d.Run()
	for w.GetStatus() != worker.WorkerStatusBusy {
	}

	c.WaitForWatcherAndIncrement(time.Second * 6)
	for d.GetTasks().Len() != 0 {
	}

	assert.Equal(s.T(), d.GetTasks().Len(), 0)

	d.Kill()
}

func (s *DispatcherSuite) Test_IsRunningAndAddTask_ReturnOneSizeOfTasksList() {
	d := NewDispatcher()
	go d.Run()
	for d.GetStatus() != DispatcherStatusProcess {
	}

	d.AddWorker()
	d.AddTaskByFunc(s.func1)

	assert.Equal(s.T(), d.GetTasks().Len(), 1)
	d.Kill()
}

func (s *DispatcherSuite) Test_IsRunningAndAddTaskWithDuration_ReturnsZeroSizeOfTasksListBeforeExpirationDuration() {
	c := fakeclock.NewFakeClock(s.clockTime)
	d := NewDispatcherWithClock(c)
	go d.Run()
	for d.GetStatus() != DispatcherStatusProcess {
	}

	t := task.NewTask(s.func1)
	t.SetDuration(time.Second * 2)
	d.AddTask(t)

	assert.Equal(s.T(), d.GetTasks().Len(), 0)

	d.Kill()
}

func (s *DispatcherSuite) Test_IsRunningAndAddTaskWithDuration_ReturnsOneSizeOfTasksListAfterExpirationDuration() {
	c := fakeclock.NewFakeClock(s.clockTime)
	d := NewDispatcherWithClock(c)
	go d.Run()
	for d.GetStatus() != DispatcherStatusProcess {
	}

	w := d.AddWorker()
	for w.GetStatus() != worker.WorkerStatusWait {
	}

	t := task.NewTask(s.func1)
	t.SetDuration(time.Second * 2)
	d.AddTask(t)

	assert.Equal(s.T(), d.GetTasks().Len(), 0)
	c.IncrementBySeconds(2)
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
	d.AddTaskByFunc(s.func1)

	assert.Equal(s.T(), d.GetWaitTasks().Len(), 1)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskAndRun_ReturnsZeroSizeOfWaitTasksList() {
	d := NewDispatcher()
	w := d.AddWorker()
	d.AddTaskByFunc(s.func1)
	go d.Run()
	for w.GetStatus() != worker.WorkerStatusBusy {
	}

	assert.Equal(s.T(), d.GetWaitTasks().Len(), 0)
	d.Kill()
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskByInnerFunc_ReturnsTask() {
	t := NewDispatcher().AddTaskByFunc(s.func1)

	assert.IsType(s.T(), &task.Task{}, t)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskByInnerFunc_ReturnsTaskWithAutoGenerateName() {
	t := NewDispatcher().AddTaskByFunc(s.jobInner)

	assert.Equal(s.T(), "github.com/kihamo/go-workers/dispatcher.jobInner", t.GetName())
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskByFunc_ReturnsTaskWithAutoGenerateName() {
	t := NewDispatcher().AddTaskByFunc(jobOuter)

	assert.Equal(s.T(), "github.com/kihamo/go-workers/dispatcher.jobOuter", t.GetName())
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskByAnonymousFunc_ReturnsTaskWithAutoGenerateName() {
	t := NewDispatcher().AddTaskByFunc(func(attempts int64, quit chan bool, args ...interface{}) (int64, time.Duration, error) {
		return 1, time.Second, nil
	})

	assert.Equal(s.T(), "github.com/kihamo/go-workers/dispatcher.func", t.GetName())
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskByAnonymousFuncFromNonExportVariable_ReturnsTaskWithAutoGenerateName() {
	t := NewDispatcher().AddTaskByFunc(jobFuncInNonExportVariable)

	assert.Equal(s.T(), "github.com/kihamo/go-workers/dispatcher.func", t.GetName())
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddTaskByAnonymousFuncFromExportVariable_ReturnsTaskWithAutoGenerateName() {
	t := NewDispatcher().AddTaskByFunc(jobFuncInExportVariable)

	assert.Equal(s.T(), "github.com/kihamo/go-workers/dispatcher.func", t.GetName())
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddNameTaskByInnerFuncWithConflictName_ReturnsTaskWithAutoGenerateName() {
	t := NewDispatcher().AddTaskByFunc(s.func1)

	assert.Equal(s.T(), "github.com/kihamo/go-workers/dispatcher.func1", t.GetName())
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddNamedTaskByInnerFunc_ReturnsTask() {
	t := NewDispatcher().AddNamedTaskByFunc("task.test", s.func1)

	assert.IsType(s.T(), &task.Task{}, t)
}

func (s *DispatcherSuite) Test_CreateNewInstanceAndAddNameTaskByInnerFunc_ReturnsTaskWithAutoGenerateName() {
	t := NewDispatcher().AddNamedTaskByFunc("task.test", s.func1)

	assert.Equal(s.T(), "task.test", t.GetName())
}
