// Package jsonrpc is a minimal JSON-RPC implementation
package jsonrpc

import (
	"encoding/json"
	"fmt"
)

// As per JSON-RPC 2.0 Specification
// https://www.jsonrpc.org/specification#error_object
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

type JSONRPCResponse struct {
	ID      interface{}     `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
	Version string          `json:"jsonrpc"`
}

// JSONRPCError as per spec: https://www.jsonrpc.org/specification#error_object
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (err JSONRPCError) Error() string {
	return fmt.Sprintf("Error %d (%s)", err.Code, err.Message)
}

func NewJSONRPCResponse(id interface{}, result json.RawMessage) *JSONRPCResponse {
	return &JSONRPCResponse{
		ID:      id,
		Result:  result,
		Version: "2.0",
	}
}

func NewJSONRPCErrorResponse(id interface{}, code int, message string) *JSONRPCResponse {
	return &JSONRPCResponse{
		ID:      id,
		Error:   &JSONRPCError{code, message},
		Version: "2.0",
	}
}
