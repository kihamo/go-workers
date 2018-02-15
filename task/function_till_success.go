package task

import (
	"context"

	"github.com/kihamo/go-workers"
)

type FunctionTillSuccessTask struct {
	FunctionTask
}

func NewFunctionTillSuccessTask(function func(context.Context) (interface{}, error)) *FunctionTillSuccessTask {
	t := &FunctionTillSuccessTask{}
	t.function = function
	t.FunctionTask.Init()

	return t
}

func (t *FunctionTillSuccessTask) Run(ctx context.Context) (interface{}, error) {
	result, err := t.function(ctx)

	if err == nil {
		attempt, ok := workers.AttemptFromContext(ctx)
		if ok {
			t.SetRepeats(attempt)
		}
	}

	return result, err
}
