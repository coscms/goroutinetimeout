package goroutinetimeout

import (
	"context"
	"time"
)

type Executor interface {
	TaskName() string
	Execute(c context.Context) error
	ExecuteWithChan(c context.Context, s <-chan interface{}, receiver func(interface{})) error
}

type BaseExecutor interface {
	Executor
	WithInterval(interval time.Duration, executor func(time.Time)) Executor
	WithIntervalGenerator(intervalGenerator func(time.Time) time.Duration, executor func(time.Time)) Executor
	WithNextTimeGenerator(intervalGenerator func(time.Time) time.Time, executor func(time.Time)) Executor
}
