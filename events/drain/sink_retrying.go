package drain

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// retryingSink retries the writing until success or an ErrSinkClosed is
// returned. Underlying sink must have p > 0 of succeeding or the sink will
// block. Retry is configured with a RetrySinkStrategy. Concurrent calls to a
// retrying sink are serialized through the sink, meaning that if one is
// in-flight, another will not proceed.
type retryingSink[M any] struct {
	*baseSink
	sink         Sink[M]
	strategy     RetrySinkStrategy[M]
	dropHandling WriteErrorFn[M]
}

// NewRetrying returns a sink that will retry writes to a sink, backing
// off on failure. Parameters threshold and backoff adjust the behavior of the
// circuit breaker.
func NewRetrying[M any](sink Sink[M], strategy RetrySinkStrategy[M], dropHandling WriteErrorFn[M]) Sink[M] {
	dh := dropHandling
	if dh == nil {
		dh = noopWriteError[M]
	}

	rs := &retryingSink[M]{
		baseSink:     newCloseTrait(),
		sink:         sink,
		strategy:     strategy,
		dropHandling: dh,
	}

	return rs
}

// Write attempts to flush the messages to the downstream sink until it succeeds
// or the sink is closed.
func (rs *retryingSink[M]) Write(message M) error {
retry:
	if rs.baseSink.IsClosed() {
		return fmt.Errorf("%w: retrying sink could not write message %T", ErrSinkClosed, message)
	}

	if backoff := rs.strategy.Proceed(message); backoff > 0 {
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

	if err := rs.sink.Write(message); err != nil {
		if errors.Is(err, ErrSinkClosed) {
			// terminal!
			return err
		}

		if rs.strategy.Failure(message, err) {
			rs.dropHandling(message, err)

			return nil
		}

		goto retry
	}

	rs.strategy.Success(message)

	return nil
}

// Close closes the sink and the underlying sink.
func (rs *retryingSink[M]) Close() error {
	if errS := rs.sink.Close(); errS != nil {
		return fmt.Errorf("%w: retrying sink could not close underlying sink", errS)
	}

	if errB := rs.baseSink.Close(); errB != nil {
		return fmt.Errorf("%w: retrying sink could not close", errB)
	}

	return nil
}

// RetrySinkStrategy defines a strategy for retrying message sink writes.
//
// All methods should be goroutine safe.
type RetrySinkStrategy[M any] interface {
	// Proceed is called before every message send. If proceed returns a
	// positive, non-zero integer, the retryer will back off by the provided
	// duration.
	//
	// A message is provided, by may be ignored.
	Proceed(M) time.Duration

	// Failure reports a failure to the strategy. If this method returns true,
	// the message should be dropped.
	Failure(M, error) bool

	// Success should be called when a message is sent successfully.
	Success(M)
}

// NewBreakerStrategy returns a breaker that will backoff after the threshold has been
// tripped. A Breaker is thread safe and may be shared by many goroutines.
func NewBreakerStrategy[M any](threshold int, backoff time.Duration) RetrySinkStrategy[M] {
	return &breakerStrategy[M]{
		threshold: threshold,
		backoff:   backoff,
	}
}

// Breaker implements a circuit breaker retry strategy.
//
// The current implementation never drops messages.
type breakerStrategy[M any] struct {
	threshold int
	recent    int
	last      time.Time
	backoff   time.Duration // time after which we retry after failure.
	mu        sync.Mutex
}

// Proceed checks the failures against the threshold.
func (b *breakerStrategy[M]) Proceed(M) time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.recent < b.threshold {
		return 0
	}

	return time.Until(b.last.Add(b.backoff))
}

// Success resets the breaker.
func (b *breakerStrategy[M]) Success(M) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.recent = 0
	b.last = time.Time{}
}

// Failure records the failure and latest failure time.
func (b *breakerStrategy[M]) Failure(M, error) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.recent++
	b.last = time.Now().UTC()

	return false // never drop messages.
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
func NewExponentialBackoff[M any](config ExponentialBackoffConfig) RetrySinkStrategy[M] {
	return &exponentialBackoffStrategy[M]{
		config: config,
	}
}

// exponentialBackoffStrategy implements random backoff with exponentially increasing
// bounds as the number consecutive failures increase.
type exponentialBackoffStrategy[M any] struct {
	failures uint64 // consecutive failure counter (needs to be 64-bit aligned)
	config   ExponentialBackoffConfig
}

// Proceed returns the next randomly bound exponential backoff time.
func (b *exponentialBackoffStrategy[M]) Proceed(M) time.Duration {
	return b.backoff(atomic.LoadUint64(&b.failures))
}

// Success resets the failures counter.
func (b *exponentialBackoffStrategy[M]) Success(M) {
	atomic.StoreUint64(&b.failures, 0)
}

// Failure increments the failure counter.
func (b *exponentialBackoffStrategy[M]) Failure(M, error) bool {
	atomic.AddUint64(&b.failures, 1)

	return false
}

// backoff calculates the amount of time to wait based on the number of
// consecutive failures.
func (b *exponentialBackoffStrategy[M]) backoff(failures uint64) time.Duration {
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
