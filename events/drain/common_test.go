package drain_test

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tangelo-labs/go-domain/events"
	"github.com/tangelo-labs/go-domain/events/drain"
)

type tOrB interface {
	Fatalf(format string, args ...interface{})
	Logf(format string, args ...interface{})
}

type testSink struct {
	t tOrB

	events   []events.Event
	expected int

	closed bool
	mu     sync.Mutex
}

func newTestSink(t tOrB, expected int) *testSink {
	return &testSink{
		t:        t,
		events:   make([]events.Event, 0, expected), // pre-allocate so we aren't benching alloc
		expected: expected,
	}
}

func (ts *testSink) Write(event events.Event) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.closed {
		return drain.ErrSinkClosed
	}

	ts.events = append(ts.events, event)

	if len(ts.events) > ts.expected {
		ts.t.Fatalf("len(ts.events) == %v, expected %v", len(ts.events), ts.expected)
	}

	return nil
}

func (ts *testSink) Close() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.closed {
		return nil
	}

	ts.closed = true

	if len(ts.events) != ts.expected {
		ts.t.Fatalf("len(ts.events) == %v, expected %v", len(ts.events), ts.expected)
	}

	return nil
}

type delayedSink struct {
	drain.Sink
	delay time.Duration
}

func (ds *delayedSink) Write(event events.Event) error {
	time.Sleep(ds.delay)

	return ds.Sink.Write(event)
}

type flakySink struct {
	drain.Sink
	rate float64
	mu   sync.Mutex
}

func (fs *flakySink) Write(event events.Event) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if rand.Float64() < fs.rate {
		return fmt.Errorf("error writing event: %v", event)
	}

	return fs.Sink.Write(event)
}

type dropperSink struct {
	err error

	closed bool
	mu     sync.Mutex
}

func (d *dropperSink) Write(_ events.Event) error {
	return d.err
}

func (d *dropperSink) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil
	}

	d.closed = true

	return nil
}

func checkClose(t *testing.T, sink drain.Sink) {
	if err := sink.Close(); err != nil {
		t.Fatalf("unexpected error closing: %v", err)
	}

	// second close should not crash and should not return error.
	if err := sink.Close(); err != nil {
		t.Fatalf("unexpected error on double close: %v", err)
	}

	// Write after closed should be an error
	if err := sink.Write("fail"); err == nil {
		t.Fatalf("write after closed did not have an error")
	} else if !errors.Is(err, drain.ErrSinkClosed) {
		t.Fatalf("error should be ErrSinkClosed")
	}
}

func benchmarkSink(b *testing.B, sink drain.Sink) {
	defer func() {
		require.NoError(b, sink.Close())
	}()

	var event = "myevent"

	for i := 0; i < b.N; i++ {
		require.NoError(b, sink.Write(event))
	}
}
