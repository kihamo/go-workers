package collection

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type WorkersSuite struct {
	suite.Suite
}

func TestWorkersSuite(t *testing.T) {
	suite.Run(t, new(WorkersSuite))
}
