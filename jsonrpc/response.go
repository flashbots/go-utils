// Package jsonrpc is a minimal JSON-RPC implementation
package jsonrpc

import (
	"encoding/json"
	"fmt"
)

// As per JSON-RPC 2.0 Specification
// https://www.jsonrpc.org/specification#error_object
const (
	ErrParse          int = -32700
	ErrInvalidRequest int = -32600
	ErrMethodNotFound int = -32601
	ErrInvalidParams  int = -32602
	ErrInternal       int = -32603
)

type JSONRPCResponse struct {
	ID      interface{}     `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
	Version string          `json:"jsonrpc"`
}

// JSONRPCError as per spec: https://www.jsonrpc.org/specification#error_object
type JSONRPCError struct {
	// A Number that indicates the error type that occurred.
	Code int `json:"code"`

	// A String providing a short description of the error.
	// The message SHOULD be limited to a concise single sentence.
	Message string `json:"message"`

	// A Primitive or Structured value that contains additional information about the error.
	Data interface{} `json:"data,omitempty"` /* optional */
}

func (err *JSONRPCError) Error() string {
	if err.Message == "" {
		return fmt.Sprintf("json-rpc error %d", err.Code)
	}
	return err.Message
}

func (err *JSONRPCError) ErrorCode() int {
	return err.Code
}

func (err *JSONRPCError) ErrorData() interface{} {
	return err.Data
}

// Error wraps RPC errors, which contain an error code in addition to the message.
type Error interface {
	Error() string  // returns the message
	ErrorCode() int // returns the code
}

type DataError interface {
	Error() string          // returns the message
	ErrorData() interface{} // returns the error data
}

func errorPayload(err error) *JSONRPCError {
	msg := &JSONRPCError{
		Code:    ErrInternal,
		Message: err.Error(),
	}
	ec, ok := err.(Error)
	if ok {
		msg.Code = ec.ErrorCode()
	}
	de, ok := err.(DataError)
	if ok {
		msg.Data = de.ErrorData()
	}
	return msg
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
		ID: id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
		Version: "2.0",
	}
}
