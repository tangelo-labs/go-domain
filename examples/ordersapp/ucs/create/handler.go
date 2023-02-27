package create

import (
	"context"

	"github.com/tangelo-labs/go-domain/events"
	"github.com/tangelo-labs/go-domain/events/dispatcher"
	"github.com/tangelo-labs/go-domain/examples/ordersapp/domain/order"
	"github.com/tangelo-labs/go-domain/examples/ordersapp/ucs/repos"
)

// Handler sugar syntax for the UC.
type Handler interface {
	Handle(ctx context.Context, id domain.ID, lines ...order.Line) error
}

type handler struct {
	repo repos.OrdersRepository
	dsp  dispatcher.Dispatcher
}

// NewHandler builds a new UC handler.
func NewHandler(repo repos.OrdersRepository, dsp dispatcher.Dispatcher) Handler {
	return handler{
		repo: repo,
		dsp:  dsp,
	}
}

func (h handler) Handle(ctx context.Context, id domain.ID, lines ...order.Line) error {
	ord := order.NewOrder(id)

	for i := range lines {
		if err := ord.AddLine(lines[i]); err != nil {
			return err
		}
	}

	if err := h.repo.Create(ctx, ord); err != nil {
		return err
	}

	defer h.dispatchEvents(ctx, ord)

	return nil
}

func (h handler) dispatchEvents(ctx context.Context, recorder events.Recorder) {
	for _, event := range recorder.Changes() {
		if err := h.dsp.Dispatch(ctx, event); err != nil {
			println("event dispatch failed:", err.Error())
		}
	}

	recorder.ClearChanges()
}
