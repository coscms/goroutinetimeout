package goroutinetimeout

import (
	"context"
	"log"
	"time"
)

var (
	_ Executor = &GoWithInterval{}
	_ Executor = &GoWithIntervalGenerator{}
)

type GoWithInterval struct {
	*GoBase
	intervalFunc func(time.Time)
	interval     time.Duration
}

type GoWithIntervalGenerator struct {
	*GoBase
	intervalFunc      func(time.Time)
	intervalGenerator func(time.Time) time.Duration
}

func (g *GoBase) WithInterval(intervalFunc func(time.Time), interval time.Duration) Executor {
	return &GoWithInterval{
		GoBase:       g,
		intervalFunc: intervalFunc,
		interval:     interval,
	}
}

func (g *GoBase) WithIntervalGenerator(intervalFunc func(time.Time), intervalGenerator func(time.Time) time.Duration) Executor {
	return &GoWithIntervalGenerator{
		GoBase:            g,
		intervalFunc:      intervalFunc,
		intervalGenerator: intervalGenerator,
	}
}

func (g *GoWithIntervalGenerator) ExecuteWithChan(c context.Context, s <-chan interface{}, f func(interface{})) {
	ctx, cancel := g.SetFuncWithChan(c, s, f)
	g.Execute(ctx)
	cancel()
}

func (g *GoWithIntervalGenerator) Execute(c context.Context) {
	done, exec := g.makeChan()
	t := time.NewTimer(g.intervalGenerator(time.Now()))
	defer t.Stop()
	for {
		select {
		case <-done:
			go exec()
		case tm := <-t.C:
			g.intervalFunc(tm)
			t.Reset(g.intervalGenerator(tm))
		case <-c.Done():
			log.Println(g.taskName+`:`, context.Canceled)
			return
		}
	}
}

func (g *GoWithInterval) ExecuteWithChan(c context.Context, s <-chan interface{}, f func(interface{})) {
	ctx, cancel := g.SetFuncWithChan(c, s, f)
	g.Execute(ctx)
	cancel()
}

func (g *GoWithInterval) Execute(c context.Context) {
	done, exec := g.makeChan()
	t := time.NewTicker(g.interval)
	defer t.Stop()
	for {
		select {
		case <-done:
			go exec()
		case tm := <-t.C:
			g.intervalFunc(tm)
		case <-c.Done():
			log.Println(g.taskName+`:`, context.Canceled)
			return
		}
	}
}
