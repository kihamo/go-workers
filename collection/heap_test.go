package collection

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type HeapSuite struct {
	suite.Suite
}

func TestHeapSuite(t *testing.T) {
	suite.Run(t, new(HeapSuite))
}
