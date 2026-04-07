package goroutinetimeout

import (
	"context"
	"time"
)

type Executor interface {
	TaskName() string
	Execute(c context.Context) error
	ExecuteWithChan(c context.Context, s <-chan interface{}, f func(interface{})) error
}

type BaseExecutor interface {
	Executor
	WithInterval(interval time.Duration, intervalFunc ...func(time.Time)) Executor
	WithIntervalGenerator(intervalGenerator func(time.Time) time.Duration, intervalFunc ...func(time.Time)) Executor
	WithNextTimeGenerator(intervalGenerator func(time.Time) time.Time, intervalFunc ...func(time.Time)) Executor
}
