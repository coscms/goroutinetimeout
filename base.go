package goroutinetimeout

import (
	"context"
	"errors"
	"log"
	"time"
)

var _ BaseExecutor = (*goBase)(nil)

func New(taskName string, concurrent ...uint) BaseExecutor {
	var n uint
	if len(concurrent) > 0 {
		n = concurrent[0]
	}
	if n < 1 {
		n = 1
	}
	return &goBase{
		taskName:   taskName,
		concurrent: n,
	}
}

type goBase struct {
	taskName   string
	concurrent uint
}

func (g *goBase) makeChan(recv func()) (chan struct{}, func()) {
	done := make(chan struct{}, g.concurrent)
	for i := uint(0); i < g.concurrent; i++ {
		done <- struct{}{}
	}
	return done, func() {
		recv()
		done <- struct{}{}
	}
}

func (g *goBase) TaskName() string {
	return g.taskName
}

func (g *goBase) WithInterval(interval time.Duration, executor func(time.Time)) Executor {
	a := &goWithInterval{
		goBase:   g,
		executor: executor,
		interval: interval,
	}
	return a
}

func (g *goBase) WithIntervalGenerator(intervalGenerator func(time.Time) time.Duration, executor func(time.Time)) Executor {
	a := &goWithIntervalGenerator{
		goBase:            g,
		executor:          executor,
		intervalGenerator: intervalGenerator,
	}
	return a
}

func (g *goBase) WithNextTimeGenerator(nextTimeGenerator func(time.Time) time.Time, executor func(time.Time)) Executor {
	a := &goWithNextTimeGenerator{
		goBase:            g,
		executor:          executor,
		nextTimeGenerator: nextTimeGenerator,
	}
	return a
}

func (g *goBase) SetFuncWithChan(c context.Context, s <-chan interface{}, receiver func(interface{})) (context.Context, context.CancelFunc, func()) {
	ctx, cancel := context.WithCancel(c)
	return ctx, cancel, func() {
		v, y := <-s
		if y {
			receiver(v)
		} else {
			cancel()
		}
	}
}

func (g *goBase) ExecuteWithChan(c context.Context, s <-chan interface{}, f func(interface{})) error {
	ctx, cancel, receiver := g.SetFuncWithChan(c, s, f)
	err := g.Execute(ctx, receiver)
	cancel()
	return err
}

func (g *goBase) Execute(c context.Context, receiver func()) error {
	done, recv := g.makeChan(receiver)
	for {
		select {
		case <-done:
			go recv()
		case <-c.Done():
			log.Println(g.taskName+`:`, c.Err())
			return c.Err()
		}
	}
}

func IsCanceled(err error) bool {
	return errors.Is(err, context.Canceled)
}

func IsTimeout(err error) bool {
	return errors.Is(err, context.DeadlineExceeded)
}
