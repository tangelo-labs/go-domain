package domain

import "github.com/tangelo-labs/go-domainkit/events"

// Entity is the base interface for all domain entities.
type Entity interface {
	ID() ID
}

// AggregateRoot is the base interface for all domain aggregate roots.
type AggregateRoot interface {
	Entity
	events.Recorder
}
