package drain

import "github.com/tangelo-labs/go-domainkit/events"

type nopSink struct {
	*baseSink
}

// NewNop builds a sink that does nothing.
func NewNop() Sink {
	return &nopSink{
		baseSink: newBaseSink(),
	}
}

func (n nopSink) Write(_ events.Event) error {
	return nil
}
