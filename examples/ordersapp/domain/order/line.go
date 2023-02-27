package order

import (
	"errors"
	"fmt"
)

// ErrInvalidLine is returned when a line is invalid.
var ErrInvalidLine = errors.New("invalid order line")

// Line is a line in an order.
type Line struct {
	ProductID string
	Quantity  int
	UnitPrice float64
}

// Validate validates the line.
func (l Line) Validate() error {
	if l.ProductID == "" {
		return fmt.Errorf("%w: product ID is required", ErrInvalidLine)
	}

	if l.Quantity <= 0 {
		return fmt.Errorf("%w: quantity must be greater than zero", ErrInvalidLine)
	}

	if l.UnitPrice <= 0 {
		return fmt.Errorf("%w: unit price must be greater than zero", ErrInvalidLine)
	}

	return nil
}
