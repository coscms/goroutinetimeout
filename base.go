package goroutinetimeout

import (
	"context"
	"errors"
	"log"
	"time"
)

var _ BaseExecutor = (*goBase)(nil)

func New(taskName string, f func(), concurrent ...uint) BaseExecutor {
	var n uint
	if len(concurrent) > 0 {
		n = concurrent[0]
	}
	if n < 1 {
		n = 1
	}
	return &goBase{
		taskName:   taskName,
		goFunc:     f,
		concurrent: n,
	}
}

type goBase struct {
	taskName   string
	goFunc     func()
	concurrent uint
}

func (g *goBase) makeChan() (chan struct{}, func()) {
	done := make(chan struct{}, g.concurrent)
	for i := uint(0); i < g.concurrent; i++ {
		done <- struct{}{}
	}
	return done, func() {
		g.goFunc()
		done <- struct{}{}
	}
}

func (g *goBase) TaskName() string {
	return g.taskName
}

func (g *goBase) WithInterval(interval time.Duration, intervalFunc ...func(time.Time)) Executor {
	a := &goWithInterval{
		goBase:   g,
		interval: interval,
	}
	if len(intervalFunc) > 0 {
		a.intervalFunc = intervalFunc[0]
	}
	return a
}

func (g *goBase) WithIntervalGenerator(intervalGenerator func(time.Time) time.Duration, intervalFunc ...func(time.Time)) Executor {
	a := &goWithIntervalGenerator{
		goBase:            g,
		intervalGenerator: intervalGenerator,
	}
	if len(intervalFunc) > 0 {
		a.intervalFunc = intervalFunc[0]
	}
	return a
}

func (g *goBase) WithNextTimeGenerator(nextTimeGenerator func(time.Time) time.Time, intervalFunc ...func(time.Time)) Executor {
	a := &goWithNextTimeGenerator{
		goBase:            g,
		nextTimeGenerator: nextTimeGenerator,
	}
	if len(intervalFunc) > 0 {
		a.intervalFunc = intervalFunc[0]
	}
	return a
}

func (g *goBase) SetFuncWithChan(c context.Context, s <-chan interface{}, f func(interface{})) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(c)
	g.goFunc = func() {
		v, y := <-s
		if y {
			f(v)
		} else {
			cancel()
		}
	}
	return ctx, cancel
}

func (g *goBase) ExecuteWithChan(c context.Context, s <-chan interface{}, f func(interface{})) error {
	ctx, cancel := g.SetFuncWithChan(c, s, f)
	err := g.Execute(ctx)
	cancel()
	return err
}

func (g *goBase) Execute(c context.Context) error {
	done, recv := g.makeChan()
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
