package eventtest

import (
	"github.com/stretchr/testify/require"
	"github.com/tangelo-labs/go-domain/events"
	"github.com/tangelo-labs/go-domain/events/eventtest/eventassert"
)

// Assert is a simple syntactic sugar for "eventassert" package.
//
// Usage:
//
//	eventtest.Assert.WasRecorded(t, recorder, event)
var Assert = &evAssert{}

type evAssert struct{}

func (a *evAssert) RecordedLen(t require.TestingT, recorder events.Recorder, event interface{}, length int, msgAndArgs ...interface{}) bool {
	return eventassert.RecordedLen(t, recorder, event, length, msgAndArgs...)
}

func (a *evAssert) WasRecorded(t require.TestingT, recorder events.Recorder, event interface{}, msgAndArgs ...interface{}) bool {
	return eventassert.WasRecorded(t, recorder, event, msgAndArgs...)
}

func (a *evAssert) WasNotRecorded(t require.TestingT, recorder events.Recorder, event interface{}, msgAndArgs ...interface{}) bool {
	return eventassert.WasNotRecorded(t, recorder, event, msgAndArgs...)
}

func (a *evAssert) SequenceWasRecorded(t require.TestingT, recorder events.Recorder, sequence []interface{}, msgAndArgs ...interface{}) bool {
	return eventassert.SequenceWasRecorded(t, recorder, sequence, msgAndArgs...)
}
