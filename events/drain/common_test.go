package drain_test

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tangelo-labs/go-domain/events/drain"
)

type tOrB interface {
	Fatalf(format string, args ...interface{})
	Logf(format string, args ...interface{})
}

type testSink[M any] struct {
	t tOrB

	messages []M
	expected int

	closed bool
	mu     sync.Mutex
}

func newTestSink[M any](t tOrB, expected int) *testSink[M] {
	return &testSink[M]{
		t:        t,
		messages: make([]M, 0, expected), // pre-allocate so we aren't benching alloc
		expected: expected,
	}
}

func (ts *testSink[M]) Write(message M) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.closed {
		return drain.ErrSinkClosed
	}

	ts.messages = append(ts.messages, message)

	if len(ts.messages) > ts.expected {
		ts.t.Fatalf("len(ts.messages) == %v, expected %v", len(ts.messages), ts.expected)
	}

	return nil
}

func (ts *testSink[M]) Close() error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.closed {
		return nil
	}

	ts.closed = true

	if len(ts.messages) != ts.expected {
		ts.t.Fatalf("len(ts.messages) == %v, expected %v", len(ts.messages), ts.expected)
	}

	return nil
}

type delayedSink[M any] struct {
	drain.Sink[M]
	delay time.Duration
}

func (ds *delayedSink[M]) Write(message M) error {
	time.Sleep(ds.delay)

	return ds.Sink.Write(message)
}

type flakySink[M any] struct {
	drain.Sink[M]
	rate float64
	mu   sync.Mutex
}

func (fs *flakySink[M]) Write(message M) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if rand.Float64() < fs.rate {
		return fmt.Errorf("error writing message: %v", message)
	}

	return fs.Sink.Write(message)
}

type dropperSink[M any] struct {
	err error

	closed bool
	mu     sync.Mutex
}

func (d *dropperSink[M]) Write(M) error {
	return d.err
}

func (d *dropperSink[M]) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return nil
	}

	d.closed = true

	return nil
}

func checkClose[M any](t *testing.T, sink drain.Sink[M]) {
	if err := sink.Close(); err != nil {
		t.Fatalf("unexpected error closing: %v", err)
	}

	// second close should not crash and should not return error.
	if err := sink.Close(); err != nil {
		t.Fatalf("unexpected error on double close: %v", err)
	}

	var fail M

	// Write after closed should be an error
	if err := sink.Write(fail); err == nil {
		t.Fatalf("write after closed did not have an error")
	} else if !errors.Is(err, drain.ErrSinkClosed) {
		t.Fatalf("error should be ErrSinkClosed")
	}
}

func benchmarkSink[M any](b *testing.B, sink drain.Sink[M]) {
	defer func() {
		require.NoError(b, sink.Close())
	}()

	var m M

	for i := 0; i < b.N; i++ {
		require.NoError(b, sink.Write(m))
	}
}
