package drain_test

import (
	"sync"
	"testing"
	"time"

	"github.com/tangelo-labs/go-domain/events"
	"github.com/tangelo-labs/go-domain/events/drain"
)

func TestRetryingSinkBreaker(t *testing.T) {
	testRetryingSink(t, drain.NewBreakerStrategy[events.Event](3, 10*time.Millisecond))
}

func TestRetryingSinkExponentialBackoff(t *testing.T) {
	testRetryingSink(t, drain.NewExponentialBackoff[events.Event](drain.ExponentialBackoffConfig{
		Base:   time.Millisecond,
		Factor: time.Millisecond,
		Max:    time.Millisecond * 5,
	}))
}

func testRetryingSink[M any](t *testing.T, strategy drain.RetrySinkStrategy[M]) {
	const nm = 100

	ts := newTestSink[M](t, nm)

	// Make a sync that fails most of the time, ensuring that all the messages
	// make it through.
	flaky := &flakySink[M]{
		rate: 1.0, // start out always failing.
		Sink: ts,
	}

	s := drain.NewRetrying[M](flaky, strategy, nil)

	var wg sync.WaitGroup

	for i := 1; i <= nm; i++ {
		var m M

		// Above 50, set the failure rate lower
		if i > 50 {
			flaky.mu.Lock()
			flaky.rate = 0.9
			flaky.mu.Unlock()
		}

		wg.Add(1)

		go func(m M) {
			defer wg.Done()

			if err := s.Write(m); err != nil {
				t.Errorf("error writing message: %v", err)
			}
		}(m)
	}

	wg.Wait()
	checkClose(t, s)
}

func TestExponentialBackoff(t *testing.T) {
	config := drain.DefaultExponentialBackoffConfig
	strategy := drain.NewExponentialBackoff[events.Event](config)
	backoff := strategy.Proceed(nil)

	if backoff != 0 {
		t.Errorf("untouched backoff should be zero-wait: %v != 0", backoff)
	}

	expected := config.Base + config.Factor

	for i := 1; i <= 10; i++ {
		if strategy.Failure(nil, nil) {
			t.Errorf("no facilities for dropping messages in ExponentialBackoffStrategy")
		}

		for j := 0; j < 1000; j++ {
			// sample this several thousand times.
			if bo := strategy.Proceed(nil); bo > expected {
				t.Fatalf("expected must be bounded by %v after %v failures: %v", expected, i, bo)
			}
		}

		expected = config.Base + config.Factor*time.Duration(1<<uint64(i))
		if expected > config.Max {
			expected = config.Max
		}
	}

	strategy.Success(nil) // recovery!

	backoff = strategy.Proceed(nil)
	if backoff != 0 {
		t.Errorf("should have recovered: %v != 0", backoff)
	}
}
