package get

import (
	"context"

	"github.com/tangelo-labs/go-domain/examples/ordersapp/domain/order"
	"github.com/tangelo-labs/go-domain/examples/ordersapp/ucs/repos"
)

// Handler sugar syntax for the UC.
type Handler interface {
	Handle(ctx context.Context, id domain.ID) (order.View, error)
}

type handler struct {
	repo repos.OrdersRepository
}

// NewHandler builds a new UC handler.
func NewHandler(repo repos.OrdersRepository) Handler {
	return handler{
		repo: repo,
	}
}

func (h handler) Handle(ctx context.Context, id domain.ID) (order.View, error) {
	ord, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return order.View{}, err
	}

	return ord.View(), nil
}
