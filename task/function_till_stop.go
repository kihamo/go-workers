package task

import (
	"context"

	"github.com/kihamo/go-workers"
)

type FunctionTillStopTask struct {
	BaseTask

	function func(context.Context) (interface{}, error, bool)
}

func NewFunctionTillStopTask(function func(context.Context) (interface{}, error, bool)) *FunctionTillStopTask {
	t := &FunctionTillStopTask{
		function: function,
	}
	t.BaseTask.Init()

	return t
}

func (t *FunctionTillStopTask) Run(ctx context.Context) (interface{}, error) {
	result, err, stop := t.function(ctx)

	if stop {
		attempt, ok := workers.AttemptFromContext(ctx)
		if ok {
			t.SetRepeats(attempt)
		}
	}

	return result, err
}
