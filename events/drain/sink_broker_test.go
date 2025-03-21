package drain_test

import (
	"context"
	"testing"
	"time"

	"github.com/botchris/go-pubsub"
	"github.com/botchris/go-pubsub/provider/memory"
	"github.com/stretchr/testify/require"
	"github.com/tangelo-labs/go-domain/events"
	"github.com/tangelo-labs/go-domain/events/drain"
)

func Test_BrokerSink(t *testing.T) {
	t.Run("GIVEN a broker sink constructor WHEN we input invalid broker", func(t *testing.T) {
		_, err := drain.NewBrokerSink[events.Event](nil, "topic")

		t.Run("THEN returns error", func(t *testing.T) {
			require.Error(t, err)
		})
	})

	t.Run("GIVEN a broker sink constructor WHEN we input invalid topic", func(t *testing.T) {
		_, err := drain.NewBrokerSink[events.Event](nil, "topic")

		t.Run("THEN returns error", func(t *testing.T) {
			require.Error(t, err)
		})
	})

	t.Run("GIVEN a broker sink constructor WHEN we input invalid timeout", func(t *testing.T) {
		_, err := drain.NewBrokerSink[events.Event](nil, "topic", drain.BrokerSinkWithTimeout(0))

		t.Run("THEN returns error", func(t *testing.T) {
			require.Error(t, err)
		})
	})

	t.Run("GIVEN a closed broker sink", func(t *testing.T) {
		brk := memory.NewBroker()

		sink, err := drain.NewBrokerSink[events.Event](brk, "topic")
		require.NoError(t, err)

		err = sink.Close()
		require.NoError(t, err)

		t.Run("WHEN we write a message", func(t *testing.T) {
			err := sink.Write(struct{}{})

			t.Run("THEN the sink should not accept messages", func(t *testing.T) {
				require.ErrorIs(t, err, drain.ErrSinkClosed)
			})
		})
	})

	t.Run("GIVEN a broker sink", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		brk := memory.NewBroker()

		sink, err := drain.NewBrokerSink[events.Event](brk, "topic")
		require.NoError(t, err)

		published := false

		_, err = brk.Subscribe(ctx, "topic", pubsub.NewHandler(func(ctx context.Context, topic pubsub.Topic, msg struct{}) error {
			published = true

			return nil
		}))
		require.NoError(t, err)

		t.Run("WHEN we write a message", func(t *testing.T) {
			err := sink.Write(struct{}{})

			t.Run("THEN the message is published to the topic", func(t *testing.T) {
				require.NoError(t, err)
				require.True(t, published)
			})
		})
	})
}
