package drain

import (
	"errors"
	"fmt"
	"sync"
)

// BroadcasterSink sends messages to multiple, reliable Sinks. The goal of this
// component is to dispatch messages to configured endpoints. Reliability can be
// provided by wrapping incoming sinks.
type BroadcasterSink[M any] interface {
	Sink[M]

	// Add adds the sink to the broadcaster.
	// The provided sink must be comparable with equality. Typically, this just
	// works with a regular pointer type.
	Add(sink Sink[M]) error

	// Remove the provided sink.
	Remove(sink Sink[M]) error
}

type broadcaster[M any] struct {
	*baseSink
	sinks    []Sink[M]
	messages chan M
	adds     chan bcConfigureRequest[M]
	removes  chan bcConfigureRequest[M]

	wErrHandler WriteErrorFn[M]
	closeOnce   sync.Once
}

type bcConfigureRequest[M any] struct {
	sink     Sink[M]
	response chan error
}

// NewBroadcaster appends one or more sinks to the list of sinks. The
// broadcaster behavior will be affected by the properties of the sink.
// Generally, the sink should accept all messages and deal with reliability on
// its own. Use of QueueSink and RetryingSink should be used here.
func NewBroadcaster[M any](wErrHandler WriteErrorFn[M], to ...Sink[M]) BroadcasterSink[M] {
	b := &broadcaster[M]{
		baseSink:    newCloseTrait(),
		sinks:       to,
		messages:    make(chan M),
		adds:        make(chan bcConfigureRequest[M]),
		removes:     make(chan bcConfigureRequest[M]),
		wErrHandler: wErrHandler,
	}

	// Start the broadcaster
	go b.run()

	return b
}

// Write accepts a message to be dispatched to all sinks. This method will never
// fail and should never block (hopefully!). The caller cedes the memory to the
// broadcaster and should not modify it after calling write.
func (b *broadcaster[M]) Write(message M) error {
	if b.baseSink.IsClosed() {
		return fmt.Errorf("%w: broadcaster sink could not write message %T", ErrSinkClosed, message)
	}

	select {
	case <-b.Closed():
		return fmt.Errorf("%w: broadcaster sink failed write message %T", ErrSinkClosed, message)
	case b.messages <- message:
	}

	return nil
}

func (b *broadcaster[M]) Add(sink Sink[M]) error {
	return b.configure(b.adds, sink)
}

func (b *broadcaster[M]) Remove(sink Sink[M]) error {
	return b.configure(b.removes, sink)
}

// Close the broadcaster, ensuring that all messages are flushed to the
// underlying sink before returning.
func (b *broadcaster[M]) Close() error {
	b.closeOnce.Do(func() {
		// try to close all the underlying sinks
		for _, sink := range b.sinks {
			if err := sink.Close(); err != nil && !errors.Is(err, ErrSinkClosed) {
				fmt.Println("[warning] error closing broadcast child sink:", err.Error())
			}
		}
	})

	if err := b.baseSink.Close(); err != nil {
		return fmt.Errorf("%w: broadcaster sink could not close", err)
	}

	return nil
}

// run is the main broadcast loop, started when the broadcaster is created.
// Under normal conditions, it waits for messages on the message channel. After
// Close is called, this goroutine will exit.
func (b *broadcaster[M]) run() {
	remove := func(target Sink[M]) {
		for i, sink := range b.sinks {
			if sink == target {
				b.sinks = append(b.sinks[:i], b.sinks[i+1:]...)

				break
			}
		}
	}

	for {
		select {
		case <-b.Closed():
			return
		case m := <-b.messages:
			for _, sink := range b.sinks {
				if err := sink.Write(m); err != nil {
					if errors.Is(err, ErrSinkClosed) {
						// remove closed sinks
						remove(sink)

						continue
					}

					b.wErrHandler(m, err)
				}
			}
		case request := <-b.adds:
			var found bool

			// while we have to iterate for add/remove, common iteration for
			// send is faster against slice.

			for _, sink := range b.sinks {
				if request.sink == sink {
					found = true

					break
				}
			}

			if !found {
				b.sinks = append(b.sinks, request.sink)
			}

			request.response <- nil
		case request := <-b.removes:
			remove(request.sink)
			request.response <- nil
		}
	}
}

func (b *broadcaster[M]) configure(ch chan bcConfigureRequest[M], sink Sink[M]) error {
	response := make(chan error, 1)

	for {
		select {
		case <-b.Closed():
			return fmt.Errorf("%w: broadcaster sink failed to configure", ErrSinkClosed)
		case ch <- bcConfigureRequest[M]{
			sink:     sink,
			response: response,
		}:
			ch = nil
		case errR := <-response:
			if errR != nil {
				return fmt.Errorf("%w: broadcaster sink received errored response when configuring", errR)
			}

			return nil
		}
	}
}
