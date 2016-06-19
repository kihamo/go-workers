package dispatcher

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type DispatcherSuite struct {
	suite.Suite
}

func TestDispatcherSuite(t *testing.T) {
	suite.Run(t, new(DispatcherSuite))
}
