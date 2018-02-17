package workers // import "github.com/kihamo/go-workers"

//go:generate goimports -w ./
//go:generate stringer -type=DispatcherStatus -trimprefix=DispatcherStatus -output dispatcher_status_string.go
//go:generate stringer -type=TaskStatus -trimprefix=TaskStatus -output task_status_string.go
//go:generate stringer -type=WorkerStatus -trimprefix=WorkerStatus -output worker_status_string.go

// https://talks.golang.org/2010/io/balance.go
// https://talks.golang.org/2012/waza.slide#53
// http://habrahabr.ru/post/198150/

// go test -bench=Setters -benchmem -run=^a -v ./task
// go test -v ./task -run TestTaskSuite/Test_NewInstance_GetNameReturnsFuncName
