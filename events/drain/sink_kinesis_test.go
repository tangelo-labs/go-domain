package drain_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
	"github.com/tangelo-labs/go-domainkit/events"
	"github.com/tangelo-labs/go-domainkit/events/drain"
	"go.uber.org/atomic"
)

func TestNewKinesisSink(t *testing.T) {
	t.Run("GIVEN a kinesis sink instance with an always-fail client AND a write-error counter", func(t *testing.T) {
		kinesisClient := &mockKinesisClient{
			err: fmt.Errorf("some error"),
		}

		marshallerCallsCount := atomic.NewInt64(0)
		marshaller := func(event events.Event) ([]byte, error) {
			marshallerCallsCount.Add(1)

			return drain.JSONMarshaller(event)
		}

		onErrorCount := atomic.NewInt64(0)
		onError := func(event events.Event, err error) {
			onErrorCount.Add(1)
		}

		sink, err := drain.NewKinesisSink(gofakeit.UUID(), kinesisClient, marshaller, time.Second, onError)
		require.NoError(t, err)

		t.Run("WHEN an event fails to be sent", func(t *testing.T) {
			err = sink.Write(fakeEvent{ID: gofakeit.UUID()})
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
		marshaller := func(event events.Event) ([]byte, error) {
			marshallerCallsCount.Add(1)

			return drain.JSONMarshaller(event)
		}

		onErrorCount := atomic.NewInt64(0)
		onError := func(event events.Event, err error) {
			onErrorCount.Add(1)
		}

		sink, err := drain.NewKinesisSink(gofakeit.UUID(), kinesisClient, marshaller, time.Second, onError)
		require.NoError(t, err)

		t.Run("WHEN an event is successfully sent", func(t *testing.T) {
			require.NoError(t, sink.Write(fakeEvent{ID: gofakeit.UUID()}))

			t.Run("THEN marshaller is called AND error callback is not invoked", func(t *testing.T) {
				require.EqualValues(t, 1, marshallerCallsCount.Load())
				require.EqualValues(t, 0, onErrorCount.Load())
			})
		})
	})
}

type fakeEvent struct {
	ID string
}

type mockKinesisClient struct {
	err error
}

func (m *mockKinesisClient) PutRecord(_ context.Context, _ *kinesis.PutRecordInput, _ ...func(*kinesis.Options)) (*kinesis.PutRecordOutput, error) {
	return &kinesis.PutRecordOutput{}, m.err
}
