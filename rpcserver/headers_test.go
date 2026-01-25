package rpcserver

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractHeaders(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		names    []string
		expected map[string]string
	}{
		{
			name: "extract edge headers",
			headers: map[string]string{
				"User-Agent": "MetaMask/16.4",
				"Origin":     "https://app.uniswap.org",
			},
			names: EdgeHeaders,
			expected: map[string]string{
				"User-Agent": "MetaMask/16.4",
				"Origin":     "https://app.uniswap.org",
			},
		},
		{
			name: "partial match",
			headers: map[string]string{
				"User-Agent": "MetaMask/16.4",
			},
			names: EdgeHeaders,
			expected: map[string]string{
				"User-Agent": "MetaMask/16.4",
			},
		},
		{
			name:     "no matching headers",
			headers:  map[string]string{},
			names:    EdgeHeaders,
			expected: nil,
		},
		{
			name: "unrelated headers ignored",
			headers: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "TestClient",
			},
			names: EdgeHeaders,
			expected: map[string]string{
				"User-Agent": "TestClient",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/", nil)
			require.NoError(t, err)

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			result := ExtractHeaders(req, tt.names)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractHeaders_NilRequest_Panics(t *testing.T) {
	require.Panics(t, func() {
		ExtractHeaders(nil, EdgeHeaders)
	})
}

func TestHeaderConstants(t *testing.T) {
	// Verify constants match expected values
	require.Equal(t, "User-Agent", HeaderUserAgent)
	require.Equal(t, "Origin", HeaderOrigin)
	require.Equal(t, "X-Flashbots-Origin", HeaderFlashbotsOrigin)
	require.Equal(t, "X-Flashbots-Signature", HeaderFlashbotsSignature)
}

func TestEdgeHeaders(t *testing.T) {
	// EdgeHeaders should contain standard HTTP headers
	require.Contains(t, EdgeHeaders, HeaderUserAgent)
	require.Contains(t, EdgeHeaders, HeaderOrigin)
}
