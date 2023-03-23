package drain

import (
	"fmt"
)

// MapperFn defines a function maps an input event into another.
type MapperFn[M any] func(M) (M, error)

// mapperSink provides a sink that maps messages into other messages.
type mapperSink[M any] struct {
	*baseSink
	dst    Sink[M]
	mapper MapperFn[M]
}

// NewMapper builds sink that passes to dst mapped messages.
func NewMapper[M any](dst Sink[M], mapper MapperFn[M]) Sink[M] {
	sink := &mapperSink[M]{
		baseSink: newCloseTrait(),
		mapper:   mapper,
		dst:      dst,
	}

	return sink
}

func (ms *mapperSink[M]) Write(message M) error {
	if ms.baseSink.IsClosed() {
		return fmt.Errorf("%w: mapper sink could not write message %T", ErrSinkClosed, message)
	}

	ev, errM := ms.mapper(message)
	if errM != nil {
		return fmt.Errorf("%w: mapper sink could not map message %T", errM, message)
	}

	if errD := ms.dst.Write(ev); errD != nil {
		return fmt.Errorf("%w: mapper sink could not write message %T in underlying sink", errD, message)
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
