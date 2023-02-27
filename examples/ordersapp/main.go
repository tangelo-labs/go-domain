// Package main provides a simple application example that uses the different
// "go-domain" package definitions.
package main

import (
	"net/http"

	"github.com/tangelo-labs/go-domain/events/dispatcher"
	"github.com/tangelo-labs/go-domain/examples/ordersapp/infra/memory"
	"github.com/tangelo-labs/go-domain/examples/ordersapp/transport/httpadapter"
	"github.com/tangelo-labs/go-domain/examples/ordersapp/ucs/addline"
	"github.com/tangelo-labs/go-domain/examples/ordersapp/ucs/create"
	"github.com/tangelo-labs/go-domain/examples/ordersapp/ucs/get"
)

func main() {
	repo := memory.NewOrdersRepo()
	disp := dispatcher.NewMemoryDispatcher()

	createUC := create.NewHandler(repo, disp)
	addLineUC := addline.NewHandler(repo, disp)
	getUC := get.NewHandler(repo)

	// GET /orders/{id}
	http.HandleFunc("/orders/{id}", httpadapter.GetHandler(getUC))

	// PUT /orders
	http.HandleFunc("/orders", httpadapter.CreateHandler(createUC))

	// POST /orders/{id}/lines
	http.HandleFunc("/orders/{id}/lines", httpadapter.AddLineHandler(addLineUC))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
