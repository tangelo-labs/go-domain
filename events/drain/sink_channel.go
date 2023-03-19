package drain

import (
	"fmt"
	"sync"
)

// ChannelSink defines a sink that can be listened on. The writer and channel
// listener must operate in separate goroutines.
//
// Consumers should listen on Channel.C until Closed is closed.
type ChannelSink[M any] interface {
	Sink[M]

	// Done returns a channel that will always proceed once the sink is closed.
	Done() <-chan struct{}

	// Wait returns a channel that unblocks when a new Event arrives.
	// Must be called in a separate goroutine from the writer.
	Wait() <-chan M
}

type channelSink[M any] struct {
	*baseSink
	c         chan M
	closeOnce sync.Once
}

// NewChannel returns a channel. If buffer is zero, the channel is
// unbuffered.
func NewChannel[M any](buffer int) ChannelSink[M] {
	return &channelSink[M]{
		baseSink: newBaseSink(),
		c:        make(chan M, buffer),
	}
}

func (ch *channelSink[M]) Done() <-chan struct{} {
	return ch.Closed()
}

// Write the event to the channel. Must be called in a separate goroutine from
// the listener.
func (ch *channelSink[M]) Write(event M) error {
	if ch.baseSink.IsClosed() {
		return fmt.Errorf("%w: channel sink could not write event %T", ErrSinkClosed, event)
	}

	select {
	case <-ch.Closed():
		return fmt.Errorf("%w: channel sink could not write event %T", ErrSinkClosed, event)
	case ch.c <- event:
		return nil
	}
}

func (ch *channelSink[M]) Wait() <-chan M {
	return ch.c
}

// Close the channel sink.
func (ch *channelSink[M]) Close() error {
	if errB := ch.baseSink.Close(); errB != nil {
		return fmt.Errorf("%w: channel sink could not close", errB)
	}

	ch.closeOnce.Do(func() {
		close(ch.c)
	})

	return nil
}
