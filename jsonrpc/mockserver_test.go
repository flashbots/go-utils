package jsonrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorResponse(t *testing.T) {
	server := NewMockJSONRPCServer()
	server.Handlers["eth_call"] = func(req *JSONRPCRequest) (interface{}, error) {
		return nil, &JSONRPCError{Code: 123, Message: "test"}
	}

	req := NewJSONRPCRequest(1, "eth_call", []interface{}{"0xabc"})
	res, err := SendJSONRPCRequest(*req, server.URL)
	assert.Nil(t, err, err)
	assert.NotNil(t, res.Error)
	assert.Equal(t, 123, res.Error.Code)
	assert.Equal(t, "test", res.Error.Message)
}

func TestMockJSONRPCServer_IncrementRequestCounter(t *testing.T) {
	srv := NewMockJSONRPCServer()
	srv.RequestCounter.Store("EXISTING", 0)

	testCases := []struct {
		name                 string
		method               string
		expectedRequestCount int
	}{
		{
			name:                 "Existing value in map",
			method:               "EXISTING",
			expectedRequestCount: 1,
		},
		{
			name:                 "Non existing value in map",
			method:               "UNKNOWN",
			expectedRequestCount: 1,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			srv.IncrementRequestCounter(tt.method)

			value, ok := srv.RequestCounter.Load(tt.method)

			assert.Equal(t, true, ok)
			assert.Equal(t, value, tt.expectedRequestCount)
		})
	}
}
