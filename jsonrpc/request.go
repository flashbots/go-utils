// Package jsonrpc is a minimal JSON-RPC implementation
package jsonrpc

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/goccy/go-json"
)

type JSONRPCRequest struct {
	ID      interface{}   `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Version string        `json:"jsonrpc,omitempty"`
}

func NewJSONRPCRequest(id interface{}, method string, args interface{}) *JSONRPCRequest {
	return &JSONRPCRequest{
		ID:      id,
		Method:  method,
		Params:  []interface{}{args},
		Version: "2.0",
	}
}

// SendJSONRPCRequest sends the request to URL and returns the general JsonRpcResponse, or an error (note: not the JSONRPCError)
func SendJSONRPCRequest(req JSONRPCRequest, url string) (res *JSONRPCResponse, err error) {
	buf, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	rawResp, err := http.Post(url, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}

	defer rawResp.Body.Close()

	res = new(JSONRPCResponse)
	if err := json.NewDecoder(rawResp.Body).Decode(res); err != nil {
		return nil, err
	}

	return res, nil
}

// SendNewJSONRPCRequest constructs a request and sends it to the URL
func SendNewJSONRPCRequest(id interface{}, method string, args interface{}, url string) (res *JSONRPCResponse, err error) {
	req := NewJSONRPCRequest(id, method, args)
	return SendJSONRPCRequest(*req, url)
}

// SendJSONRPCRequestAndParseResult sends the request and decodes the response into the reply interface. If the JSON-RPC response
// contains an Error property, the it's returned as this function's error.
func SendJSONRPCRequestAndParseResult(req JSONRPCRequest, url string, reply interface{}) (err error) {
	res, err := SendJSONRPCRequest(req, url)
	if err != nil {
		return err
	}

	if res.Error != nil {
		return res.Error
	}

	if res.Result == nil {
		return errors.New("result is null")
	}

	return json.Unmarshal(res.Result, reply)
}
