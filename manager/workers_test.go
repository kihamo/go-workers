package manager

import (
	"errors"
	"testing"

	"github.com/kihamo/go-workers"
	"github.com/kihamo/go-workers/worker"
	"github.com/stretchr/testify/suite"
)

type WorkersSuite struct {
	suite.Suite
}

func TestWorkersSuite(t *testing.T) {
	suite.Run(t, new(WorkersSuite))
}

func (s *WorkersSuite) Test_CreateNewInstance_PullReturnsNil() {
	m := NewWorkersManager()

	s.Nil(m.Pull())
}

func (s *WorkersSuite) Test_CreateNewInstance_GetAllZeroLenOfItems() {
	m := NewWorkersManager()

	s.Zero(len(m.GetAll()))
}

func (s *WorkersSuite) Test_CreateNewInstance_CheckReturnsFalse() {
	m := NewWorkersManager()

	s.False(m.Check())
}

func (s *WorkersSuite) Test_CreateNewInstanceAndPushNil_GetAllZeroLenOfItems() {
	m := NewWorkersManager()

	m.Push(nil)

	s.Zero(len(m.GetAll()))
}

func (s *WorkersSuite) Test_CreateNewInstanceAndPushNil_CheckReturnsFalse() {
	m := NewWorkersManager()

	m.Push(nil)

	s.False(m.Check())
}

func (s *WorkersSuite) Test_CreateNewInstanceAndPushNil_ReturnsNotNil() {
	m := NewWorkersManager()

	r := m.Push(nil)

	s.NotNil(r)
}

func (s *WorkersSuite) Test_CreateNewInstanceAndPushNil_ReturnsError() {
	m := NewWorkersManager()

	r := m.Push(nil)

	s.Equal(errors.New("Worker can't be nil"), r)
}

func (s *WorkersSuite) Test_CreateNewInstanceAndPushItem_ReturnsNil() {
	m := NewWorkersManager()
	w := worker.NewSimpleWorker()
	st := workers.WorkerStatusWait
	wi := NewWorkersManagerItem(w, st)

	r := m.Push(wi)

	s.Nil(r)
}

func (s *WorkersSuite) Test_CreateNewInstanceAndPushItem_PullReturnsNotNil() {
	m := NewWorkersManager()
	w := worker.NewSimpleWorker()
	st := workers.WorkerStatusWait
	wi := NewWorkersManagerItem(w, st)

	m.Push(wi)

	s.NotNil(m.Pull())
}

func (s *WorkersSuite) Test_CreateNewInstanceAndPushItem_PullReturnsPushedItem() {
	m := NewWorkersManager()
	w := worker.NewSimpleWorker()
	st := workers.WorkerStatusWait
	wi := NewWorkersManagerItem(w, st)

	m.Push(wi)

	s.Equal(wi, m.Pull())
}

func (s *WorkersSuite) Test_CreateNewInstanceAndPushItem_GetAllReturnsOneItem() {
	m := NewWorkersManager()
	w := worker.NewSimpleWorker()
	st := workers.WorkerStatusWait
	wi := NewWorkersManagerItem(w, st)

	m.Push(wi)

	s.Equal(1, len(m.GetAll()))
}

func (s *WorkersSuite) Test_CreateNewInstanceAndPushItem_CheckReturnsTrue() {
	m := NewWorkersManager()
	w := worker.NewSimpleWorker()
	st := workers.WorkerStatusWait
	wi := NewWorkersManagerItem(w, st)

	m.Push(wi)

	s.True(m.Check())
}

func (s *WorkersSuite) Test_CreateNewInstanceAndPushItem_WorkerItemIsNotLocked() {
	m := NewWorkersManager()
	w := worker.NewSimpleWorker()
	st := workers.WorkerStatusWait
	wi := NewWorkersManagerItem(w, st)

	m.Push(wi)

	s.False(wi.IsLocked())
}

func (s *WorkersSuite) Test_CreateNewInstanceAndPushItem_GetAllReturnsWorkerItemIsNotLocked() {
	m := NewWorkersManager()
	w := worker.NewSimpleWorker()
	st := workers.WorkerStatusWait
	wi := NewWorkersManagerItem(w, st)

	m.Push(wi)
	a := m.GetAll()

	for _, aw := range a {
		s.False(aw.IsLocked())
	}
}

func (s *WorkersSuite) Test_CreateNewInstanceAndPushItemAndPull_GetAllReturnsOneItem() {
	m := NewWorkersManager()
	w := worker.NewSimpleWorker()
	st := workers.WorkerStatusWait
	wi := NewWorkersManagerItem(w, st)

	m.Push(wi)
	m.Pull()

	s.Equal(1, len(m.GetAll()))
}

func (s *WorkersSuite) Test_CreateNewInstanceAndPushItemAndPull_WorkerItemIsLocked() {
	m := NewWorkersManager()
	w := worker.NewSimpleWorker()
	st := workers.WorkerStatusWait
	wi := NewWorkersManagerItem(w, st)

	m.Push(wi)
	m.Pull()

	s.True(wi.IsLocked())
}

func (s *WorkersSuite) Test_CreateNewInstanceAndPushItemAndPull_PullReturnsWorkerItemIsLocked() {
	m := NewWorkersManager()
	w := worker.NewSimpleWorker()
	st := workers.WorkerStatusWait
	wi := NewWorkersManagerItem(w, st)

	m.Push(wi)
	pw := m.Pull()

	s.True(pw.IsLocked())
}

func (s *WorkersSuite) Test_CreateNewInstanceAndPushItemAndPull_GetAllReturnsWorkerItemIsLocked() {
	m := NewWorkersManager()
	w := worker.NewSimpleWorker()
	st := workers.WorkerStatusWait
	wi := NewWorkersManagerItem(w, st)

	m.Push(wi)
	m.Pull()
	a := m.GetAll()

	for _, aw := range a {
		s.True(aw.IsLocked())
	}
}
