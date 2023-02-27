package drain

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tangelo-labs/go-domain/events"
)

// retryingSink retries the writing until success or an ErrSinkClosed is
// returned. Underlying sink must have p > 0 of succeeding or the sink will
// block. Retry is configured with a RetrySinkStrategy. Concurrent calls to a
// retrying sink are serialized through the sink, meaning that if one is
// in-flight, another will not proceed.
type retryingSink struct {
	*baseSink
	sink         Sink
	strategy     RetrySinkStrategy
	dropHandling WriteErrorFn
}

// NewRetrying returns a sink that will retry writes to a sink, backing
// off on failure. Parameters threshold and backoff adjust the behavior of the
// circuit breaker.
func NewRetrying(sink Sink, strategy RetrySinkStrategy, dropHandling WriteErrorFn) Sink {
	dh := dropHandling
	if dh == nil {
		dh = noopWriteError
	}

	rs := &retryingSink{
		baseSink:     newBaseSink(),
		sink:         sink,
		strategy:     strategy,
		dropHandling: dh,
	}

	return rs
}

// Write attempts to flush the events to the downstream sink until it succeeds
// or the sink is closed.
func (rs *retryingSink) Write(event events.Event) error {
retry:
	if rs.baseSink.IsClosed() {
		return fmt.Errorf("%w: retrying sink could not write event %T", ErrSinkClosed, event)
	}

	if backoff := rs.strategy.Proceed(event); backoff > 0 {
		select {
		case <-time.After(backoff):
			// TODO(stevvooe): This branch holds up the next try. Before, we
			// would simply break to the "retry" label and then possibly wait
			// again. However, this requires all retry strategies to have a
			// large probability of probing the sync for success, rather than
			// just backing off and sending the request.
		case <-rs.Closed():
			return ErrSinkClosed
		}
	}

	if err := rs.sink.Write(event); err != nil {
		if errors.Is(err, ErrSinkClosed) {
			// terminal!
			return err
		}

		if rs.strategy.Failure(event, err) {
			rs.dropHandling(event, err)

			return nil
		}

		goto retry
	}

	rs.strategy.Success(event)

	return nil
}

// Close closes the sink and the underlying sink.
func (rs *retryingSink) Close() error {
	if errS := rs.sink.Close(); errS != nil {
		return fmt.Errorf("%w: retrying sink could not close underlying sink", errS)
	}

	if errB := rs.baseSink.Close(); errB != nil {
		return fmt.Errorf("%w: retrying sink could not close", errB)
	}

	return nil
}

// RetrySinkStrategy defines a strategy for retrying event sink writes.
//
// All methods should be goroutine safe.
type RetrySinkStrategy interface {
	// Proceed is called before every event send. If proceed returns a
	// positive, non-zero integer, the retryer will back off by the provided
	// duration.
	//
	// An event is provided, by may be ignored.
	Proceed(event events.Event) time.Duration

	// Failure reports a failure to the strategy. If this method returns true,
	// the event should be dropped.
	Failure(event events.Event, err error) bool

	// Success should be called when an event is sent successfully.
	Success(event events.Event)
}

// NewBreakerStrategy returns a breaker that will backoff after the threshold has been
// tripped. A Breaker is thread safe and may be shared by many goroutines.
func NewBreakerStrategy(threshold int, backoff time.Duration) RetrySinkStrategy {
	return &breakerStrategy{
		threshold: threshold,
		backoff:   backoff,
	}
}

// Breaker implements a circuit breaker retry strategy.
//
// The current implementation never drops events.
type breakerStrategy struct {
	threshold int
	recent    int
	last      time.Time
	backoff   time.Duration // time after which we retry after failure.
	mu        sync.Mutex
}

// Proceed checks the failures against the threshold.
func (b *breakerStrategy) Proceed(event events.Event) time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.recent < b.threshold {
		return 0
	}

	return time.Until(b.last.Add(b.backoff))
}

// Success resets the breaker.
func (b *breakerStrategy) Success(event events.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.recent = 0
	b.last = time.Time{}
}

// Failure records the failure and latest failure time.
func (b *breakerStrategy) Failure(event events.Event, err error) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.recent++
	b.last = time.Now().UTC()

	return false // never drop events.
}

// ExponentialBackoffConfig configures backoff parameters.
//
// Note that these parameters operate on the upper bound for choosing a random
// value. For example, at Base=1s, a random value in [0,1s) will be chosen for
// the backoff value.
type ExponentialBackoffConfig struct {
	// Base is the minimum bound for backing off after failure.
	Base time.Duration

	// Factor sets the amount of time by which the backoff grows with each
	// failure.
	Factor time.Duration

	// Max is the absolute maxiumum bound for a single backoff.
	Max time.Duration
}

// DefaultExponentialBackoffConfig provides a default configuration for
// exponential backoff.
var DefaultExponentialBackoffConfig = ExponentialBackoffConfig{
	Base:   time.Second,
	Factor: time.Second,
	Max:    20 * time.Second,
}

// NewExponentialBackoff returns an exponential backoff strategy with the
// desired config. If config is nil, the default is returned.
func NewExponentialBackoff(config ExponentialBackoffConfig) RetrySinkStrategy {
	return &exponentialBackoffStrategy{
		config: config,
	}
}

// exponentialBackoffStrategy implements random backoff with exponentially increasing
// bounds as the number consecutive failures increase.
type exponentialBackoffStrategy struct {
	failures uint64 // consecutive failure counter (needs to be 64-bit aligned)
	config   ExponentialBackoffConfig
}

// Proceed returns the next randomly bound exponential backoff time.
func (b *exponentialBackoffStrategy) Proceed(event events.Event) time.Duration {
	return b.backoff(atomic.LoadUint64(&b.failures))
}

// Success resets the failures counter.
func (b *exponentialBackoffStrategy) Success(event events.Event) {
	atomic.StoreUint64(&b.failures, 0)
}

// Failure increments the failure counter.
func (b *exponentialBackoffStrategy) Failure(event events.Event, err error) bool {
	atomic.AddUint64(&b.failures, 1)

	return false
}

// backoff calculates the amount of time to wait based on the number of
// consecutive failures.
func (b *exponentialBackoffStrategy) backoff(failures uint64) time.Duration {
	if failures <= 0 {
		// proceed normally when there are no failures.
		return 0
	}

	factor := b.config.Factor
	if factor <= 0 {
		factor = DefaultExponentialBackoffConfig.Factor
	}

	backoff := b.config.Base + factor*time.Duration(1<<(failures-1))

	max := b.config.Max
	if max <= 0 {
		max = DefaultExponentialBackoffConfig.Max
	}

	if backoff > max || backoff < 0 {
		backoff = max
	}

	// Choose a uniformly distributed value from [0, backoff).
	return time.Duration(rand.Int63n(int64(backoff)))
}
