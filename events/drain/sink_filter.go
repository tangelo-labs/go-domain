package drain

import (
	"fmt"
)

// FilterFn defines a function filters out messages.
// If the function returns true, the message will be passed to the underlying
// sink. Otherwise, the message will be silently dropped.
type FilterFn[M any] func(M) bool

type filterSink[M any] struct {
	*baseSink
	dst    Sink[M]
	filter FilterFn[M]
}

// NewFilter returns a new filter that will send to messages to dst that return
// true for FilterFn.
func NewFilter[M any](dst Sink[M], matcher FilterFn[M]) Sink[M] {
	return &filterSink[M]{
		baseSink: newCloseTrait(),
		dst:      dst,
		filter:   matcher,
	}
}

// Write a message to the filter.
func (f *filterSink[M]) Write(message M) error {
	if f.baseSink.IsClosed() {
		return fmt.Errorf("%w: filter sink could not write message %T", ErrSinkClosed, message)
	}

	if f.filter(message) {
		if errD := f.dst.Write(message); errD != nil {
			return fmt.Errorf("%w: filter sink could not write message %T in underlying sink", errD, message)
		}
	}

	return nil
}

// Close the filter and allow no more messages to pass through.
func (f *filterSink[M]) Close() error {
	if errS := f.dst.Close(); errS != nil {
		return fmt.Errorf("%w: filter sink could not close underlying sink", errS)
	}

	if errB := f.baseSink.Close(); errB != nil {
		return fmt.Errorf("%w: filter sink could not close underlying sink", errB)
	}

	return nil
}
