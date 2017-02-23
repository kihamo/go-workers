package dispatcher

import (
	"testing"
	"time"
)

func runDispatcherByWorkersCountAndTasksCount(b *testing.B, workersCount int, tasksCount int) {
	b.StopTimer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		quit := make(chan bool, 1)

		listener := NewDefaultListener()

		d := NewDispatcher()
		d.AddListener(listener)

		for w := 1; w <= workersCount; w++ {
			d.AddWorker()
		}

		go d.Run()

		go func() {
			finished := 0

			for {
				select {
				case <-listener.TaskDone:
					finished++

					if finished == tasksCount {
						d.Kill()
						quit <- true
					}
				}
			}
		}()

		for d.GetStatus() != DispatcherStatusProcess {
		}

		f := func(_ int64, _ chan bool, _ ...interface{}) (int64, time.Duration, interface{}, error) {
			return 0, 0, nil, nil
		}

		b.StartTimer()

		for t := 1; t <= tasksCount; t++ {
			d.AddTaskByFunc(f)
		}

		<-quit
		b.StopTimer()
	}
}

func Benchmark100Workers10Tasks(b *testing.B) {
	runDispatcherByWorkersCountAndTasksCount(b, 100, 10)
}

func Benchmark1000Workers1000Tasks(b *testing.B) {
	runDispatcherByWorkersCountAndTasksCount(b, 1000, 1000)
}
