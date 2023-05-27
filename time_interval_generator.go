package goroutinetimeout

import (
	"context"
	"log"
	"time"
)

var _ Executor = &goWithTimeIntervalGenerator{}

type goWithTimeIntervalGenerator struct {
	*goBase
	intervalFunc      func(time.Time)
	intervalGenerator func(time.Time) (time.Time, time.Duration)
}

func (g *goWithTimeIntervalGenerator) ExecuteWithChan(c context.Context, s <-chan interface{}, f func(interface{})) {
	ctx, cancel := g.SetFuncWithChan(c, s, f)
	g.Execute(ctx)
	cancel()
}

func (g *goWithTimeIntervalGenerator) Execute(c context.Context) {
	done, exec := g.makeChan()
	next, duration := g.intervalGenerator(time.Now())
	t := time.NewTimer(duration)
	defer t.Stop()
	for {
		select {
		case <-done:
			go exec()
		case tm := <-t.C:
			if tm.Before(next) {
				time.Sleep(next.Sub(tm))
			}
			g.intervalFunc(tm)
			next, duration = g.intervalGenerator(tm)
			t.Reset(duration)
		case <-c.Done():
			log.Println(g.taskName+`:`, context.Canceled)
			return
		}
	}
}
