// Package drain implements a composable message distribution package for Go.
//
// This package is a forked and altered version of the original package from the
// Docker project, which can be found at: https://github.com/docker/go-events
package drain

import (
	"fmt"
	"io"
)

var (
	// ErrSinkClosed is returned if Writer.Write call is issued to a sink that
	// has been closed. If encountered, the error should be considered terminal
	// and retries will not be successful.
	ErrSinkClosed = fmt.Errorf("sink closed")
)

// Writer defines a component where messages can be written to.
type Writer[M any] interface {
	// Write writes a message. If no error is returned, the caller can assume
	// that the message have been committed. If an error is received, the caller
	// may retry sending the message.
	Write(M) error
}

// Sink accepts and sends messages.
// A sink once closed, will not accept any more messages.
type Sink[M any] interface {
	Writer[M]
	io.Closer
}

// WriteErrorFn defines a function that is invoked each time a message fails
// to be written to the underlying sink.
type WriteErrorFn[M any] func(M, error)

func noopWriteError[M any](M, error) {}
