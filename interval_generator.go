package goroutinetimeout

import (
	"context"
	"log"
	"time"
)

var _ Executor = &goWithIntervalGenerator{}

type goWithIntervalGenerator struct {
	*goBase
	intervalFunc      func(time.Time)
	intervalGenerator func(time.Time) time.Duration
}

func (g *goWithIntervalGenerator) ExecuteWithChan(c context.Context, s <-chan interface{}, f func(interface{})) error {
	ctx, cancel := g.SetFuncWithChan(c, s, f)
	err := g.Execute(ctx)
	cancel()
	return err
}

func (g *goWithIntervalGenerator) Execute(c context.Context) error {
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
			log.Println(g.taskName+`:`, c.Err())
			return c.Err()
		}
	}
}
