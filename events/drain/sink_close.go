package drain

import "sync"

// baseSink is a trait that can be embedded into other sinks implementations
// to provide closing functionality.
type baseSink struct {
	closed chan struct{}
	once   sync.Once
}

func newCloseTrait() *baseSink {
	return &baseSink{
		closed: make(chan struct{}),
	}
}

// Close the sink, possibly waiting for pending messages to flush.
func (bs *baseSink) Close() error {
	bs.once.Do(func() {
		close(bs.closed)
	})

	return nil
}

// Closed returns channel that to check if sink is closed or not.
func (bs *baseSink) Closed() <-chan struct{} {
	return bs.closed
}

// IsClosed returns true if the sink is closed.
func (bs *baseSink) IsClosed() bool {
	select {
	case <-bs.closed:
		return true
	default:
		return false
	}
}
