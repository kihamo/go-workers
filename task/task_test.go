package task

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type TaskSuite struct {
	suite.Suite
}

func TestDispatcherSuite(t *testing.T) {
	suite.Run(t, new(TaskSuite))
}
