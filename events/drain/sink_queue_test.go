package drain_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/tangelo-labs/go-domain/events"
	"github.com/tangelo-labs/go-domain/events/drain"
	"go.uber.org/atomic"
)

func TestQueue(t *testing.T) {
	const nm = 1000

	ts := newTestSink[events.Event](t, nm)
	eq := drain.NewQueue[events.Event](
		// delayed sync simulates destination slower than channel.
		&delayedSink[events.Event]{
			Sink:  ts,
			delay: time.Millisecond * 1,
		}, 1, nil)

	time.Sleep(10 * time.Millisecond) // let's queue settle to wait condition.

	var wg sync.WaitGroup

	for i := 1; i <= nm; i++ {
		wg.Add(1)

		go func(message events.Event) {
			defer wg.Done()

			if err := eq.Write(message); err != nil {
				t.Errorf("error writing message: %v", err)
			}
		}("message-" + fmt.Sprint(i))
	}

	wg.Wait()
	checkClose(t, eq)

	ts.mu.Lock()
	defer ts.mu.Unlock()

	if len(ts.messages) != nm {
		t.Fatalf("messages did not make it to the sink: %d != %d", len(ts.messages), 1000)
	}

	if !ts.closed {
		t.Fatalf("sink should have been closed")
	}
}

func TestQueueDrop(t *testing.T) {
	const nm = 10

	cc := atomic.NewInt64(0)
	eq := drain.NewQueue[events.Event](
		&dropperSink[events.Event]{err: errors.New("dropped")},
		1,
		func(m events.Event, err error) {
			cc.Add(1)
		},
	)

	time.Sleep(10 * time.Millisecond) // let's queue settle to wait condition.

	var wg sync.WaitGroup

	for i := 1; i <= nm; i++ {
		wg.Add(1)

		go func(m events.Event) {
			defer wg.Done()

			if err := eq.Write(m); err != nil {
				t.Errorf("error writing message: %v", err)
			}
		}("message-" + fmt.Sprint(i))
	}

	wg.Wait()
	checkClose(t, eq)
}
