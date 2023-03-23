package drain_test

import (
	"testing"

	"github.com/tangelo-labs/go-domain/events"
	"github.com/tangelo-labs/go-domain/events/drain"
)

func TestFilter(t *testing.T) {
	const nm = 100

	ts := newTestSink[events.Event](t, nm/2)
	filter := drain.NewFilter[events.Event](ts, func(message events.Event) bool {
		i, ok := message.(int)

		return ok && i%2 == 0
	})

	for i := 0; i < nm; i++ {
		if err := filter.Write(i); err != nil {
			t.Fatalf("unexpected error writing message: %v", err)
		}
	}

	checkClose(t, filter)
}
