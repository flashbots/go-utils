package main

//
// This example demonstrates how to use the rpcserver package to create a simple JSON-RPC server.
//
// It includes profiling test handlers, inspired by https://goperf.dev/02-networking/bench-and-load
//

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"time"

	_ "net/http/pprof"

	"github.com/flashbots/go-utils/rpcserver"
)

var (
	// Servers
	listenAddr = "localhost:8080"
	pprofAddr  = "localhost:6060"

	// Logger for the server
	log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Profiling utilities
	fastDelay     = flag.Duration("fast-delay", 0, "Fixed delay for fast handler (if any)")
	slowMin       = flag.Duration("slow-min", 1*time.Millisecond, "Minimum delay for slow handler")
	slowMax       = flag.Duration("slow-max", 300*time.Millisecond, "Maximum delay for slow handler")
	gcMinAlloc    = flag.Int("gc-min-alloc", 50, "Minimum number of allocations in GC heavy handler")
	gcMaxAlloc    = flag.Int("gc-max-alloc", 1000, "Maximum number of allocations in GC heavy handler")
	longLivedData [][]byte
)

func main() {
	handler, err := rpcserver.NewJSONRPCHandler(
		rpcserver.Methods{
			"test_foo": rpcHandlerTestFoo,
			"fast":     rpcHandlerFast,
			"slow":     rpcHandlerSlow,
			"gc":       rpcHandlerGCHeavy,
		},
		rpcserver.JSONRPCHandlerOpts{
			Log:                log,
			ServerName:         "public_server",
			GetResponseContent: []byte("static GET content hurray \\o/\n"),
		},
	)
	if err != nil {
		panic(err)
	}

	// Start separate pprof server
	go startPprofServer()

	// API server
	server := &http.Server{
		Addr:    listenAddr,
		Handler: handler,
	}
	fmt.Println("Starting server.", "listenAddr:", listenAddr)
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func startPprofServer() {
	fmt.Println("Starting pprof server.", "pprofAddr:", pprofAddr)
	if err := http.ListenAndServe(pprofAddr, nil); err != nil {
		fmt.Println("Error starting pprof server:", err)
	}
}

func randRange(min, max int) int {
	return rand.IntN(max-min) + min
}

func rpcHandlerTestFoo(ctx context.Context) (string, error) {
	return "foo", nil
}

func rpcHandlerFast(ctx context.Context) (string, error) {
	if *fastDelay > 0 {
		time.Sleep(*fastDelay)
	}

	return "fast response", nil
}

func rpcHandlerSlow(ctx context.Context) (string, error) {
	delayRange := int((*slowMax - *slowMin) / time.Millisecond)
	delay := time.Duration(randRange(1, delayRange)) * time.Millisecond
	time.Sleep(delay)

	return fmt.Sprintf("slow response with delay %d ms", delay.Milliseconds()), nil
}

func rpcHandlerGCHeavy(ctx context.Context) (string, error) {
	numAllocs := randRange(*gcMinAlloc, *gcMaxAlloc)
	var data [][]byte
	for i := 0; i < numAllocs; i++ {
		// Allocate 10KB slices. Occasionally retain a reference to simulate long-lived objects.
		b := make([]byte, 1024*10)
		data = append(data, b)
		if i%100 == 0 { // every 100 allocations, keep the data alive
			longLivedData = append(longLivedData, b)
		}
	}
	return fmt.Sprintf("allocated %d KB\n", len(data)*10), nil
}
