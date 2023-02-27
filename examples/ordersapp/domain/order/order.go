package order

import (
	"github.com/tangelo-labs/go-domain"
	"github.com/tangelo-labs/go-domain/events"
)

// Order is a products order.
type Order struct {
	id    domain.ID
	lines []Line

	events.BaseRecorder
}

// View is a read-only projection of an order.
type View struct {
	ID    domain.ID
	Lines []Line
}

// NewOrder creates a new empty order.
func NewOrder(id domain.ID) *Order {
	ord := &Order{
		id:    id,
		lines: make([]Line, 0),
	}

	ord.Record(CreatedEvent{
		OrderID: id,
	})

	return ord
}

// ID returns the order ID.
func (o *Order) ID() domain.ID {
	return o.id
}

// AddLine adds a line to the order.
func (o *Order) AddLine(line Line) error {
	if err := line.Validate(); err != nil {
		return err
	}

	merged := false

	for i := range o.lines {
		if o.lines[i].ProductID == line.ProductID && o.lines[i].UnitPrice == line.UnitPrice {
			o.lines[i].Quantity += line.Quantity
			merged = true

			break
		}
	}

	if !merged {
		o.lines = append(o.lines, line)
	}

	o.Record(LineAddedEvent{
		OrderID: o.id,
		Line:    line,
	})

	return nil
}

// Total returns the total amount of the order.
func (o *Order) Total() float64 {
	total := float64(0)

	for i := range o.lines {
		total += float64(o.lines[i].Quantity) * o.lines[i].UnitPrice
	}

	return total
}

// View returns a read-only projection of the order.
func (o *Order) View() View {
	lines := make([]Line, len(o.lines))
	copy(lines, o.lines)

	return View{
		ID:    o.id,
		Lines: lines,
	}
}
