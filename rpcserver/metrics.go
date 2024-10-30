package rpcserver

import (
	"fmt"

	"github.com/VictoriaMetrics/metrics"
)

var (
	// incremented when user made incorrect request
	incorrectRequestCounter = metrics.NewCounter("goutils_rpcserver_incorrect_request_total")
	// incremented when server has a bug (e.g. can't marshall response)
	internalErrorsCounter = metrics.NewCounter("goutils_rpcserver_internal_errors_total")
)

const (
	// we use unknown method label for methods that server does not support because otherwise
	// users can create arbitrary number of metrics
	unknownMethodLabel = "unknown"

	// incremented when request comes in
	requestCountLabel = `goutils_rpcserver_request_count{method="%s"}`
	// incremented when handler method returns JSONRPC error
	errorCountLabel = `goutils_rpcserver_error_count{method="%s"}`
	// total duration of the request
	requestDurationLabel = `goutils_rpcserver_request_duration_milliseconds{method="%s"}`
)

func incRequestCount(method string) {
	l := fmt.Sprintf(requestCountLabel, method)
	metrics.GetOrCreateCounter(l).Inc()
}

func incIncorrectRequest() {
	incorrectRequestCounter.Inc()
}

func incRequestErrorCount(method string) {
	l := fmt.Sprintf(errorCountLabel, method)
	metrics.GetOrCreateCounter(l).Inc()
}

func incRequestDuration(method string, duration int64) {
	l := fmt.Sprintf(requestDurationLabel, method)
	metrics.GetOrCreateSummary(l).Update(float64(duration))
}

func incInternalErrors() {
	internalErrorsCounter.Inc()
}
