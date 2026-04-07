package goroutinetimeout

import (
	"context"
	"log"
	"time"
)

var _ Executor = &goWithInterval{}

type goWithInterval struct {
	*goBase
	executor func(time.Time)
	interval time.Duration
}

func (g *goWithInterval) ExecuteWithChan(c context.Context, s <-chan interface{}, receiver func(interface{})) error {
	ctx, cancel, recv := g.SetFuncWithChan(c, s, receiver)
	err := g.Execute(ctx, recv)
	cancel()
	return err
}

func (g *goWithInterval) Execute(c context.Context, receiver func()) error {
	done, recv := g.makeChan(receiver)
	t := time.NewTicker(g.interval)
	defer t.Stop()
	for {
		select {
		case <-done:
			go recv()
		case tm := <-t.C:
			g.executor(tm)
		case <-c.Done():
			log.Println(g.taskName+`:`, c.Err())
			return c.Err()
		}
	}
}
