package drain_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tangelo-labs/go-domain/events"
	"github.com/tangelo-labs/go-domain/events/drain"
)

func TestChannel(t *testing.T) {
	const nEvents = 100

	sink := drain.NewChannel[events.Event](0)

	go func() {
		var wg sync.WaitGroup

		for i := 1; i <= nEvents; i++ {
			event := "event-" + fmt.Sprint(i)

			wg.Add(1)

			go func(event events.Event) {
				defer wg.Done()

				if err := sink.Write(event); err != nil {
					t.Errorf("error writing event: %v", err)
				}
			}(event)
		}

		wg.Wait()
		require.NoError(t, sink.Close())

		// now send another bunch of events and ensure we stay closed
		for i := 1; i <= nEvents; i++ {
			if err := sink.Write(i); !errors.Is(err, drain.ErrSinkClosed) {
				t.Errorf("unexpected error: %v != %v", err, drain.ErrSinkClosed)
			}
		}
	}()

	var received int
loop:
	for {
		select {
		case <-sink.Wait():
			received++
		case <-sink.Done():
			break loop
		}
	}

	require.NoError(t, sink.Close())

	// test will timeout if this hangs
	if _, ok := <-sink.Done(); ok {
		t.Fatalf("done should be a closed channel")
	}

	if received != nEvents {
		t.Fatalf("events did not make it through sink: %v != %v", received, nEvents)
	}
}
