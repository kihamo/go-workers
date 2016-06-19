package workers // import "github.com/kihamo/go-workers"

//go:generate goimports -w ./
//go:generate mockery -name=Tasker -dir=task -case=underscore
//go:generate mockery -name=Worker -dir=worker -case=underscore

// https://talks.golang.org/2010/io/balance.go
// https://talks.golang.org/2012/waza.slide#53
// http://habrahabr.ru/post/198150/
