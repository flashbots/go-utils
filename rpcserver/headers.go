package rpcserver

import "net/http"

// Standard HTTP headers
const (
	HeaderUserAgent = "User-Agent"
	HeaderOrigin    = "Origin"
)

// X-Flashbots-* headers for internal service communication.
const (
	HeaderFlashbotsOrigin    = "X-Flashbots-Origin"
	HeaderFlashbotsSignature = "X-Flashbots-Signature"
)

// EdgeHeaders are standard HTTP headers captured at the edge and propagated through the service chain.
var EdgeHeaders = []string{HeaderUserAgent, HeaderOrigin}

// ExtractHeaders returns header values for the specified names.
// Returns nil if no matching headers are found.
// Panics if req is nil - callers must ensure valid request context.
func ExtractHeaders(req *http.Request, names []string) map[string]string {
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
