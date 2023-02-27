package addline

import (
	"context"

	"github.com/tangelo-labs/go-domainkit"
	"github.com/tangelo-labs/go-domainkit/events"
	"github.com/tangelo-labs/go-domainkit/events/dispatcher"
	"github.com/tangelo-labs/go-domainkit/examples/ordersapp/domain/order"
	"github.com/tangelo-labs/go-domainkit/examples/ordersapp/ucs/repos"
)

// Handler sugar syntax for the UC.
type Handler interface {
	Handle(ctx context.Context, id domain.ID, lines []order.Line) error
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

func (h handler) Handle(ctx context.Context, id domain.ID, lines []order.Line) error {
	ord, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	for i := range lines {
		if aErr := ord.AddLine(lines[i]); aErr != nil {
			return aErr
		}
	}

	if uErr := h.repo.Update(ctx, ord); uErr != nil {
		return uErr
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
