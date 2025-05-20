package rpcserver

import (
	"fmt"
	"time"

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
	requestCountLabel = `goutils_rpcserver_request_count{method="%s",server_name="%s",is_big="%t"}`
	// incremented when handler method returns JSONRPC error
	errorCountLabel = `goutils_rpcserver_error_count{method="%s",server_name="%s"}`
	// total duration of the request
	requestDurationLabel = `goutils_rpcserver_request_duration_milliseconds{method="%s",server_name="%s",is_big="%t"}`
	// partial duration of the request
	requestDurationStepLabel = `goutils_rpcserver_request_step_duration_milliseconds{method="%s",server_name="%s",step="%s",is_big="%t"}`

	// request size in bytes
	requestSizeBytes = `goutils_rpcserver_request_size_bytes{method="%s",server_name="%s"}`
)

func incRequestCount(method, serverName string, isBig bool) {
	l := fmt.Sprintf(requestCountLabel, method, serverName, isBig)
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

func incRequestDuration(duration time.Duration, method string, serverName string, isBig bool) {
	millis := float64(duration.Microseconds()) / 1000.0
	l := fmt.Sprintf(requestDurationLabel, method, serverName, isBig)
	metrics.GetOrCreateSummary(l).Update(millis)
}

func incInternalErrors(serverName string) {
	l := fmt.Sprintf(internalErrorsCounter, serverName)
	metrics.GetOrCreateCounter(l).Inc()
}

func incRequestDurationStep(duration time.Duration, method, serverName, step string, isBig bool) {
	millis := float64(duration.Microseconds()) / 1000.0
	l := fmt.Sprintf(requestDurationStepLabel, method, serverName, step, isBig)
	metrics.GetOrCreateSummary(l).Update(millis)
}

func incRequestSizeBytes(size int, method string, serverName string) {
	l := fmt.Sprintf(requestSizeBytes, method, serverName)
	metrics.GetOrCreateSummary(l).Update(float64(size))
}
