package repos

import (
	"context"
	"errors"

	"github.com/tangelo-labs/go-domainkit"
	"github.com/tangelo-labs/go-domainkit/examples/ordersapp/domain/order"
)

// Repo-related errors.
var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderAlreadyExists = errors.New("order already exists")
)

// OrdersRepository allows to access orders.
type OrdersRepository interface {
	Create(ctx context.Context, order *order.Order) error
	Update(ctx context.Context, order *order.Order) error
	FindByID(ctx context.Context, id domain.ID) (*order.Order, error)
}
