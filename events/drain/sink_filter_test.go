package drain_test

import (
	"testing"

	"github.com/tangelo-labs/go-domainkit/events"
	"github.com/tangelo-labs/go-domainkit/events/drain"
)

func TestFilter(t *testing.T) {
	const nevents = 100
	ts := newTestSink(t, nevents/2)
	filter := drain.NewFilter(ts, func(event events.Event) bool {
		i, ok := event.(int)

		return ok && i%2 == 0
	})

	for i := 0; i < nevents; i++ {
		if err := filter.Write(i); err != nil {
			t.Fatalf("unexpected error writing event: %v", err)
		}
	}

	checkClose(t, filter)
}
