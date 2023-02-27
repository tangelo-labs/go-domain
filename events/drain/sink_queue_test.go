package drain_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/tangelo-labs/go-domainkit/events"
	"github.com/tangelo-labs/go-domainkit/events/drain"
	"go.uber.org/atomic"
)

func TestQueue(t *testing.T) {
	const nevents = 1000

	ts := newTestSink(t, nevents)
	eq := drain.NewQueue(
		// delayed sync simulates destination slower than channel comms
		&delayedSink{
			Sink:  ts,
			delay: time.Millisecond * 1,
		}, 1, nil)

	time.Sleep(10 * time.Millisecond) // let's queue settle to wait conidition.

	var wg sync.WaitGroup

	for i := 1; i <= nevents; i++ {
		wg.Add(1)

		go func(event events.Event) {
			defer wg.Done()

			if err := eq.Write(event); err != nil {
				t.Errorf("error writing event: %v", err)
			}
		}("event-" + fmt.Sprint(i))
	}

	wg.Wait()
	checkClose(t, eq)

	ts.mu.Lock()
	defer ts.mu.Unlock()

	if len(ts.events) != nevents {
		t.Fatalf("events did not make it to the sink: %d != %d", len(ts.events), 1000)
	}

	if !ts.closed {
		t.Fatalf("sink should have been closed")
	}
}

func TestQueueDrop(t *testing.T) {
	const nevents = 10

	cc := atomic.NewInt64(0)
	eq := drain.NewQueue(
		&dropperSink{err: errors.New("dropped")},
		1,
		func(event events.Event, err error) {
			cc.Add(1)
		},
	)

	time.Sleep(10 * time.Millisecond) // let's queue settle to wait condition.

	var wg sync.WaitGroup

	for i := 1; i <= nevents; i++ {
		wg.Add(1)

		go func(event events.Event) {
			defer wg.Done()

			if err := eq.Write(event); err != nil {
				t.Errorf("error writing event: %v", err)
			}
		}("event-" + fmt.Sprint(i))
	}

	wg.Wait()
	checkClose(t, eq)
}
