package manager

import (
	"context"
	"fmt"
	"github.com/kihamo/go-workers"
	"github.com/kihamo/go-workers/task"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPop(t *testing.T) {
	m := NewTasksManager()

	item := m.Pull()
	assert.Nil(t, item)

	for i := 0; i < 3; i++ {
		tsk := task.NewFunctionTask(func(context.Context) (interface{}, error) {
			return nil, nil
		})
		tsk.SetName(fmt.Sprintf("task-%d", i))
		item := NewTasksManagerItem(tsk, workers.TaskStatusWait)

		m.Push(item)
	}

	item = m.Pull()
	if assert.NotNil(t, item) {
		assert.IsType(t, &TasksManagerItem{}, item)
		assert.Equal(t, item.(*TasksManagerItem).Task().Name(), "task-0")
		m.Push(item)
		item.(*TasksManagerItem).Lock()
	}

	item = m.Pull()
	if assert.NotNil(t, item) {
		assert.IsType(t, &TasksManagerItem{}, item)
		assert.Equal(t, item.(*TasksManagerItem).Task().Name(), "task-1")
	}

	item = m.Pull()
	if assert.NotNil(t, item) {
		assert.IsType(t, &TasksManagerItem{}, item)
		assert.Equal(t, item.(*TasksManagerItem).Task().Name(), "task-2")
	}

	item = m.Pull()
	assert.Nil(t, item)

	for i, it := range m.GetAll() {
		cast := it.(*TasksManagerItem)

		assert.Equal(t, cast.Task().Name(), fmt.Sprintf("task-%d", i))
		assert.Equal(t, i, cast.Index())
		assert.True(t, cast.IsLocked())
	}
}

func BenchmarkPull(b *testing.B) {
	m := NewTasksManager()

	for i := 0; i < 1000; i++ {
		t := task.NewFunctionTask(func(context.Context) (interface{}, error) {
			return nil, nil
		})
		item := NewTasksManagerItem(t, workers.TaskStatusWait)

		m.Push(item)
	}

	//b.ReportAllocs()
	//b.StopTimer()
	//b.ResetTimer()

	for i := 0; i < b.N; i++ {
		//b.StartTimer()
		item := m.Pull()
		//b.StopTimer()

		if item == nil {
			b.Fail()
		}

		m.Push(item)
	}
}
