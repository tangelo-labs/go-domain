package drain

import (
	"encoding/json"
	"fmt"

	"github.com/tangelo-labs/go-domain/events"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Marshaller converts an input event into a byte stream.
type Marshaller func(event events.Event) ([]byte, error)

// List of commonly used marshallers.
var (
	// JSONMarshaller a simple JSON marshaller function that uses the standard
	// library `encoding/json` package.
	JSONMarshaller = func(event events.Event) ([]byte, error) {
		return json.Marshal(event)
	}

	// ProtoMarshaller assumes the input event as `proto.Message`, and marshall
	// using `proto.Marshal()`.
	ProtoMarshaller = func(event events.Event) ([]byte, error) {
		if pb, ok := event.(proto.Message); ok {
			return proto.Marshal(pb)
		}

		return nil, fmt.Errorf("could not marshal (proto) event of type `%T`, not a proto message", event)
	}

	// AnyPBMarshaller assumes the input event as a `proto.Message`, and
	// marshall using `anypb` package. That is, it wraps the original proto
	// message in an `Any` message, and then marshall the `Any` message.
	AnyPBMarshaller = func(event events.Event) ([]byte, error) {
		if pb, ok := event.(proto.Message); ok {
			anyMsg, err := anypb.New(pb)
			if err != nil {
				return nil, fmt.Errorf("%w: could not marshal (proto+anypb) event", err)
			}

			return proto.Marshal(anyMsg)
		}

		return nil, fmt.Errorf("could not marshal (proto+anypb) event of type `%T`, not a proto message", event)
	}

	// ProtoJSONMarshaller similar to "AnyPBMarshaller". Assumes that the event
	// is a `proto.Message` instance, and marshall it using `protojson` package.
	ProtoJSONMarshaller = func(event events.Event) ([]byte, error) {
		if pb, ok := event.(proto.Message); ok {
			return protojson.Marshal(pb)
		}

		return nil, fmt.Errorf("could not marshal (proto+json) event of type `%T`, not a proto message", event)
	}
)
