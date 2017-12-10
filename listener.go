package workers

import (
	"time"
)

type Listener func(time.Time, ...interface{})

func (l Listener) String() string {
	return FunctionName(l)
}
