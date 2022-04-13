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

	req := NewJSONRPCRequest(1, "eth_call", "0xabc")
	res, err := SendJSONRPCRequest(*req, server.URL)
	assert.Nil(t, err, err)
	assert.NotNil(t, res.Error)
	assert.Equal(t, 123, res.Error.Code)
	assert.Equal(t, "test", res.Error.Message)
}
