package goroutinetimeout

import (
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestChan(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		time.Sleep(time.Second)
	}()
	queue := make(chan interface{}, 5)
	push := func() {
		for i := 0; i < 10; i++ {
			queue <- i
		}
		close(queue)
	}
	g := New(`TestChan`, nil, 2)
	f := func(v interface{}) {
		i := v.(int)
		time.Sleep(2 * time.Second)
		t.Logf(`Execute.%d`, i)
	}

	// 1.
	// go g.ExecuteWithChan(ctx, queue, f)
	// push()
	// for len(queue) > 0 {
	// }
	// time.Sleep(3 * time.Second)

	// 2.
	go push()
	g.ExecuteWithChan(ctx, queue, f)
}
