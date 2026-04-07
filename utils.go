package goroutinetimeout

import (
	"context"
	"time"
)

func NextTimeExecute(c context.Context, executor func(time.Time) error, nextTimeGenerator func(time.Time) time.Time) error {
	next := nextTimeGenerator(time.Now())
	duration := time.Until(next)
	t := time.NewTimer(duration)
	defer t.Stop()
	for {
		select {
		case tm := <-t.C:
			if tm.Before(next) {
				time.Sleep(next.Sub(tm))
			}
			err := executor(tm)
			if err != nil {
				return err
			}
			next = nextTimeGenerator(time.Now())
			duration = time.Until(next)
			t.Reset(duration)
		case <-c.Done():
			return c.Err()
		}
	}
}

func IntervalExecute(c context.Context, executor func(time.Time) error, intervalGenerator func(time.Time) time.Duration) error {
	t := time.NewTimer(intervalGenerator(time.Now()))
	defer t.Stop()
	for {
		select {
		case tm := <-t.C:
			err := executor(tm)
			if err != nil {
				return err
			}
			t.Reset(intervalGenerator(time.Now()))
		case <-c.Done():
			return c.Err()
		}
	}
}
