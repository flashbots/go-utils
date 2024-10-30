package rpcserver

import (
	"fmt"

	"github.com/VictoriaMetrics/metrics"
)

const (
	// we use unknown method label for methods that server does not support because otherwise
	// users can create arbitrary number of metrics
	unknownMethodLabel = "unknown"

	// incremented when user made incorrect request
	incorrectRequestCounter = `goutils_rpcserver_incorrect_request_total{server_name="%s"}`

	// incremented when server has a bug (e.g. can't marshall response)
	internalErrorsCounter = `goutils_rpcserver_internal_errors_total{server_name="%s"}`

	// incremented when request comes in
	requestCountLabel = `goutils_rpcserver_request_count{method="%s",server_name="%s"}`
	// incremented when handler method returns JSONRPC error
	errorCountLabel = `goutils_rpcserver_error_count{method="%s",server_name="%s"}`
	// total duration of the request
	requestDurationLabel = `goutils_rpcserver_request_duration_milliseconds{method="%s",server_name="%s"}`
)

func incRequestCount(method, serverName string) {
	l := fmt.Sprintf(requestCountLabel, method, serverName)
	metrics.GetOrCreateCounter(l).Inc()
}

func incIncorrectRequest(serverName string) {
	l := fmt.Sprintf(incorrectRequestCounter, serverName)
	metrics.GetOrCreateCounter(l).Inc()
}

func incRequestErrorCount(method, serverName string) {
	l := fmt.Sprintf(errorCountLabel, method, serverName)
	metrics.GetOrCreateCounter(l).Inc()
}

func incRequestDuration(method string, duration int64, serverName string) {
	l := fmt.Sprintf(requestDurationLabel, method, serverName)
	metrics.GetOrCreateSummary(l).Update(float64(duration))
}

func incInternalErrors(serverName string) {
	l := fmt.Sprintf(internalErrorsCounter, serverName)
	metrics.GetOrCreateCounter(l).Inc()
}
