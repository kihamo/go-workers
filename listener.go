package workers

import (
	"time"
)

type Listener func(time.Time, ...interface{})
