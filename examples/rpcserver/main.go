package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/flashbots/go-utils/rpcserver"
)

var listenAddr = ":8080"

func main() {
	handler, err := rpcserver.NewJSONRPCHandler(
		rpcserver.Methods{
			"test_foo": HandleTestFoo,
		},
		rpcserver.JSONRPCHandlerOpts{
			ServerName:         "public_server",
			GetResponseContent: []byte("Hello world"),
		},
		nil,
	)
	if err != nil {
		panic(err)
	}

	// server
	server := &http.Server{
		Addr:    listenAddr,
		Handler: handler,
	}
	fmt.Println("Starting server.", "listenAddr:", listenAddr)
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func HandleTestFoo(ctx context.Context) (string, error) {
	return "foo", nil
}
