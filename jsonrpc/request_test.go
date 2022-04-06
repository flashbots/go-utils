package jsonrpc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupMockServer() string {
	server := NewMockJSONRPCServer()
	server.handlers["eth_call"] = func(req *JSONRPCRequest) (interface{}, error) {
		return "0x12345", nil
	}
	return server.URL
}

func TestSendJsonRpcRequest(t *testing.T) {
	addr := setupMockServer()

	req, err := NewJSONRPCRequest(1, "eth_call", "0xabc")
	assert.Nil(t, err, err)
	res, err := SendJSONRPCRequest(*req, addr)
	assert.Nil(t, err, err)

	reply := new(string)
	err = json.Unmarshal(res.Result, reply)
	assert.Nil(t, err, err)
	assert.Equal(t, "0x12345", *reply)

	// Test an unknown RPC method
	req2, err := NewJSONRPCRequest(2, "unknown", "foo")
	assert.Nil(t, err, err)
	res2, err := SendJSONRPCRequest(*req2, addr)
	assert.Nil(t, err, err)
	assert.NotNil(t, res2.Error)
}

func TestSendJSONRPCRequestAndParseResult(t *testing.T) {
	addr := setupMockServer()

	req, err := NewJSONRPCRequest(1, "eth_call", "0xabc")
	assert.Nil(t, err, err)
	res := new(string)
	err = SendJSONRPCRequestAndParseResult(*req, addr, res)
	assert.Nil(t, err, err)
	assert.Equal(t, "0x12345", *res)

	req2, err := NewJSONRPCRequest(2, "unknown", "foo")
	assert.Nil(t, err, err)
	res2 := new(string)
	err = SendJSONRPCRequestAndParseResult(*req2, addr, res2)
	assert.NotNil(t, err, err)
}
