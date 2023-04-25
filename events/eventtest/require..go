package eventtest

import (
	"github.com/stretchr/testify/require"
	"github.com/tangelo-labs/go-domain/events"
	"github.com/tangelo-labs/go-domain/events/eventtest/eventrequire"
)

// Require is a simple syntactic sugar for "eventrequire" package.
//
// Usage:
//
//	eventtest.Require.WasRecorded(t, recorder, event)
var Require = &evRequire{}

type evRequire struct{}

func (r *evRequire) RecordedLen(t require.TestingT, recorder events.Recorder, event interface{}, length int, msgAndArgs ...interface{}) {
	eventrequire.RecordedLen(t, recorder, event, length, msgAndArgs...)
}

func (r *evRequire) WasRecorded(t require.TestingT, recorder events.Recorder, event interface{}, msgAndArgs ...interface{}) {
	eventrequire.WasRecorded(t, recorder, event, msgAndArgs...)
}

func (r *evRequire) WasNotRecorded(t require.TestingT, recorder events.Recorder, event interface{}, msgAndArgs ...interface{}) {
	eventrequire.WasNotRecorded(t, recorder, event, msgAndArgs...)
}

func (r *evRequire) SequenceWasRecorded(t require.TestingT, recorder events.Recorder, sequence []interface{}, msgAndArgs ...interface{}) {
	eventrequire.SequenceWasRecorded(t, recorder, sequence, msgAndArgs...)
}
