package drain

import (
	"fmt"

	"github.com/tangelo-labs/go-domainkit/events"
)

// FilterFn defines a function filters out events.
// If the function returns true, the event will be passed to the underlying
// sink. Otherwise, the event will be silently dropped.
type FilterFn func(event events.Event) bool

type filterSink struct {
	*baseSink
	dst    Sink
	filter FilterFn
}

// NewFilter returns a new filter that will send to events to dst that return
// true for FilterFn.
func NewFilter(dst Sink, matcher FilterFn) Sink {
	return &filterSink{
		baseSink: newBaseSink(),
		dst:      dst,
		filter:   matcher,
	}
}

// Write an event to the filter.
func (f *filterSink) Write(event events.Event) error {
	if f.baseSink.IsClosed() {
		return fmt.Errorf("%w: filter sink could not write event %T", ErrSinkClosed, event)
	}

	if f.filter(event) {
		if errD := f.dst.Write(event); errD != nil {
			return fmt.Errorf("%w: filter sink could not write event %T in underlying sink", errD, event)
		}
	}

	return nil
}

// Close the filter and allow no more events to pass through.
func (f *filterSink) Close() error {
	if errS := f.dst.Close(); errS != nil {
		return fmt.Errorf("%w: filter sink could not close underlying sink", errS)
	}

	if errB := f.baseSink.Close(); errB != nil {
		return fmt.Errorf("%w: filter sink could not close underlying sink", errB)
	}

	return nil
}
