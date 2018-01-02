package workers

import (
	"context"
	"fmt"
)

var (
	attemptContextKey = &ContextKey{"attempt"}
)

type ContextKey struct {
	Name string
}

func (k *ContextKey) String() string {
	return "dashboard context value " + k.Name
}

func (k *ContextKey) GoString() string {
	return fmt.Sprintf("%s %#p", k.String(), k)
}

func AttemptFromContext(ctx context.Context) (int64, bool) {
	attempt, ok := ctx.Value(attemptContextKey).(int64)
	return attempt, ok
}

func NewContextWithAttempt(ctx context.Context, attempt int64) context.Context {
	return context.WithValue(ctx, attemptContextKey, attempt)
}
