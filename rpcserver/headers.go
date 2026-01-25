package rpcserver

import "net/http"

// Standard HTTP headers
const (
	HeaderUserAgent = "User-Agent"
	HeaderOrigin    = "Origin"
)

// X-Flashbots-* headers for propagation through the service chain.
const (
	HeaderFlashbotsUserAgent  = "X-Flashbots-User-Agent"
	HeaderFlashbotsHttpOrigin = "X-Flashbots-Http-Origin"
	HeaderFlashbotsOrigin     = "X-Flashbots-Origin"
	HeaderFlashbotsSignature  = "X-Flashbots-Signature"
)

// EdgeHeaders are standard HTTP headers captured at the edge service.
var EdgeHeaders = []string{HeaderUserAgent, HeaderOrigin}

// FlashbotsHeaders are X-Flashbots-* headers that propagate through the service chain.
var FlashbotsHeaders = []string{
	HeaderFlashbotsUserAgent,
	HeaderFlashbotsHttpOrigin,
	HeaderFlashbotsOrigin,
	HeaderFlashbotsSignature,
}

// transformMap defines how standard HTTP headers map to X-Flashbots-* headers.
var transformMap = map[string]string{
	HeaderUserAgent: HeaderFlashbotsUserAgent,
	HeaderOrigin:    HeaderFlashbotsHttpOrigin,
}

// ExtractHeaders returns header values for the specified names.
// Returns nil if no matching headers are found.
func ExtractHeaders(req *http.Request, names []string) map[string]string {
	if req == nil {
		return nil
	}

	result := make(map[string]string)
	for _, name := range names {
		if value := req.Header.Get(name); value != "" {
			result[name] = value
		}
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

// TransformToFlashbots converts standard HTTP headers to X-Flashbots-* format.
// User-Agent → X-Flashbots-User-Agent
// Origin → X-Flashbots-Http-Origin
// Returns nil if input is nil or empty.
func TransformToFlashbots(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return nil
	}

	result := make(map[string]string, len(headers))
	for key, value := range headers {
		if flashbotsKey, ok := transformMap[key]; ok {
			result[flashbotsKey] = value
		} else {
			result[key] = value
		}
	}
	return result
}
