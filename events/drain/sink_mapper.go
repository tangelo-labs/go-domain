package drain

import (
	"fmt"

	"github.com/tangelo-labs/go-domainkit/events"
)

// MapperFn defines a function maps an input event into another.
type MapperFn func(event events.Event) (events.Event, error)

// mapperSink provides a sink that maps events into other events.
type mapperSink struct {
	*baseSink
	dst    Sink
	mapper MapperFn
}

// NewMapper builds sink that passes to dst mapped events.
func NewMapper(dst Sink, mapper MapperFn) Sink {
	sink := &mapperSink{
		baseSink: newBaseSink(),
		mapper:   mapper,
		dst:      dst,
	}

	return sink
}

func (ms *mapperSink) Write(event events.Event) error {
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

func (ms *mapperSink) Close() error {
	if errD := ms.dst.Close(); errD != nil {
		return fmt.Errorf("%w: could not close underlying sink", errD)
	}

	if errB := ms.baseSink.Close(); errB != nil {
		return fmt.Errorf("%w: mapper sink failed to close", errB)
	}

	return nil
}
