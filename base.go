package goroutinetimeout

import (
	"context"
	"log"
	"time"
)

var _ BaseExecutor = &goBase{}

func New(taskName string, f func(), concurrent ...uint) Executor {
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

func (g *goBase) WithInterval(intervalFunc func(time.Time), interval time.Duration) Executor {
	return &goWithInterval{
		goBase:       g,
		intervalFunc: intervalFunc,
		interval:     interval,
	}
}

func (g *goBase) WithIntervalGenerator(intervalFunc func(time.Time), intervalGenerator func(time.Time) time.Duration) Executor {
	return &goWithIntervalGenerator{
		goBase:            g,
		intervalFunc:      intervalFunc,
		intervalGenerator: intervalGenerator,
	}
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

func (g *goBase) ExecuteWithChan(c context.Context, s <-chan interface{}, f func(interface{})) {
	ctx, cancel := g.SetFuncWithChan(c, s, f)
	g.Execute(ctx)
	cancel()
}

func (g *goBase) Execute(c context.Context) {
	done, exec := g.makeChan()
	for {
		select {
		case <-done:
			go exec()
		case <-c.Done():
			log.Println(g.taskName+`:`, context.Canceled)
			return
		}
	}
}
