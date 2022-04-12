package jsonrpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/ethereum/go-ethereum/log"
)

type MockJSONRPCServer struct {
	Handlers       map[string]func(req *JSONRPCRequest) (interface{}, error)
	RequestCounter map[string]int
	server         *httptest.Server
	URL            string
}

func NewMockJSONRPCServer() *MockJSONRPCServer {
	s := &MockJSONRPCServer{
		Handlers:       make(map[string]func(req *JSONRPCRequest) (interface{}, error)),
		RequestCounter: make(map[string]int),
	}
	s.server = httptest.NewServer(http.HandlerFunc(s.handleHTTPRequest))
	s.URL = s.server.URL
	return s
}

func (s *MockJSONRPCServer) SetHandler(method string, handler func(req *JSONRPCRequest) (interface{}, error)) {
	s.Handlers[method] = handler
}

func (s *MockJSONRPCServer) handleHTTPRequest(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	testHeader := req.Header.Get("Test")
	w.Header().Set("Test", testHeader)

	returnError := func(id interface{}, msg string) {
		res := JSONRPCResponse{
			ID: id,
			Error: &JSONRPCError{
				Code:    -32603,
				Message: msg,
			},
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			log.Error("error writing response", "err", err, "data", res)
		}
	}

	// Parse JSON RPC
	jsonReq := new(JSONRPCRequest)
	if err := json.NewDecoder(req.Body).Decode(jsonReq); err != nil {
		returnError(0, fmt.Sprintf("failed to parse request body: %v", err))
		return
	}

	jsonRPCHandler, found := s.Handlers[jsonReq.Method]
	if !found {
		returnError(jsonReq.ID, fmt.Sprintf("no RPC method handler implemented for %s", jsonReq.Method))
		return
	}

	s.RequestCounter[jsonReq.Method]++
	rawRes, err := jsonRPCHandler(jsonReq)
	if err != nil {
		returnError(jsonReq.ID, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	resBytes, err := json.Marshal(rawRes)
	if err != nil {
		log.Error("error mashalling rawRes", "err", err, "data", rawRes)
		return
	}

	res := NewJSONRPCResponse(jsonReq.ID, resBytes)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Error("error writing response 2", "err", err, "data", rawRes)
		return
	}
}
