package httpadapter

import (
	"net/http"

	"github.com/tangelo-labs/go-domain/examples/ordersapp/ucs/create"
)

// CreateHandler adapts the UC to the HTTP transport.
func CreateHandler(uc create.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: out of the scope of this example. Here you should decode the
		// request input, call the UC and encode the response.
	}
}
