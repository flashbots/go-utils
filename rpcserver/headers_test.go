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
			name: "extract flashbots headers",
			headers: map[string]string{
				"X-Flashbots-User-Agent":  "MetaMask/16.4",
				"X-Flashbots-Http-Origin": "https://app.uniswap.org",
				"X-Flashbots-Origin":      "wallet",
			},
			names: FlashbotsHeaders,
			expected: map[string]string{
				"X-Flashbots-User-Agent":  "MetaMask/16.4",
				"X-Flashbots-Http-Origin": "https://app.uniswap.org",
				"X-Flashbots-Origin":      "wallet",
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

func TestExtractHeaders_NilRequest(t *testing.T) {
	result := ExtractHeaders(nil, EdgeHeaders)
	require.Nil(t, result)
}

func TestTransformToFlashbots(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected map[string]string
	}{
		{
			name: "transform standard headers",
			input: map[string]string{
				"User-Agent": "MetaMask/16.4",
				"Origin":     "https://app.uniswap.org",
			},
			expected: map[string]string{
				"X-Flashbots-User-Agent":  "MetaMask/16.4",
				"X-Flashbots-Http-Origin": "https://app.uniswap.org",
			},
		},
		{
			name: "partial transformation",
			input: map[string]string{
				"User-Agent": "TestClient",
			},
			expected: map[string]string{
				"X-Flashbots-User-Agent": "TestClient",
			},
		},
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty input",
			input:    map[string]string{},
			expected: nil,
		},
		{
			name: "unknown headers passed through",
			input: map[string]string{
				"User-Agent":   "TestClient",
				"Custom-Header": "custom-value",
			},
			expected: map[string]string{
				"X-Flashbots-User-Agent": "TestClient",
				"Custom-Header":          "custom-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TransformToFlashbots(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestHeaderConstants(t *testing.T) {
	// Verify constants match expected values
	require.Equal(t, "User-Agent", HeaderUserAgent)
	require.Equal(t, "Origin", HeaderOrigin)
	require.Equal(t, "X-Flashbots-User-Agent", HeaderFlashbotsUserAgent)
	require.Equal(t, "X-Flashbots-Http-Origin", HeaderFlashbotsHttpOrigin)
	require.Equal(t, "X-Flashbots-Origin", HeaderFlashbotsOrigin)
	require.Equal(t, "X-Flashbots-Signature", HeaderFlashbotsSignature)
}

func TestDefaultLists(t *testing.T) {
	// EdgeHeaders should contain standard HTTP headers
	require.Contains(t, EdgeHeaders, HeaderUserAgent)
	require.Contains(t, EdgeHeaders, HeaderOrigin)

	// FlashbotsHeaders should contain X-Flashbots-* headers
	require.Contains(t, FlashbotsHeaders, HeaderFlashbotsUserAgent)
	require.Contains(t, FlashbotsHeaders, HeaderFlashbotsHttpOrigin)
	require.Contains(t, FlashbotsHeaders, HeaderFlashbotsOrigin)
	require.Contains(t, FlashbotsHeaders, HeaderFlashbotsSignature)
}
