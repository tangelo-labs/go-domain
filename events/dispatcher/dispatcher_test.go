package dispatcher_test

import (
	"context"
	"testing"
	"time"

	"github.com/Avalanche-io/counter"
	"github.com/stretchr/testify/require"
	"github.com/tangelo-labs/go-domainkit/events/dispatcher"
	"google.golang.org/protobuf/proto"
)

func TestMemoryDispatcher(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	t.Run("GIVEN an empty memory dispatcher", func(t *testing.T) {
		dpt := dispatcher.NewMemoryDispatcher()

		t.Run("WHEN subscribing a function that expect a private message pointer THEN subscription fails", func(t *testing.T) {
			sErr := dpt.Subscribe(ctx, func(ctx context.Context, msg *privateMessage) error {
				return nil
			})

			require.Error(t, sErr)
		})

		t.Run("WHEN subscribing a function that expect a string pointer THEN subscription fails", func(t *testing.T) {
			sErr := dpt.Subscribe(ctx, func(ctx context.Context, msg *string) error {
				return nil
			})

			require.Error(t, sErr)
		})

		t.Run("WHEN subscribing a function that expect an proto.Message interface THEN subscription fails", func(t *testing.T) {
			sErr := dpt.Subscribe(ctx, func(ctx context.Context, msg proto.Message) error {
				return nil
			})

			require.Error(t, sErr)
		})

		t.Run("WHEN subscribing a function that expect an empty interface THEN subscription fails", func(t *testing.T) {
			sErr := dpt.Subscribe(ctx, func(ctx context.Context, msg interface{}) error {
				return nil
			})

			require.Error(t, sErr)
		})
	})

	t.Run("GIVEN a memory dispatcher with one subscriber for string messages", func(t *testing.T) {
		dpt := dispatcher.NewMemoryDispatcher()
		rcv := counter.NewUnsigned()

		sErr := dpt.Subscribe(ctx, func(ctx context.Context, msg string) error {
			rcv.Add(1)

			return nil
		})

		require.NoError(t, sErr)

		t.Run("WHEN a string message is dispatched THEN subscriber receives the message", func(t *testing.T) {
			require.NoError(t, dpt.Dispatch(ctx, "test"))
			require.EqualValues(t, 1, rcv.Get())
		})
	})

	t.Run("GIVEN a memory dispatcher with one subscriber for privateMessage messages", func(t *testing.T) {
		dpt := dispatcher.NewMemoryDispatcher()
		rcv := counter.NewUnsigned()

		sErr := dpt.Subscribe(ctx, func(ctx context.Context, msg privateMessage) error {
			rcv.Add(1)

			return nil
		})

		require.NoError(t, sErr)

		t.Run("WHEN a private message is dispatched THEN subscriber receives the message", func(t *testing.T) {
			msg := privateMessage{Payload: gofakeit.LoremIpsumSentence(10)}

			require.NoError(t, dpt.Dispatch(ctx, msg))
			require.EqualValues(t, 1, rcv.Get())
		})
	})

	t.Run("GIVEN a memory dispatcher with one subscriber for string type and another for privateMessage", func(t *testing.T) {
		dpt := dispatcher.NewMemoryDispatcher()
		rcvPrivate := counter.NewUnsigned()
		rcvString := counter.NewUnsigned()

		sErr := dpt.Subscribe(ctx, func(ctx context.Context, msg privateMessage) error {
			rcvPrivate.Add(1)

			return nil
		})

		require.NoError(t, sErr)

		sErr = dpt.Subscribe(ctx, func(ctx context.Context, msg string) error {
			rcvString.Add(1)

			return nil
		})

		require.NoError(t, sErr)

		t.Run("WHEN a private message is dispatched THEN subscriber receives the message", func(t *testing.T) {
			msg := privateMessage{Payload: gofakeit.LoremIpsumSentence(10)}

			require.NoError(t, dpt.Dispatch(ctx, msg))
			require.EqualValues(t, 1, rcvPrivate.Get())
		})

		t.Run("WHEN a string message is dispatched THEN subscriber receives the message", func(t *testing.T) {
			msg := gofakeit.LoremIpsumSentence(10)

			require.NoError(t, dpt.Dispatch(ctx, msg))
			require.EqualValues(t, 1, rcvString.Get())
		})

		t.Run("WHEN a private message is dispatched by pointer THEN publishing does not fails AND message is not delivered", func(t *testing.T) {
			msg := &privateMessage{Payload: gofakeit.LoremIpsumSentence(10)}

			require.NoError(t, dpt.Dispatch(ctx, msg))
			require.EqualValues(t, 1, rcvString.Get())
		})
	})
}

type privateMessage struct {
	Payload string
}
