package jsonrpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/ethereum/go-ethereum/log"
)

func setupMockServer() (addr string) {
	rpcBackendServer := httptest.NewServer(http.HandlerFunc(relayBackendHandler))
	return rpcBackendServer.URL
}

func handleRPCRequest(req *JSONRPCRequest) (result interface{}, err error) {
	switch req.Method {
	case "eth_call":
		return "0x12345", nil
	default:
		return "", fmt.Errorf("no RPC method handler implemented for %s", req.Method)
	}
}

func relayBackendHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	log.Info("mock-backend request", "addr", req.RemoteAddr, "method", req.Method, "url", req.URL)

	w.Header().Set("Content-Type", "application/json")
	testHeader := req.Header.Get("Test")
	w.Header().Set("Test", testHeader)

	returnError := func(id interface{}, msg string) {
		log.Info("returnError", "msg", msg)
		res := JSONRPCResponse{
			ID: id,
			Error: &JSONRPCError{
				Code:    -32603,
				Message: msg,
			},
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			log.Error("error writing response 1", "err", err, "data", res)
		}
	}

	// Parse JSON RPC
	jsonReq := new(JSONRPCRequest)
	if err := json.NewDecoder(req.Body).Decode(jsonReq); err != nil {
		returnError(-1, fmt.Sprintf("failed to parse request body: %v", err))
		return
	}

	rawRes, err := handleRPCRequest(jsonReq)
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
