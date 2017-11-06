package workers // import "github.com/kihamo/go-workers"

//go:generate goimports -w ./

// https://talks.golang.org/2010/io/balance.go
// https://talks.golang.org/2012/waza.slide#53
// http://habrahabr.ru/post/198150/

// go test -bench=Setters -benchmem -run=^a -v ./task
// go test -v ./task -run TestTaskSuite/Test_NewInstance_GetNameReturnsFuncName
