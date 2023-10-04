package eventassert

import (
	"fmt"
	"reflect"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tangelo-labs/go-domain/events"
)

// RecordedLen asserts that the given Recorder object has recorded the
// specified event type the specified number of times.
func RecordedLen(t assert.TestingT, recorder events.Recorder, event interface{}, length int, msgAndArgs ...interface{}) bool {
	if l := len(fetchRecorded(recorder, event)); l != length {
		return assert.Fail(t, fmt.Sprintf("Event %T was not recorded #%d times, but was expecting #%d", event, l, length), msgAndArgs...)
	}

	return true
}

// WasRecorded asserts that the given Recorder object has recorded the
// specified event type.
func WasRecorded(t assert.TestingT, recorder events.Recorder, event interface{}, msgAndArgs ...interface{}) bool {
	if len(fetchRecorded(recorder, event)) == 0 {
		return assert.Fail(t, fmt.Sprintf("Event %T was not recorded", event), msgAndArgs...)
	}

	return true
}

// WasNotRecorded asserts that the given Recorder object has NOT recorded the
// specified event.
func WasNotRecorded(t assert.TestingT, recorder events.Recorder, event interface{}, msgAndArgs ...interface{}) bool {
	if len(fetchRecorded(recorder, event)) > 0 {
		return assert.Fail(t, fmt.Sprintf("Event %T was actually recorded", event), msgAndArgs...)
	}

	return true
}

// SequenceWasRecorded asserts that the given sequence of events was recorded.
func SequenceWasRecorded(t assert.TestingT, recorder events.Recorder, sequence []interface{}, msgAndArgs ...interface{}) bool {
	emitted := recorder.Changes()
	expectedLen := len(sequence)
	index := 0
	longestSeq := 0

	for _, ev := range emitted {
		if reflect.TypeOf(ev) != reflect.TypeOf(sequence[index]) {
			index = 0

			continue
		}

		if index+1 > expectedLen-1 {
			break
		}

		index++
		if index > longestSeq {
			longestSeq = index
		}
	}

	if expectedLen != index+1 {
		return assert.Fail(t, fmt.Sprintf("Expected events sequence was not found, longest seq. was %d", longestSeq), msgAndArgs...)
	}

	return true
}

// Condition asserts that the given Recorder object has recorded an event that
// matches the given condition.
//
// The provided condition function will be called for each recorded event until
// it returns true.
func Condition(t require.TestingT, recorder events.Recorder, condition func(event interface{}) bool, msgAndArgs ...interface{}) bool {
	for _, change := range recorder.Changes() {
		if condition(change) {
			return true
		}
	}

	return assert.Fail(t, "Condition failed", msgAndArgs...)
}

func fetchRecorded(recorder events.Recorder, event interface{}) []interface{} {
	expectedKind := reflect.TypeOf(event)
	found := make([]interface{}, 0)

	for _, v := range recorder.Changes() {
		if reflect.TypeOf(v) == expectedKind {
			found = append(found, v)
		}
	}

	return found
}
