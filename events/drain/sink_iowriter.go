package drain

import (
	"fmt"
	"io"
)

type writerSink[M any] struct {
	*baseSink
	iow        io.Writer
	marshaller Marshaller[M]
}

// NewIOWriter builds a sink that writes events into the provided io.Writer.
func NewIOWriter[M any](iow io.Writer, marshaller Marshaller[M]) Sink[M] {
	sink := &writerSink[M]{
		baseSink:   newBaseSink(),
		iow:        iow,
		marshaller: marshaller,
	}

	return sink
}

func (w *writerSink[M]) Write(event M) error {
	if w.baseSink.IsClosed() {
		return fmt.Errorf("%w: writer sink could not write event %T", ErrSinkClosed, event)
	}

	b, errM := w.marshaller(event)
	if errM != nil {
		return fmt.Errorf("%w: writer sink could not marshal event %T", errM, event)
	}

	if _, errF := fmt.Fprintln(w.iow, string(b)); errF != nil {
		return fmt.Errorf("%w: writer sink could not write event %T in writer", errF, event)
	}

	return nil
}
