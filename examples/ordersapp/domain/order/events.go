package order

import "github.com/tangelo-labs/go-domainkit"

// CreatedEvent a domain event that is raised when an order is created.
type CreatedEvent struct {
	OrderID domain.ID
}

// LineAddedEvent a domain event that is raised when a line is added to an order.
type LineAddedEvent struct {
	OrderID domain.ID
	Line    Line
}
