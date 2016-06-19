package collection

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type TasksSuite struct {
	suite.Suite
}

func TestTasksSuite(t *testing.T) {
	suite.Run(t, new(TasksSuite))
}
