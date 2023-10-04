package eventrequire

import (
	"github.com/stretchr/testify/require"
	"github.com/tangelo-labs/go-domain/events"
	"github.com/tangelo-labs/go-domain/events/eventtest/eventassert"
)

// RecordedLen asserts that the given Recorder object has recorded the
// specified event type the specified number of times.
func RecordedLen(t require.TestingT, recorder events.Recorder, event interface{}, length int, msgAndArgs ...interface{}) {
	if eventassert.RecordedLen(t, recorder, event, length, msgAndArgs...) {
		return
	}

	t.FailNow()
}

// WasRecorded asserts that the given Recorder object has recorded the
// specified event type.
func WasRecorded(t require.TestingT, recorder events.Recorder, event interface{}, msgAndArgs ...interface{}) {
	if eventassert.WasRecorded(t, recorder, event, msgAndArgs...) {
		return
	}

	t.FailNow()
}

// WasNotRecorded asserts that the given Recorder object has NOT recorded the
// specified event.
func WasNotRecorded(t require.TestingT, recorder events.Recorder, event interface{}, msgAndArgs ...interface{}) {
	if eventassert.WasNotRecorded(t, recorder, event, msgAndArgs...) {
		return
	}

	t.FailNow()
}

// SequenceWasRecorded asserts that the given sequence of events was recorded.
func SequenceWasRecorded(t require.TestingT, recorder events.Recorder, sequence []interface{}, msgAndArgs ...interface{}) {
	if eventassert.SequenceWasRecorded(t, recorder, sequence, msgAndArgs...) {
		return
	}

	t.FailNow()
}

// Condition asserts that the given Recorder object has recorded the
// specified event type.
func Condition(t require.TestingT, recorder events.Recorder, condition func(event interface{}) bool, msgAndArgs ...interface{}) {
	if eventassert.Condition(t, recorder, condition, msgAndArgs...) {
		return
	}

	t.FailNow()
}
