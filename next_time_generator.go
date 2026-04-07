package goroutinetimeout

import (
	"context"
	"log"
	"time"
)

var _ Executor = (*goWithNextTimeGenerator)(nil)

type goWithNextTimeGenerator struct {
	*goBase
	executor          func(time.Time)
	nextTimeGenerator func(time.Time) time.Time
}

func (g *goWithNextTimeGenerator) ExecuteWithChan(c context.Context, s <-chan interface{}, f func(interface{})) error {
	ctx, cancel := g.SetFuncWithChan(c, s, f)
	err := g.Execute(ctx)
	cancel()
	return err
}

func (g *goWithNextTimeGenerator) Execute(c context.Context) error {
	done, recv := g.makeChan()
	next := g.nextTimeGenerator(time.Now())
	duration := time.Until(next)
	t := time.NewTimer(duration)
	defer t.Stop()
	for {
		select {
		case <-done:
			go recv()
		case tm := <-t.C:
			if tm.Before(next) {
				time.Sleep(next.Sub(tm))
			}
			g.executor(tm)
			next = g.nextTimeGenerator(time.Now())
			duration = time.Until(next)
			t.Reset(duration)
		case <-c.Done():
			log.Println(g.taskName+`:`, c.Err())
			return c.Err()
		}
	}
}
