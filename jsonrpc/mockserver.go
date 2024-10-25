package jsonrpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"

	"github.com/ethereum/go-ethereum/log"
)

type MockJSONRPCServer struct {
	Handlers       map[string]func(req *JSONRPCRequest) (interface{}, error)
	RequestCounter sync.Map
	server         *httptest.Server
	URL            string
}

func NewMockJSONRPCServer() *MockJSONRPCServer {
	s := &MockJSONRPCServer{
		Handlers: make(map[string]func(req *JSONRPCRequest) (interface{}, error)),
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

	returnError := func(id interface{}, err error) {
		res := JSONRPCResponse{
			ID:    id,
			Error: errorPayload(err),
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			log.Error("error writing response", "err", err, "data", res)
		}
	}

	// Parse JSON RPC
	jsonReq := new(JSONRPCRequest)
	if err := json.NewDecoder(req.Body).Decode(jsonReq); err != nil {
		returnError(0, fmt.Errorf("failed to parse request body: %v", err))
		return
	}

	jsonRPCHandler, found := s.Handlers[jsonReq.Method]
	if !found {
		returnError(jsonReq.ID, fmt.Errorf("no RPC method handler implemented for %s", jsonReq.Method))
		return
	}

	s.IncrementRequestCounter(jsonReq.Method)

	rawRes, err := jsonRPCHandler(jsonReq)
	if err != nil {
		returnError(jsonReq.ID, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	resBytes, err := json.Marshal(rawRes)
	if err != nil {
		log.Error("error marshalling rawRes", "err", err, "data", rawRes)
		return
	}

	res := NewJSONRPCResponse(jsonReq.ID, resBytes)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Error("error writing response 2", "err", err, "data", rawRes)
		return
	}
}

func (s *MockJSONRPCServer) IncrementRequestCounter(method string) {
	newCount := 0
	currentCount, ok := s.RequestCounter.Load(method)
	if ok {
		newCount = currentCount.(int)
	}
	s.RequestCounter.Store(method, newCount+1)
}

func (s *MockJSONRPCServer) GetRequestCount(method string) int {
	currentCount, ok := s.RequestCounter.Load(method)
	if ok {
		return currentCount.(int)
	}
	return 0
}
