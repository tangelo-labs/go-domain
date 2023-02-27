package drain

import (
	"fmt"
	"io"

	"github.com/tangelo-labs/go-domainkit/events"
)

type writerSink struct {
	*baseSink
	iow        io.Writer
	marshaller Marshaller
}

// NewIOWriter builds a sink that writes events into the provided io.Writer.
func NewIOWriter(iow io.Writer, marshaller Marshaller) Sink {
	sink := &writerSink{
		baseSink:   newBaseSink(),
		iow:        iow,
		marshaller: marshaller,
	}

	return sink
}

func (w *writerSink) Write(event events.Event) error {
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
