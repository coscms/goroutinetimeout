package goroutinetimeout_test

import (
	"sync"
	"testing"
	"time"

	"github.com/coscms/goroutinetimeout"
	"golang.org/x/net/context"
)

func testBase(t *testing.T, asyncPush bool) {
	wg := sync.WaitGroup{}
	ctx := context.Background()
	queue := make(chan interface{}, 5)
	push := func(i int) {
		wg.Add(i)
		for _i := 0; _i < i; _i++ {
			queue <- _i
		}
		close(queue)
	}
	g := goroutinetimeout.New(`TestChan`, nil, 4)
	f := func(v interface{}) {
		i := v.(int)
		time.Sleep(2 * time.Second)
		t.Logf(`Execute.%d`, i)
		wg.Done()
	}

	if !asyncPush {
		// 1.
		go g.ExecuteWithChan(ctx, queue, f)
		push(10)
	} else {
		// 2.
		go push(10)
		g.ExecuteWithChan(ctx, queue, f)
	}

	wg.Wait()
}

func TestBaseAsyncPush(t *testing.T) {
	testBase(t, true)
}

func TestBaseSyncPush(t *testing.T) {
	testBase(t, false)
}
