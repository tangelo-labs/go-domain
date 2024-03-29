package drain_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"github.com/tangelo-labs/go-domain/events"
	"github.com/tangelo-labs/go-domain/events/drain"
	"go.uber.org/atomic"
)

func TestNewKinesisSink(t *testing.T) {
	t.Run("GIVEN a kinesis sink instance with an always-fail client AND a write-error counter", func(t *testing.T) {
		kinesisClient := &mockKinesisClient{
			err: fmt.Errorf("some error"),
		}

		marshallerCallsCount := atomic.NewInt64(0)
		marshaller := func(message events.Event) ([]byte, error) {
			marshallerCallsCount.Add(1)

			return drain.JSONMarshaller(message)
		}

		onErrorCount := atomic.NewInt64(0)
		onError := func(message events.Event, err error) {
			onErrorCount.Add(1)
		}

		sink, err := drain.NewKinesisSink[events.Event](gofakeit.UUID(), kinesisClient, marshaller, time.Second, onError)
		require.NoError(t, err)

		t.Run("WHEN a message fails to be sent", func(t *testing.T) {
			err = sink.Write(fakeMessage{ID: gofakeit.UUID()})
			require.ErrorIs(t, err, kinesisClient.err)

			t.Run("THEN marshaller is called AND error callback is invoked", func(t *testing.T) {
				require.EqualValues(t, 1, marshallerCallsCount.Load())
				require.EqualValues(t, 1, onErrorCount.Load())
			})
		})
	})

	t.Run("GIVEN a kinesis sink instance with an always-success client AND a write-error counter", func(t *testing.T) {
		kinesisClient := &mockKinesisClient{}

		marshallerCallsCount := atomic.NewInt64(0)
		marshaller := func(message events.Event) ([]byte, error) {
			marshallerCallsCount.Add(1)

			return drain.JSONMarshaller(message)
		}

		onErrorCount := atomic.NewInt64(0)
		onError := func(message events.Event, err error) {
			onErrorCount.Add(1)
		}

		sink, err := drain.NewKinesisSink[events.Event](gofakeit.UUID(), kinesisClient, marshaller, time.Second, onError)
		require.NoError(t, err)

		t.Run("WHEN a message is successfully sent", func(t *testing.T) {
			require.NoError(t, sink.Write(fakeMessage{ID: gofakeit.UUID()}))

			t.Run("THEN marshaller is called AND error callback is not invoked", func(t *testing.T) {
				require.EqualValues(t, 1, marshallerCallsCount.Load())
				require.EqualValues(t, 0, onErrorCount.Load())
			})
		})
	})
}

type fakeMessage struct {
	ID string
}

type mockKinesisClient struct {
	err error
}

func (m *mockKinesisClient) PutRecord(_ context.Context, _ *kinesis.PutRecordInput, _ ...func(*kinesis.Options)) (*kinesis.PutRecordOutput, error) {
	return &kinesis.PutRecordOutput{}, m.err
}
