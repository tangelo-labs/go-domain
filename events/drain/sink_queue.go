package drain

import (
	"container/list"
	"fmt"
	"sync"
)

// queue accepts all messages into a queue for asynchronous consumption
// by a sink. It is unbounded and thread safe but the sink must be reliable or
// events will be dropped.
type queueSink[M any] struct {
	*baseSink
	dst          Sink[M]
	list         *list.List
	cond         *sync.Cond
	mu           sync.Mutex
	closing      bool
	dropHandling WriteErrorFn[M]
}

type queueEnvelope[M any] struct {
	event  M
	closed bool
}

// NewQueue returns a queue Sink with a given throughput to the provided Sink dst.
// nil dropHandling will set a noop handler.
func NewQueue[M any](dst Sink[M], throughput int, dropHandling WriteErrorFn[M]) Sink[M] {
	dh := dropHandling
	if dh == nil {
		dh = noopWriteError[M]
	}

	eq := &queueSink[M]{
		baseSink:     newBaseSink(),
		dst:          dst,
		list:         list.New(),
		dropHandling: dh,
	}

	if throughput <= 0 {
		throughput = 1
	}

	eq.cond = sync.NewCond(&eq.mu)
	for i := 0; i < throughput; i++ {
		go eq.run()
	}

	return eq
}

// Write accepts the events into the queue, only failing if the queue has
// been closed.
func (eq *queueSink[M]) Write(event M) error {
	eq.mu.Lock()
	defer eq.mu.Unlock()

	if eq.baseSink.IsClosed() {
		return fmt.Errorf("%w: writer sink could not write event %T", ErrSinkClosed, event)
	}

	eq.list.PushBack(queueEnvelope[M]{event: event})
	eq.cond.Signal() // signal waiters

	return nil
}

// Close shutdown the event queue, flushing.
func (eq *queueSink[M]) Close() error {
	eq.mu.Lock()
	defer eq.mu.Unlock()

	if eq.IsClosed() {
		return nil
	}

	// set closing flag
	eq.closing = true
	eq.cond.Signal() // signal flushes queue
	eq.cond.Wait()   // wait for signal from last flush

	if errD := eq.dst.Close(); errD != nil {
		eq.closing = false

		return fmt.Errorf("%w: queue sink could not close underlying sink", errD)
	}

	if errB := eq.baseSink.Close(); errB != nil {
		eq.closing = false

		return fmt.Errorf("%w: queue sink could not close", errB)
	}

	return nil
}

// run is the main goroutine to flush events to the target sink.
func (eq *queueSink[M]) run() {
	for {
		envelope := eq.next()
		if envelope.closed {
			return // queueClosed block means event queue is closed.
		}

		if err := eq.dst.Write(envelope.event); err != nil {
			eq.dropHandling(envelope.event, err)
		}
	}
}

// next encompasses the critical section of the run loop. When the queue is
// empty, it will block on the condition. If new data arrives, it will wake
// and return a block. When closed, queueClosed constant will be returned.
func (eq *queueSink[M]) next() queueEnvelope[M] {
	eq.mu.Lock()
	defer eq.mu.Unlock()

	for eq.list.Len() < 1 {
		if eq.closing || eq.IsClosed() {
			eq.cond.Broadcast()

			return queueEnvelope[M]{closed: true}
		}

		eq.cond.Wait()
	}

	front := eq.list.Front()
	block, ok := front.Value.(queueEnvelope[M])

	if !ok {
		fmt.Printf("queue sink fatal error, not an event interface in the queue\n")
	}

	eq.list.Remove(front)

	return block
}
