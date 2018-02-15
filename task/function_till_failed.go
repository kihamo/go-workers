package task

import (
	"context"

	"github.com/kihamo/go-workers"
)

type FunctionTillFailedTask struct {
	FunctionTask
}

func NewFunctionTillFailedTask(function func(context.Context) (interface{}, error)) *FunctionTillFailedTask {
	t := &FunctionTillFailedTask{}
	t.function = function
	t.FunctionTask.Init()

	return t
}

func (t *FunctionTillFailedTask) Run(ctx context.Context) (interface{}, error) {
	result, err := t.function(ctx)

	if err != nil {
		attempt, ok := workers.AttemptFromContext(ctx)
		if ok {
			t.SetRepeats(attempt)
		}
	}

	return result, err
}
