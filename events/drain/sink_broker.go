package drain

import (
	"context"
	"fmt"
	"time"

	"github.com/botchris/go-pubsub"
)

// BrokerSinkOpts is a type for broker sink options.
type BrokerSinkOpts func(config *brokerSinkConfig)

type brokerSink[M any] struct {
	*baseSink
	broker pubsub.Broker
	topic  pubsub.Topic
	conf   brokerSinkConfig
}

type brokerSinkConfig struct {
	timeout time.Duration
}

// NewBrokerSink creates a new broker sink.
func NewBrokerSink[M any](broker pubsub.Broker, topic pubsub.Topic, opts ...BrokerSinkOpts) (Sink[M], error) {
	// defaults.
	conf := &brokerSinkConfig{
		timeout: 5 * time.Second,
	}

	for _, opt := range opts {
		opt(conf)
	}

	if broker == nil {
		return nil, fmt.Errorf("a valid broker must be provided")
	}

	if topic == "" {
		return nil, fmt.Errorf("a valid topic must be provided")
	}

	if conf.timeout <= 0 {
		return nil, fmt.Errorf("a timeout greater than 0 must be provided")
	}

	return &brokerSink[M]{
		baseSink: newCloseTrait(),
		broker:   broker,
		topic:    topic,
		conf:     *conf,
	}, nil
}

func (s *brokerSink[M]) Write(message M) error {
	if s.baseSink.IsClosed() {
		return fmt.Errorf("%w: broker writer sink could not publish message %T", ErrSinkClosed, message)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.conf.timeout)
	defer cancel()

	if err := s.broker.Publish(ctx, s.topic, message); err != nil {
		return fmt.Errorf("%w: broker writer sink could not publish message %T", err, message)
	}

	return nil
}

// BrokerSinkWithTimeout sets the timeout for the broker sink.
func BrokerSinkWithTimeout(timeout time.Duration) BrokerSinkOpts {
	return func(config *brokerSinkConfig) {
		config.timeout = timeout
	}
}
