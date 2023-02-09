package goroutinetimeout

import (
	"context"
	"time"
)

type Executor interface {
	TaskName() string
	Execute(c context.Context)
	ExecuteWithChan(c context.Context, s <-chan interface{}, f func(interface{}))
}

type BaseExecutor interface {
	Executor
	WithInterval(intervalFunc func(time.Time), interval time.Duration) Executor
	WithIntervalGenerator(intervalFunc func(time.Time), intervalGenerator func(time.Time) time.Duration) Executor
}
