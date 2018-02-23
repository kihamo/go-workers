package listener

import (
	"context"
	"time"

	"github.com/kihamo/go-workers"
)

type FunctionListener struct {
	BaseListener

	function func(context.Context, workers.Event, time.Time, ...interface{})
}

func NewFunctionListener(function func(context.Context, workers.Event, time.Time, ...interface{})) *FunctionListener {
	t := &FunctionListener{
		function: function,
	}
	t.BaseListener.Init()

	return t
}

func (l *FunctionListener) Run(ctx context.Context, event workers.Event, t time.Time, args ...interface{}) {
	l.function(ctx, event, t, args...)
}

func (l *FunctionListener) Name() string {
	n := l.BaseListener.Name()

	if n == "" {
		return workers.FunctionName(l.function)
	}

	return n
}
