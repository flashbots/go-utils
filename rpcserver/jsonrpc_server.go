// Package rpcserver allows exposing functions like:
// func Foo(context, int) (int, error)
// as a JSON RPC methods
//
// This implementation is similar to the one in go-ethereum, but the idea is to eventually replace it as a default
// JSON RPC server implementation in Flasbhots projects and for this we need to reimplement some of the quirks of existing API.
package rpcserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/flashbots/go-utils/signature"
)

var (
	// this are the only errors that are returned as http errors with http error codes
	errMethodNotAllowed = "only POST method is allowed"
	errWrongContentType = "header Content-Type must be application/json"
	errMarshalResponse  = "failed to marshal response"

	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
	CodeCustomError    = -32000

	DefaultMaxRequestBodySizeBytes = 30 * 1024 * 1024 // 30mb
)

const (
	maxOriginIDLength    = 255
	requestSizeThreshold = 50_000
)

type (
	highPriorityKey struct{}
	signerKey       struct{}
	originKey       struct{}
	sizeKey         struct{}
)

type jsonRPCRequest struct {
	JSONRPC string            `json:"jsonrpc"`
	ID      any               `json:"id"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
}

type jsonRPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      any              `json:"id"`
	Result  *json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError    `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *any   `json:"data,omitempty"`
}

type JSONRPCHandler struct {
	JSONRPCHandlerOpts
	methods map[string]methodHandler
}

type Methods map[string]any

type JSONRPCHandlerOpts struct {
	// Logger, can be nil
	Log *slog.Logger
	// Server name. Used to separate logs and metrics when having multiple servers in one binary.
	ServerName string
	// Max size of the request payload
	MaxRequestBodySizeBytes int64
	// If true payload signature from X-Flashbots-Signature will be verified
	// Result can be extracted from the context using GetSigner
	VerifyRequestSignatureFromHeader bool
	// If true signer from X-Flashbots-Signature will be extracted without verifying signature
	// Result can be extracted from the context using GetSigner
	ExtractUnverifiedRequestSignatureFromHeader bool
	// If true high_prio header value will be extracted (true or false)
	// Result can be extracted from the context using GetHighPriority
	ExtractPriorityFromHeader bool
	// If true extract value from x-flashbots-origin header
	// Result can be extracted from the context using GetOrigin
	ExtractOriginFromHeader bool
	// GET response content
	GetResponseContent []byte
}

// NewJSONRPCHandler creates JSONRPC http.Handler from the map that maps method names to method functions
// each method function must:
// - have context as a first argument
// - return error as a last argument
// - have argument types that can be unmarshalled from JSON
// - have return types that can be marshalled to JSON
func NewJSONRPCHandler(methods Methods, opts JSONRPCHandlerOpts) (*JSONRPCHandler, error) {
	if opts.MaxRequestBodySizeBytes == 0 {
		opts.MaxRequestBodySizeBytes = int64(DefaultMaxRequestBodySizeBytes)
	}

	m := make(map[string]methodHandler)
	for name, fn := range methods {
		method, err := getMethodTypes(fn)
		if err != nil {
			return nil, err
		}
		m[name] = method
	}
	return &JSONRPCHandler{
		JSONRPCHandlerOpts: opts,
		methods:            m,
	}, nil
}

func (h *JSONRPCHandler) writeJSONRPCResponse(w http.ResponseWriter, response jsonRPCResponse) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		if h.Log != nil {
			h.Log.Error("failed to marshall response", slog.Any("error", err), slog.String("serverName", h.ServerName))
		}
		http.Error(w, errMarshalResponse, http.StatusInternalServerError)
		incInternalErrors(h.ServerName)
		return
	}
}

func (h *JSONRPCHandler) writeJSONRPCError(w http.ResponseWriter, id any, code int, msg string) {
	res := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  nil,
		Error: &jsonRPCError{
			Code:    code,
			Message: msg,
			Data:    nil,
		},
	}
	h.writeJSONRPCResponse(w, res)
}

func (h *JSONRPCHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startAt := time.Now()
	methodForMetrics := unknownMethodLabel
	bigRequest := false
	ctx := r.Context()

	defer func() {
		incRequestCount(methodForMetrics, h.ServerName, bigRequest)
		incRequestDuration(time.Since(startAt), methodForMetrics, h.ServerName, bigRequest)
	}()

	stepStartAt := time.Now()

	if r.Method != http.MethodPost {
		// Respond with GET response content if it's set
		if r.Method == http.MethodGet && len(h.GetResponseContent) > 0 {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write(h.GetResponseContent)
			if err != nil {
				http.Error(w, errMarshalResponse, http.StatusInternalServerError)
				incInternalErrors(h.ServerName)
				return
			}
			return
		}

		// Responsd with "only POST method is allowed"
		http.Error(w, errMethodNotAllowed, http.StatusMethodNotAllowed)
		incIncorrectRequest(h.ServerName)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, errWrongContentType, http.StatusUnsupportedMediaType)
		incIncorrectRequest(h.ServerName)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.MaxRequestBodySizeBytes)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		msg := fmt.Sprintf("request body is too big, max size: %d", h.MaxRequestBodySizeBytes)
		h.writeJSONRPCError(w, nil, CodeInvalidRequest, msg)
		incIncorrectRequest(h.ServerName)
		return
	}
	bodySize := len(body)
	ctx = context.WithValue(ctx, sizeKey{}, bodySize)

	bigRequest = bodySize > requestSizeThreshold
	defer func(size int) {
		incRequestSizeBytes(size, methodForMetrics, h.ServerName)
	}(bodySize)

	stepTime := time.Since(stepStartAt)
	defer func(stepTime time.Duration) {
		incRequestDurationStep(stepTime, methodForMetrics, h.ServerName, "io", bigRequest)
	}(stepTime)
	stepStartAt = time.Now()

	if h.VerifyRequestSignatureFromHeader {
		signatureHeader := r.Header.Get("x-flashbots-signature")
		signer, err := signature.Verify(signatureHeader, body)
		if err != nil {
			h.writeJSONRPCError(w, nil, CodeInvalidRequest, err.Error())
			incIncorrectRequest(h.ServerName)
			return
		}
		ctx = context.WithValue(ctx, signerKey{}, signer)
	}

	// read request
	var req jsonRPCRequest
	if err := json.Unmarshal(body, &req); err != nil {
		h.writeJSONRPCError(w, nil, CodeParseError, err.Error())
		incIncorrectRequest(h.ServerName)
		return
	}

	if req.JSONRPC != "2.0" {
		h.writeJSONRPCError(w, req.ID, CodeParseError, "invalid jsonrpc version")
		incIncorrectRequest(h.ServerName)
		return
	}
	if req.ID != nil {
		// id must be string or number
		switch req.ID.(type) {
		case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		default:
			h.writeJSONRPCError(w, req.ID, CodeParseError, "invalid id type")
			incIncorrectRequest(h.ServerName)
			return
		}
	}

	if h.ExtractPriorityFromHeader {
		highPriority := r.Header.Get("high_prio") == "true"
		ctx = context.WithValue(ctx, highPriorityKey{}, highPriority)
	}

	if h.ExtractUnverifiedRequestSignatureFromHeader {
		signature := r.Header.Get("x-flashbots-signature")
		if split := strings.Split(signature, ":"); len(split) > 0 {
			signer := common.HexToAddress(split[0])
			ctx = context.WithValue(ctx, signerKey{}, signer)
		}
	}

	if h.ExtractOriginFromHeader {
		origin := r.Header.Get("x-flashbots-origin")
		if origin != "" {
			if len(origin) > maxOriginIDLength {
				h.writeJSONRPCError(w, req.ID, CodeInvalidRequest, "x-flashbots-origin header is too long")
				incIncorrectRequest(h.ServerName)
				return
			}
			ctx = context.WithValue(ctx, originKey{}, origin)
		}
	}

	// get method
	method, ok := h.methods[req.Method]
	if !ok {
		h.writeJSONRPCError(w, req.ID, CodeMethodNotFound, "method not found")
		incIncorrectRequest(h.ServerName)
		return
	}
	methodForMetrics = req.Method

	incRequestDurationStep(time.Since(stepStartAt), methodForMetrics, h.ServerName, "parse", bigRequest)
	stepStartAt = time.Now()

	// call method
	result, err := method.call(ctx, req.Params)
	if err != nil {
		h.writeJSONRPCError(w, req.ID, CodeCustomError, err.Error())
		incRequestErrorCount(methodForMetrics, h.ServerName)
		incRequestDurationStep(time.Since(stepStartAt), methodForMetrics, h.ServerName, "call", bigRequest)
		return
	}

	incRequestDurationStep(time.Since(stepStartAt), methodForMetrics, h.ServerName, "call", bigRequest)
	stepStartAt = time.Now()

	marshaledResult, err := json.Marshal(result)
	if err != nil {
		h.writeJSONRPCError(w, req.ID, CodeInternalError, err.Error())
		incInternalErrors(h.ServerName)

		incRequestDurationStep(time.Since(stepStartAt), methodForMetrics, h.ServerName, "response", bigRequest)
		return
	}

	// write response
	rawMessageResult := json.RawMessage(marshaledResult)
	res := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  &rawMessageResult,
		Error:   nil,
	}
	h.writeJSONRPCResponse(w, res)

	incRequestDurationStep(time.Since(stepStartAt), methodForMetrics, h.ServerName, "response", bigRequest)
}

func GetHighPriority(ctx context.Context) bool {
	value, ok := ctx.Value(highPriorityKey{}).(bool)
	if !ok {
		return false
	}
	return value
}

func GetSigner(ctx context.Context) common.Address {
	value, ok := ctx.Value(signerKey{}).(common.Address)
	if !ok {
		return common.Address{}
	}
	return value
}

func GetOrigin(ctx context.Context) string {
	value, ok := ctx.Value(originKey{}).(string)
	if !ok {
		return ""
	}
	return value
}

func GetRequestSize(ctx context.Context) int {
	return ctx.Value(sizeKey{}).(int)
}
