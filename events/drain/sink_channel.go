package drain

import (
	"fmt"
	"sync"

	"github.com/tangelo-labs/go-domain/events"
)

// ChannelSink defines a sink that can be listened on. The writer and channel
// listener must operate in separate goroutines.
//
// Consumers should listen on Channel.C until Closed is closed.
type ChannelSink interface {
	Sink

	// Done returns a channel that will always proceed once the sink is closed.
	Done() <-chan struct{}

	// Wait returns a channel that unblocks when a new Event arrives.
	// Must be called in a separate goroutine from the writer.
	Wait() <-chan events.Event
}

type channelSink struct {
	*baseSink
	c         chan events.Event
	closeOnce sync.Once
}

// NewChannel returns a channel. If buffer is zero, the channel is
// unbuffered.
func NewChannel(buffer int) ChannelSink {
	return &channelSink{
		baseSink: newBaseSink(),
		c:        make(chan events.Event, buffer),
	}
}

func (ch *channelSink) Done() <-chan struct{} {
	return ch.Closed()
}

// Write the event to the channel. Must be called in a separate goroutine from
// the listener.
func (ch *channelSink) Write(event events.Event) error {
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

func (ch *channelSink) Wait() <-chan events.Event {
	return ch.c
}

// Close the channel sink.
func (ch *channelSink) Close() error {
	if errB := ch.baseSink.Close(); errB != nil {
		return fmt.Errorf("%w: channel sink could not close", errB)
	}

	ch.closeOnce.Do(func() {
		close(ch.c)
	})

	return nil
}
