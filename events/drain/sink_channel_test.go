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
	const nm = 100

	sink := drain.NewChannel[events.Event](0)

	go func() {
		var wg sync.WaitGroup

		for i := 1; i <= nm; i++ {
			m := "msg-" + fmt.Sprint(i)

			wg.Add(1)

			go func(m events.Event) {
				defer wg.Done()

				if err := sink.Write(m); err != nil {
					t.Errorf("error writing message: %v", err)
				}
			}(m)
		}

		wg.Wait()
		require.NoError(t, sink.Close())

		// now send another bunch of messages and ensure we stay closed
		for i := 1; i <= nm; i++ {
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

	if received != nm {
		t.Fatalf("messages did not make it through sink: %v != %v", received, nm)
	}
}
