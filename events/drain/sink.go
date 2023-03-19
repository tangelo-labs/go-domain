// Package drain implements a composable event distribution package for Go.
//
// This package is a forked and altered version of the original package from the
// Docker project, which can be found at: https://github.com/docker/go-events
package drain

import (
	"fmt"
	"io"
	"sync"
)

var (
	// ErrSinkClosed is returned if Writer.Write call is issued to a sink that
	// has been closed. If encountered, the error should be considered terminal
	// and retries will not be successful.
	ErrSinkClosed = fmt.Errorf("sink closed")
)

// Writer defines a component where events can be written to.
type Writer[M any] interface {
	// Write writes an event. If no error is returned, the caller will
	// assume that all events have been committed. If an error is received, the
	// caller may retry sending the event.
	Write(event M) error
}

// Sink accepts and sends events.
type Sink[M any] interface {
	Writer[M]
	io.Closer
}

// WriteErrorFn defines a function that is invoked each time an event fails
// to be written to the underlying sink.
type WriteErrorFn[M any] func(event M, err error)

func noopWriteError[M any](M, error) {}

// baseSink is a trait that can be embedded into other sinks to provide
// common functionality.
type baseSink struct {
	closed chan struct{}
	once   sync.Once
}

// newBaseSink builds a new baseSink instance.
func newBaseSink() *baseSink {
	return &baseSink{
		closed: make(chan struct{}),
	}
}

// Close the sink, possibly waiting for pending events to flush.
func (bs *baseSink) Close() error {
	bs.once.Do(func() {
		close(bs.closed)
	})

	return nil
}

// Closed returns channel that to check if sink is closed or not.
func (bs *baseSink) Closed() <-chan struct{} {
	return bs.closed
}

// IsClosed returns true if the sink is closed.
func (bs *baseSink) IsClosed() bool {
	select {
	case <-bs.closed:
		return true
	default:
		return false
	}
}
