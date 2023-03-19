package drain

import (
	"fmt"
)

// FilterFn defines a function filters out events.
// If the function returns true, the event will be passed to the underlying
// sink. Otherwise, the event will be silently dropped.
type FilterFn[M any] func(event M) bool

type filterSink[M any] struct {
	*baseSink
	dst    Sink[M]
	filter FilterFn[M]
}

// NewFilter returns a new filter that will send to events to dst that return
// true for FilterFn.
func NewFilter[M any](dst Sink[M], matcher FilterFn[M]) Sink[M] {
	return &filterSink[M]{
		baseSink: newBaseSink(),
		dst:      dst,
		filter:   matcher,
	}
}

// Write an event to the filter.
func (f *filterSink[M]) Write(event M) error {
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
func (f *filterSink[M]) Close() error {
	if errS := f.dst.Close(); errS != nil {
		return fmt.Errorf("%w: filter sink could not close underlying sink", errS)
	}

	if errB := f.baseSink.Close(); errB != nil {
		return fmt.Errorf("%w: filter sink could not close underlying sink", errB)
	}

	return nil
}
