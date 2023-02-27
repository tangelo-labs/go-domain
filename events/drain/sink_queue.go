package drain

import (
	"container/list"
	"fmt"
	"sync"

	"github.com/tangelo-labs/go-domainkit/events"
)

const queueClosed = ":closed:"

// queue accepts all messages into a queue for asynchronous consumption
// by a sink. It is unbounded and thread safe but the sink must be reliable or
// events will be dropped.
type queueSink struct {
	*baseSink
	dst          Sink
	events       *list.List
	cond         *sync.Cond
	mu           sync.Mutex
	closing      bool
	dropHandling func(dropped events.Event, err error)
}

// NewQueue returns a queue Sink with a given throughput to the provided Sink dst.
// nil dropHandling will set a noop handler.
func NewQueue(dst Sink, throughput int, dropHandling WriteErrorFn) Sink {
	dh := dropHandling
	if dh == nil {
		dh = noopWriteError
	}

	eq := &queueSink{
		baseSink:     newBaseSink(),
		dst:          dst,
		events:       list.New(),
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
func (eq *queueSink) Write(event events.Event) error {
	eq.mu.Lock()
	defer eq.mu.Unlock()

	if eq.baseSink.IsClosed() {
		return fmt.Errorf("%w: writer sink could not write event %T", ErrSinkClosed, event)
	}

	eq.events.PushBack(event)
	eq.cond.Signal() // signal waiters

	return nil
}

// Close shutdown the event queue, flushing.
func (eq *queueSink) Close() error {
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
func (eq *queueSink) run() {
	for {
		event := eq.next()

		if event == queueClosed {
			return // queueClosed block means event queue is closed.
		}

		if err := eq.dst.Write(event); err != nil {
			eq.dropHandling(event, err)
		}
	}
}

// next encompasses the critical section of the run loop. When the queue is
// empty, it will block on the condition. If new data arrives, it will wake
// and return a block. When closed, queueClosed constant will be returned.
func (eq *queueSink) next() events.Event {
	eq.mu.Lock()
	defer eq.mu.Unlock()

	for eq.events.Len() < 1 {
		if eq.closing || eq.IsClosed() {
			eq.cond.Broadcast()

			return queueClosed
		}

		eq.cond.Wait()
	}

	front := eq.events.Front()
	block, ok := front.Value.(events.Event)

	if !ok {
		fmt.Printf("queue sink fatal error, not an event interface in the queue\n")
	}

	eq.events.Remove(front)

	return block
}
