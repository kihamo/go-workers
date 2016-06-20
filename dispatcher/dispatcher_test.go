package dispatcher

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DispatcherSuite struct {
	suite.Suite

	dispatcher *Dispatcher
}

func TestDispatcherSuite(t *testing.T) {
	suite.Run(t, new(DispatcherSuite))
}

func (s *DispatcherSuite) SetupTest() {
	s.dispatcher = NewDispatcher()
}

func (s *DispatcherSuite) TearDownTest() {
	s.dispatcher.Kill()
}

func (s *DispatcherSuite) Test_FirstRun_ReturnEmptyError() {
	var err error

	go func() {
		err = s.dispatcher.Run()
	}()
	time.Sleep(time.Second)

	assert.Nil(s.T(), err)
}

func (s *DispatcherSuite) Test_TwiceRun_ReturnErrorForSecondRun() {
	go s.dispatcher.Run()
	time.Sleep(time.Second)

	err := s.dispatcher.Run()

	assert.Equal(s.T(), err, errors.New("Dispatcher is running"))
}

func (s *DispatcherSuite) Test_NewInstance_ReturnsStatusWait() {
	assert.Equal(s.T(), s.dispatcher.GetStatus(), DispatcherStatusWait)
}

func (s *DispatcherSuite) Test_Run_ReturnsStatusProcess() {
	go s.dispatcher.Run()

	time.Sleep(time.Second)

	assert.Equal(s.T(), s.dispatcher.GetStatus(), DispatcherStatusProcess)
	s.dispatcher.Kill()
}

func (s *DispatcherSuite) Test_IsRunningAndKill_ReturnEmptyError() {
	go s.dispatcher.Run()
	time.Sleep(time.Second)

	err := s.dispatcher.Kill()

	assert.Nil(s.T(), err)
}

func (s *DispatcherSuite) Test_NewInstanceAndKill_ReturnError() {
	err := s.dispatcher.Kill()

	assert.Equal(s.T(), err, errors.New("Dispatcher isn't running"))
}

func (s *DispatcherSuite) Test_IsRunningTwiceKill_ReturnErrorForSecondKill() {
	go s.dispatcher.Run()
	time.Sleep(time.Second)
	s.dispatcher.Kill()
	time.Sleep(time.Second)

	err := s.dispatcher.Kill()

	assert.Equal(s.T(), err, errors.New("Dispatcher isn't running"))
}

func (s *DispatcherSuite) Test_Kill_ReturnsStatusWait() {
	go s.dispatcher.Run()

	time.Sleep(time.Second)

	s.dispatcher.Kill()

	time.Sleep(time.Second)

	assert.Equal(s.T(), s.dispatcher.GetStatus(), DispatcherStatusWait)
	s.dispatcher.Kill()
}

func (s *DispatcherSuite) Test_NewInstance_ReturnsZeroSIzeOfWorkersList() {
	assert.Equal(s.T(), s.dispatcher.GetWorkers().Len(), 0)
}

func (s *DispatcherSuite) Test_AddOneWorker_ReturnsOneSizeOfWorkersList() {
	s.dispatcher.AddWorker()

	assert.Equal(s.T(), s.dispatcher.GetWorkers().Len(), 1)
	s.dispatcher.Kill()
}

func (s *DispatcherSuite) Test_NewInstance_ReturnsZeroSIzeOfTasksList() {
	assert.Equal(s.T(), s.dispatcher.GetTasks().Len(), 0)
}
