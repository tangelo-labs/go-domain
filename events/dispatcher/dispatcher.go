// Package dispatcher provides a simple mechanism for dispatching domain events
// locally using a simple in-memory broker.
//
// Dispatcher must only be used to communicate between different parts of the
// same application or "Bounded Context". For communicating between different
// applications or services, please use a message broker instead.
package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/botchris/go-pubsub"
	"github.com/botchris/go-pubsub/provider/memory"
	"github.com/tangelo-labs/go-domain/events"
)

// ErrInvalidHandlerFunc returned when subscribing with an invalid handler function.
var ErrInvalidHandlerFunc = errors.New("invalid handler function")

var (
	contextType = reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
)

// Dispatcher defines a component capable of registering listeners and
// dispatching events to them.
type Dispatcher interface {
	// Dispatch dispatches an event to active subscribers. If this method fails,
	// means that the event was not delivered to any listener and could be
	// retried.
	Dispatch(ctx context.Context, event events.Event) error

	// Subscribe registers a handler function to react on specific events.
	// The handler function must have the following signature:
	//
	// 	func(ctx context.Context, event <T>) error
	//
	// Where <T> is the type of the event to be handled. This type cannot be
	// a pointer. This enforces the immutability of the events when dispatched.
	//
	// Examples:
	//
	// 	func(ctx context.Context, event string) error
	// 	func(ctx context.Context, event myMessage) error
	Subscribe(ctx context.Context, handlerFn interface{}) error
}

type memoryDispatcher struct {
	topic  pubsub.Topic
	broker pubsub.Broker
}

// NewMemoryDispatcher builds a dispatcher that moves events using local memory
// in a thread-safe way.
func NewMemoryDispatcher() Dispatcher {
	return &memoryDispatcher{
		topic:  "default",
		broker: memory.NewBroker(),
	}
}

func (m *memoryDispatcher) Dispatch(ctx context.Context, event events.Event) error {
	return m.broker.Publish(ctx, m.topic, event)
}

func (m *memoryDispatcher) Subscribe(ctx context.Context, handlerFn interface{}) error {
	handlerType, err := m.validateHandler(handlerFn)
	if err != nil {
		return err
	}

	handler := pubsub.NewHandler(func(ctx context.Context, t pubsub.Topic, catchAll interface{}) error {
		inputType := reflect.TypeOf(catchAll)
		if inputType != handlerType.In(1) {
			return nil
		}

		inputValue := reflect.ValueOf(catchAll)
		output := reflect.ValueOf(handlerFn).
			Call([]reflect.Value{
				reflect.ValueOf(ctx),
				inputValue,
			})

		if output[0].IsNil() {
			return nil
		}

		return output[0].Interface().(error)
	})

	_, err = m.broker.Subscribe(ctx, m.topic, handler)

	return err
}

func (m *memoryDispatcher) validateHandler(fn interface{}) (reflect.Type, error) {
	handlerType := reflect.TypeOf(fn)

	if handlerType.Kind() != reflect.Func {
		return nil, fmt.Errorf("%w: must be a function, got %T", ErrInvalidHandlerFunc, fn)
	}

	if handlerType.NumIn() != 2 {
		return nil, fmt.Errorf("%w: must have 2 input parameters", ErrInvalidHandlerFunc)
	}

	if handlerType.In(0) != contextType {
		return nil, fmt.Errorf("%w: must have a context.Context as first input parameter", ErrInvalidHandlerFunc)
	}

	if handlerType.NumOut() != 1 {
		return nil, fmt.Errorf("%w: must have 1 output parameter", ErrInvalidHandlerFunc)
	}

	if handlerType.Out(0) != errorType {
		return nil, fmt.Errorf("%w: must have an error as output parameter", ErrInvalidHandlerFunc)
	}

	if handlerType.In(1).Kind() == reflect.Ptr {
		return nil, fmt.Errorf("%w: expected event cannot be a pointer", ErrInvalidHandlerFunc)
	}

	if handlerType.In(1).Kind() == reflect.Interface {
		return nil, fmt.Errorf("%w: expected event cannot be an interface", ErrInvalidHandlerFunc)
	}

	return handlerType, nil
}
