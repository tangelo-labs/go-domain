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

// NewIOWriter builds a sink that writes messages into the provided io.Writer.
func NewIOWriter[M any](iow io.Writer, marshaller Marshaller[M]) Sink[M] {
	sink := &writerSink[M]{
		baseSink:   newCloseTrait(),
		iow:        iow,
		marshaller: marshaller,
	}

	return sink
}

func (w *writerSink[M]) Write(message M) error {
	if w.baseSink.IsClosed() {
		return fmt.Errorf("%w: writer sink could not write message %T", ErrSinkClosed, message)
	}

	b, errM := w.marshaller(message)
	if errM != nil {
		return fmt.Errorf("%w: writer sink could not marshal message %T", errM, message)
	}

	if _, errF := fmt.Fprintln(w.iow, string(b)); errF != nil {
		return fmt.Errorf("%w: writer sink could not write message %T in writer", errF, message)
	}

	return nil
}
