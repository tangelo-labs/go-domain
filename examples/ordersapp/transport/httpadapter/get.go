package httpadapter

import (
	"net/http"

	"github.com/tangelo-labs/go-domainkit/examples/ordersapp/ucs/get"
)

// GetHandler adapts the UC to the HTTP transport.
func GetHandler(uc get.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: out of the scope of this example. Here you should decode the
		// request input, call the UC and encode the response.
	}
}
