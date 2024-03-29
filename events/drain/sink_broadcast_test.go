package drain_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tangelo-labs/go-domain/events"
	"github.com/tangelo-labs/go-domain/events/drain"
)

func TestBroadcaster(t *testing.T) {
	const nm = 1000

	sinks := make([]drain.Sink[events.Event], 0)
	b := drain.NewBroadcaster[events.Event](noopWriteErrFn)

	for i := 0; i < 10; i++ {
		sinks = append(sinks, newTestSink[events.Event](t, nm))
		require.NoError(t, b.Add(sinks[i]))
		require.NoError(t, b.Add(sinks[i])) // noop
	}

	var wg sync.WaitGroup

	for i := 1; i <= nm; i++ {
		wg.Add(1)

		go func(m events.Event) {
			if err := b.Write(m); err != nil {
				t.Errorf("error writing message %v: %v", m, err)
			}

			wg.Done()
		}("message")
	}

	wg.Wait() // Wait until writes complete

	for i := range sinks {
		require.NoError(t, b.Remove(sinks[i]))
	}

	// sending one more should trigger test failure if they weren't removed.
	if err := b.Write("onemore"); err != nil {
		t.Fatalf("unexpected error sending one more: %v", err)
	}

	// add them back to test closing.
	for i := range sinks {
		require.NoError(t, b.Add(sinks[i]))
	}

	checkClose[events.Event](t, b)

	// Iterate through the sinks and check that they all have the expected length.
	for _, sink := range sinks {
		ts, ok := sink.(*testSink[events.Event])
		if !ok {
			continue
		}

		ts.mu.Lock()

		if len(ts.messages) != nm {
			ts.mu.Unlock()
			t.Fatalf("not all messages ended up in testsink: len(testSink) == %d, not %d", len(ts.messages), nm)
		}

		if !ts.closed {
			ts.mu.Unlock()
			t.Fatalf("sink should have been closed")
		}

		ts.mu.Unlock()
	}
}

func BenchmarkBroadcast10(b *testing.B) {
	benchmarkBroadcast(b, 10)
}

func BenchmarkBroadcast100(b *testing.B) {
	benchmarkBroadcast(b, 100)
}

func BenchmarkBroadcast1000(b *testing.B) {
	benchmarkBroadcast(b, 1000)
}

func BenchmarkBroadcast10000(b *testing.B) {
	benchmarkBroadcast(b, 10000)
}

func benchmarkBroadcast(b *testing.B, nsinks int) {
	b.StopTimer()

	sinks := make([]drain.Sink[events.Event], 0)

	for i := 0; i < nsinks; i++ {
		sinks = append(sinks, newTestSink[events.Event](b, b.N))
	}

	b.StartTimer()
	benchmarkSink[events.Event](b, drain.NewBroadcaster[events.Event](noopWriteErrFn, sinks...))
}

func noopWriteErrFn(_ events.Event, _ error) {}
