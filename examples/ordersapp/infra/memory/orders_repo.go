package memory

import (
	"context"
	"sync"

	"github.com/tangelo-labs/go-domain/examples/ordersapp/domain/order"
	"github.com/tangelo-labs/go-domain/examples/ordersapp/ucs/repos"
)

type ordersRepo struct {
	orders map[domain.ID]*order.Order
	mu     sync.RWMutex
}

// NewOrdersRepo creates a new Orders repository.
func NewOrdersRepo() repos.OrdersRepository {
	return &ordersRepo{
		orders: make(map[domain.ID]*order.Order),
	}
}

func (o *ordersRepo) Create(ctx context.Context, order *order.Order) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if _, exists := o.orders[order.ID()]; exists {
		return repos.ErrOrderAlreadyExists
	}

	o.orders[order.ID()] = order

	return nil
}

func (o *ordersRepo) Update(ctx context.Context, order *order.Order) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if _, exists := o.orders[order.ID()]; !exists {
		return repos.ErrOrderNotFound
	}

	o.orders[order.ID()] = order

	return nil
}

func (o *ordersRepo) FindByID(ctx context.Context, id domain.ID) (*order.Order, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if ord, exists := o.orders[id]; exists {
		return ord, nil
	}

	return nil, repos.ErrOrderNotFound
}
