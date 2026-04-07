package goroutinetimeout

import (
	"context"
	"log"
	"time"
)

var _ Executor = (*goWithIntervalGenerator)(nil)

type goWithIntervalGenerator struct {
	*goBase
	executor          func(time.Time)
	intervalGenerator func(time.Time) time.Duration
}

func (g *goWithIntervalGenerator) ExecuteWithChan(c context.Context, s <-chan interface{}, receiver func(interface{})) error {
	ctx, cancel, recv := g.SetFuncWithChan(c, s, receiver)
	err := g.Execute(ctx, recv)
	cancel()
	return err
}

func (g *goWithIntervalGenerator) Execute(c context.Context, receiver func()) error {
	done, recv := g.makeChan(receiver)
	t := time.NewTimer(g.intervalGenerator(time.Now()))
	defer t.Stop()
	for {
		select {
		case <-done:
			go recv()
		case tm := <-t.C:
			g.executor(tm)
			t.Reset(g.intervalGenerator(time.Now()))
		case <-c.Done():
			log.Println(g.taskName+`:`, c.Err())
			return c.Err()
		}
	}
}
