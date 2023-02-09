package goroutinetimeout

import (
	"context"
	"log"
)

var (
	_ BaseExecutor = &GoBase{}
)

func New(taskName string, f func(), concurrent ...uint) Executor {
	var n uint
	if len(concurrent) > 0 {
		n = concurrent[0]
	}
	if n < 1 {
		n = 1
	}
	return &GoBase{
		taskName:   taskName,
		goFunc:     f,
		concurrent: n,
	}
}

type GoBase struct {
	taskName   string
	goFunc     func()
	concurrent uint
}

func (g *GoBase) makeChan() (chan struct{}, func()) {
	done := make(chan struct{}, g.concurrent)
	for i := uint(0); i < g.concurrent; i++ {
		done <- struct{}{}
	}
	return done, func() {
		g.goFunc()
		done <- struct{}{}
	}
}

func (g *GoBase) TaskName() string {
	return g.taskName
}

func (g *GoBase) setFunc(f func()) {
	g.goFunc = f
}

func (g *GoBase) SetFuncWithChan(c context.Context, s <-chan interface{}, f func(interface{})) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(c)
	g.setFunc(func() {
		v, y := <-s
		if y {
			f(v)
		} else {
			cancel()
		}
	})
	return ctx, cancel
}

func (g *GoBase) ExecuteWithChan(c context.Context, s <-chan interface{}, f func(interface{})) {
	ctx, cancel := g.SetFuncWithChan(c, s, f)
	g.Execute(ctx)
	cancel()
}

func (g *GoBase) Execute(c context.Context) {
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
