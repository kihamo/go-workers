package task

import (
	"context"

	"github.com/kihamo/go-workers"
)

type FunctionTask struct {
	BaseTask

	function func() (interface{}, error)
}

func NewFunctionTask(function func() (interface{}, error)) *FunctionTask {
	t := &FunctionTask{
		function: function,
	}
	t.BaseTask.Init()

	return t
}

func (t *FunctionTask) Run(context.Context) (interface{}, error) {
	return t.function()
}

func (t *FunctionTask) Name() string {
	n := t.BaseTask.Name()

	if n == "" {
		return workers.FunctionName(t.function)
	}

	return n
}
