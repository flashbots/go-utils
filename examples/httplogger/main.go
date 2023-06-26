package main

import (
	"errors"
	"flag"
	"net/http"

	"github.com/ethereum/go-ethereum/log"
	"github.com/flashbots/go-utils/httplogger"
	"github.com/flashbots/go-utils/logutils"
	"go.uber.org/zap"
)

var (
	listenAddr = "localhost:8124"
)

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!"))
	w.WriteHeader(http.StatusOK)
}

func ErrorHandler(w http.ResponseWriter, r *http.Request) {
	l := logutils.ZapFromRequest(r)
	l.Error("this is an error", zap.Error(errors.New("testError")))
	http.Error(w, "this is an error", http.StatusInternalServerError)
}

func PanicHandler(w http.ResponseWriter, r *http.Request) {
	panic("foo!")
}

func main() {
	logLevel := flag.String("log-level", "info", "Log level")
	logDev := flag.Bool("log-dev", false, "Log in development mode")
	flag.Parse()

	l := logutils.GetZapLogger(
		logutils.LogDevMode(*logDev),
		logutils.LogLevel(*logLevel),
	)
	defer logutils.FlushZap(l)

	l.Info("Webserver running at " + listenAddr)

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", HelloHandler)
	mux.HandleFunc("/error", ErrorHandler)
	mux.HandleFunc("/panic", PanicHandler)
	loggedRouter := httplogger.LoggingMiddlewareZap(l, mux)

	if err := http.ListenAndServe(listenAddr, loggedRouter); err != nil {
		log.Crit("webserver failed", "err", err)
	}
}
