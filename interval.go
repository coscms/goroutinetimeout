package goroutinetimeout

import (
	"context"
	"log"
	"time"
)

var _ Executor = &goWithInterval{}

type goWithInterval struct {
	*goBase
	intervalFunc func(time.Time)
	interval     time.Duration
}

func (g *goWithInterval) ExecuteWithChan(c context.Context, s <-chan interface{}, f func(interface{})) error {
	ctx, cancel := g.SetFuncWithChan(c, s, f)
	err := g.Execute(ctx)
	cancel()
	return err
}

func (g *goWithInterval) Execute(c context.Context) error {
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
			return context.Canceled
		}
	}
}
