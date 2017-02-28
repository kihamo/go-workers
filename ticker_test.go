package workers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TickerSuite struct {
	suite.Suite
}

func TestTickerSuite(t *testing.T) {
	suite.Run(t, new(TickerSuite))
}

func (s *TickerSuite) Test_CreateNewInstance_IsStopped() {
	t := NewTicker(time.Second)

	s.False(t.IsStart())
}

func (s *TickerSuite) Test_CreateNewInstanceAndStart_IsStarted() {
	t := NewTicker(time.Second)
	t.Start()

	time.Sleep(time.Second)

	s.True(t.IsStart())
}

func (s *TickerSuite) Test_CreateNewInstanceAndStopAfterStart_IsStopped() {
	t := NewTicker(time.Second)
	t.Start()
	time.Sleep(time.Second)
	t.Stop()
	time.Sleep(time.Second)

	s.False(t.IsStart())
}

func (s *TickerSuite) Test_CreateNewInstanceAndListenChannel_IsAutoStarted() {
	t := NewTicker(time.Second)
	timeout := time.NewTicker(time.Second * 5)

	for {
		select {
		case <-t.C():
			s.True(t.IsStart())
			return
		case <-timeout.C:
			s.Fail("Stop by timeout")
		}
	}
}

func (s *TickerSuite) Test_CreateNewInstanceAndChangeDuration_DurationChanged() {
	t := NewTicker(time.Second * 10)
	time.Sleep(time.Second)
	t.SetDuration(time.Second)
	timeout := time.NewTicker(time.Second * 2)
	i := 0

	for {
		select {
		case <-t.C():
			i++
		case <-timeout.C:
			return
		}
	}

	if i != 1 {
		s.Fail("Fire must equal 1")
	}
}

func (s *TickerSuite) Test_CreateNewInstanceAndChangeDurationAfterStart_DurationChanged() {
	t := NewTicker(time.Second * 10)
	time.Sleep(time.Second)
	t.Start()
	time.Sleep(time.Second)
	t.SetDuration(time.Second)
	timeout := time.NewTicker(time.Second * 2)
	i := 0

	for {
		select {
		case <-t.C():
			i++
		case <-timeout.C:
			return
		}
	}

	if i != 1 {
		s.Fail("Fire must equal 1")
	}
}
