package workers // import "github.com/kihamo/go-workers"

import (
	"github.com/pivotal-golang/clock"
)

//go:generate goimports -w ./

// https://talks.golang.org/2010/io/balance.go
// https://talks.golang.org/2012/waza.slide#53
// http://habrahabr.ru/post/198150/

var Clock = clock.NewClock()
