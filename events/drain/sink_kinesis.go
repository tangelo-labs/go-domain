package drain

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/tangelo-labs/go-domain"
)

// KinesisAPI represents a Kinesis client for sending messages.
type KinesisAPI interface {
	PutRecord(ctx context.Context, params *kinesis.PutRecordInput, optFns ...func(*kinesis.Options)) (*kinesis.PutRecordOutput, error)
}

type kinesisSink[M any] struct {
	*baseSink
	streamName string
	kinesis    KinesisAPI
	marshaller Marshaller[M]
	timeout    time.Duration
	onError    WriteErrorFn[M]
}

// NewKinesisSink builds a new sink that sends messages to a Kinesis Stream.
func NewKinesisSink[M any](
	streamName string,
	api KinesisAPI,
	marshaller Marshaller[M],
	timeout time.Duration,
	onError WriteErrorFn[M],
) (Sink[M], error) {
	if streamName == "" {
		return nil, fmt.Errorf("a kinesis stream name must be provided")
	}

	if api == nil {
		return nil, fmt.Errorf("a valid kinesis client must be provided")
	}

	if marshaller == nil {
		return nil, fmt.Errorf("a valid marshaller function must be provided")
	}

	if timeout < time.Second {
		return nil, fmt.Errorf("a timeout greater or equal than 1 second must be provided, got %s", timeout)
	}

	if onError == nil {
		onError = noopWriteError[M]
	}

	return &kinesisSink[M]{
		baseSink:   newCloseTrait(),
		streamName: streamName,
		kinesis:    api,
		marshaller: marshaller,
		timeout:    timeout,
		onError:    onError,
	}, nil
}

func (k *kinesisSink[M]) Write(message M) error {
	if k.baseSink.IsClosed() {
		return fmt.Errorf("%w: writer sink could not write message %T", ErrSinkClosed, message)
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), k.timeout)
	defer cancelFunc()

	data, err := k.marshaller(message)
	if err != nil {
		return err
	}

	if _, err = k.kinesis.PutRecord(ctx, &kinesis.PutRecordInput{
		StreamName:   aws.String(k.streamName),
		PartitionKey: aws.String(domain.NewID().String()),
		Data:         data,
	}); err != nil {
		if k.onError != nil {
			k.onError(message, err)
		}

		return err
	}

	return nil
}
