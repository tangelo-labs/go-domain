package events

import "sync"

// Recorder defines an element capable of recording events.
type Recorder interface {
	// Record tracks an event in the list of events.
	Record(event Event)

	// Changes retrieves the list of event tracked so far.
	Changes() []Event

	// ClearChanges clears the list of recorded events.
	ClearChanges()
}

// BaseRecorder is a trait that implements the common functionality for the
// Recorder interface. This object can be safely shared by multiple goroutines.
type BaseRecorder struct {
	events []Event
	mu     sync.RWMutex
}

// Record tracks an event in the list of events.
func (b *BaseRecorder) Record(event Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.events == nil {
		b.events = make([]Event, 0)
	}

	b.events = append(b.events, event)
}

// Changes retrieves the list of event tracked so far.
func (b *BaseRecorder) Changes() []Event {
	b.mu.RLock()
	defer b.mu.RUnlock()

	out := make([]Event, len(b.events))
	copy(out, b.events)

	return b.events
}

// ClearChanges clears the list of recorded events.
func (b *BaseRecorder) ClearChanges() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.events = make([]Event, 0)
}
