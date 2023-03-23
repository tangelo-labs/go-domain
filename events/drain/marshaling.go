package drain

import (
	"encoding/json"
	"fmt"

	"github.com/tangelo-labs/go-domain/events"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Marshaller converts an input message into a byte stream.
type Marshaller[M any] func(M) ([]byte, error)

// List of commonly used marshallers.
var (
	// JSONMarshaller a simple JSON marshaller function that uses the standard
	// library `encoding/json` package to marshal events.Event messages.
	JSONMarshaller = func(m events.Event) ([]byte, error) {
		return json.Marshal(m)
	}

	// ProtoMarshaller assumes the input message as `proto.Message`, and marshall
	// using `proto.Marshal()`
	ProtoMarshaller = func(m events.Event) ([]byte, error) {
		if pb, ok := m.(proto.Message); ok {
			return proto.Marshal(pb)
		}

		return nil, fmt.Errorf("could not marshal (proto) message of type `%T`, not a proto message", m)
	}

	// AnyPBMarshaller assumes the input message as a `proto.Message`, and
	// marshall using `anypb` package. That is, it wraps the original proto
	// message in an `Any` message, and then marshall the `Any` message.
	AnyPBMarshaller = func(m events.Event) ([]byte, error) {
		if pb, ok := m.(proto.Message); ok {
			anyMsg, err := anypb.New(pb)
			if err != nil {
				return nil, fmt.Errorf("%w: could not marshal (proto+anypb) message", err)
			}

			return proto.Marshal(anyMsg)
		}

		return nil, fmt.Errorf("could not marshal (proto+anypb) message of type `%T`, not a proto message", m)
	}

	// ProtoJSONMarshaller similar to "AnyPBMarshaller". Assumes that the message
	// is a `proto.Message` instance, and marshall it using `protojson` package.
	ProtoJSONMarshaller = func(m events.Event) ([]byte, error) {
		if pb, ok := m.(proto.Message); ok {
			return protojson.Marshal(pb)
		}

		return nil, fmt.Errorf("could not marshal (proto+json) message of type `%T`, not a proto message", m)
	}
)
