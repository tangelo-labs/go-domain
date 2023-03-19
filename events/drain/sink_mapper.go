package drain

import (
	"fmt"
)

// MapperFn defines a function maps an input event into another.
type MapperFn[M any] func(event M) (M, error)

// mapperSink provides a sink that maps events into other events.
type mapperSink[M any] struct {
	*baseSink
	dst    Sink[M]
	mapper MapperFn[M]
}

// NewMapper builds sink that passes to dst mapped events.
func NewMapper[M any](dst Sink[M], mapper MapperFn[M]) Sink[M] {
	sink := &mapperSink[M]{
		baseSink: newBaseSink(),
		mapper:   mapper,
		dst:      dst,
	}

	return sink
}

func (ms *mapperSink[M]) Write(event M) error {
	if ms.baseSink.IsClosed() {
		return fmt.Errorf("%w: mapper sink could not write event %T", ErrSinkClosed, event)
	}

	ev, errM := ms.mapper(event)
	if errM != nil {
		return fmt.Errorf("%w: mapper sink could not map event %T", errM, event)
	}

	if errD := ms.dst.Write(ev); errD != nil {
		return fmt.Errorf("%w: mapper sink could not write event %T in underlying sink", errD, event)
	}

	return nil
}

func (ms *mapperSink[M]) Close() error {
	if errD := ms.dst.Close(); errD != nil {
		return fmt.Errorf("%w: could not close underlying sink", errD)
	}

	if errB := ms.baseSink.Close(); errB != nil {
		return fmt.Errorf("%w: mapper sink failed to close", errB)
	}

	return nil
}
