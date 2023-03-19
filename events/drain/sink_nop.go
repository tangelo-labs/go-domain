package drain

type nopSink[M any] struct {
	*baseSink
}

// NewNop builds a sink that does nothing.
func NewNop[M any]() Sink[M] {
	return &nopSink[M]{
		baseSink: newBaseSink(),
	}
}

func (n nopSink[M]) Write(M) error {
	return nil
}
