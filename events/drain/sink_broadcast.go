package drain

import (
	"errors"
	"fmt"
	"sync"

	"github.com/tangelo-labs/go-domainkit/events"
)

// BroadcasterSink sends events to multiple, reliable Sinks. The goal of this
// component is to dispatch events to configured endpoints. Reliability can be
// provided by wrapping incoming sinks.
type BroadcasterSink interface {
	Sink

	// Add adds the sink to the broadcaster.
	// The provided sink must be comparable with equality. Typically, this just
	// works with a regular pointer type.
	Add(sink Sink) error

	// Remove the provided sink.
	Remove(sink Sink) error
}

type broadcaster struct {
	*baseSink
	sinks   []Sink
	events  chan events.Event
	adds    chan bcConfigureRequest
	removes chan bcConfigureRequest

	wErrHandler WriteErrorFn
	closeOnce   sync.Once
}

type bcConfigureRequest struct {
	sink     Sink
	response chan error
}

// NewBroadcaster appends one or more sinks to the list of sinks. The
// broadcaster behavior will be affected by the properties of the sink.
// Generally, the sink should accept all messages and deal with reliability on
// its own. Use of QueueSink and RetryingSink should be used here.
func NewBroadcaster(wErrHandler WriteErrorFn, to ...Sink) BroadcasterSink {
	b := &broadcaster{
		baseSink:    newBaseSink(),
		sinks:       to,
		events:      make(chan events.Event),
		adds:        make(chan bcConfigureRequest),
		removes:     make(chan bcConfigureRequest),
		wErrHandler: wErrHandler,
	}

	// Start the broadcaster
	go b.run()

	return b
}

// Write accepts an event to be dispatched to all sinks. This method will never
// fail and should never block (hopefully!). The caller cedes the memory to the
// broadcaster and should not modify it after calling write.
func (b *broadcaster) Write(event events.Event) error {
	if b.baseSink.IsClosed() {
		return fmt.Errorf("%w: broadcaster sink could not write event %T", ErrSinkClosed, event)
	}

	select {
	case <-b.Closed():
		return fmt.Errorf("%w: broadcaster sink failed write event %T", ErrSinkClosed, event)
	case b.events <- event:
	}

	return nil
}

func (b *broadcaster) Add(sink Sink) error {
	return b.configure(b.adds, sink)
}

func (b *broadcaster) Remove(sink Sink) error {
	return b.configure(b.removes, sink)
}

// Close the broadcaster, ensuring that all messages are flushed to the
// underlying sink before returning.
func (b *broadcaster) Close() error {
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
// Under normal conditions, it waits for events on the event channel. After
// Close is called, this goroutine will exit.
func (b *broadcaster) run() {
	remove := func(target Sink) {
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
		case event := <-b.events:
			for _, sink := range b.sinks {
				if err := sink.Write(event); err != nil {
					if errors.Is(err, ErrSinkClosed) {
						// remove closed sinks
						remove(sink)

						continue
					}

					b.wErrHandler(event, err)
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

func (b *broadcaster) configure(ch chan bcConfigureRequest, sink Sink) error {
	response := make(chan error, 1)

	for {
		select {
		case <-b.Closed():
			return fmt.Errorf("%w: broadcaster sink failed to configure", ErrSinkClosed)
		case ch <- bcConfigureRequest{
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
