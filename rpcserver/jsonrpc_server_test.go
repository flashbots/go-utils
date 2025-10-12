package rpcserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flashbots/go-utils/rpcclient"
	"github.com/flashbots/go-utils/signature"
	"github.com/stretchr/testify/require"
)

func testHandler(opts JSONRPCHandlerOpts) *JSONRPCHandler {
	var (
		errorArg = -1
		errorOut = errors.New("custom error") //nolint:goerr113
	)
	handlerMethod := func(ctx context.Context, arg1 int) (dummyStruct, error) {
		if arg1 == errorArg {
			return dummyStruct{}, errorOut
		}
		return dummyStruct{arg1}, nil
	}

	handler, err := NewJSONRPCHandler(map[string]interface{}{
		"function": handlerMethod,
	}, opts)
	if err != nil {
		panic(err)
	}
	return handler
}

func TestHandler_ServeHTTP(t *testing.T) {
	handler := testHandler(JSONRPCHandlerOpts{})

	testCases := map[string]struct {
		requestBody      string
		expectedResponse string
	}{
		"success": {
			requestBody:      `{"jsonrpc":"2.0","id":1,"method":"function","params":[1]}`,
			expectedResponse: `{"jsonrpc":"2.0","id":1,"result":{"field":1}}`,
		},
		"error": {
			requestBody:      `{"jsonrpc":"2.0","id":1,"method":"function","params":[-1]}`,
			expectedResponse: `{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"custom error"}}`,
		},
		"invalid json": {
			requestBody:      `{"jsonrpc":"2.0","id":1,"method":"function","params":[1]`,
			expectedResponse: `{"jsonrpc":"2.0","id":null,"error":{"code":-32700,"message":"expected comma after object element"}}`,
		},
		"method not found": {
			requestBody:      `{"jsonrpc":"2.0","id":1,"method":"not_found","params":[1]}`,
			expectedResponse: `{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"method not found"}}`,
		},
		"invalid params": {
			requestBody:      `{"jsonrpc":"2.0","id":1,"method":"function","params":[1,2]}`,
			expectedResponse: `{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"too much arguments"}}`, // TODO: return correct code here
		},
		"invalid params type": {
			requestBody:      `{"jsonrpc":"2.0","id":1,"method":"function","params":["1"]}`,
			expectedResponse: `{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"json: cannot unmarshal number \" into Go value of type int"}}`,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			body := bytes.NewReader([]byte(testCase.requestBody))
			request, err := http.NewRequest(http.MethodPost, "/", body)
			require.NoError(t, err)
			request.Header.Add("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, request)
			require.Equal(t, http.StatusOK, rr.Code)

			require.JSONEq(t, testCase.expectedResponse, rr.Body.String())
		})
	}
}

func TestJSONRPCServerWithClient(t *testing.T) {
	handler := testHandler(JSONRPCHandlerOpts{})
	httpServer := httptest.NewServer(handler)
	defer httpServer.Close()

	client := rpcclient.NewClient(httpServer.URL)

	var resp dummyStruct
	err := client.CallFor(context.Background(), &resp, "function", 123)
	require.NoError(t, err)
	require.Equal(t, 123, resp.Field)
}

func TestJSONRPCServerWithSignatureWithClient(t *testing.T) {
	handler := testHandler(JSONRPCHandlerOpts{VerifyRequestSignatureFromHeader: true})
	httpServer := httptest.NewServer(handler)
	defer httpServer.Close()

	// first we do request without signature
	client := rpcclient.NewClient(httpServer.URL)
	resp, err := client.Call(context.Background(), "function", 123)
	require.NoError(t, err)
	require.Equal(t, "no signature provided", resp.Error.Message)

	// call with signature
	signer, err := signature.NewRandomSigner()
	require.NoError(t, err)
	client = rpcclient.NewClientWithOpts(httpServer.URL, &rpcclient.RPCClientOpts{
		Signer: signer,
	})

	var structResp dummyStruct
	err = client.CallFor(context.Background(), &structResp, "function", 123)
	require.NoError(t, err)
	require.Equal(t, 123, structResp.Field)
}

func TestJSONRPCServerDefaultLiveAndReady(t *testing.T) {
	handler := testHandler(JSONRPCHandlerOpts{})
	httpServer := httptest.NewServer(handler)
	defer httpServer.Close()

	// /livez (200 by default)
	request, err := http.NewRequest(http.MethodGet, "/livez", nil)
	require.NoError(t, err)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, request)
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "", rr.Body.String())

	// /readyz (404 by default)
	request, err = http.NewRequest(http.MethodGet, "/readyz", nil)
	require.NoError(t, err)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, request)
	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestJSONRPCErrorDataIsPreserved(t *testing.T) {
	handlerMethod := func(ctx context.Context, arg int) (int, error) {
		errorData := any("some error data")
		return 0, &JSONRPCError{
			Code:    1234,
			Message: "test error",
			Data:    &errorData,
		}
	}

	handler, err := NewJSONRPCHandler(map[string]interface{}{
		"testError": handlerMethod,
	}, JSONRPCHandlerOpts{})
	require.NoError(t, err)

	httpServer := httptest.NewServer(handler)
	defer httpServer.Close()

	client := rpcclient.NewClient(httpServer.URL)
	resp, err := client.Call(context.Background(), "testError", 1)
	require.NoError(t, err)
	require.NotNil(t, resp.Error)
	require.Equal(t, 1234, resp.Error.Code)
	require.Equal(t, "test error", resp.Error.Message)
	require.Equal(t, "some error data", resp.Error.Data)
}

func TestJSONRPCServerReadyzOK(t *testing.T) {
	handler := testHandler(JSONRPCHandlerOpts{
		ReadyHandler: func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("ready"))
			return err
		},
	})
	httpServer := httptest.NewServer(handler)
	defer httpServer.Close()

	request, err := http.NewRequest(http.MethodGet, "/readyz", nil)
	require.NoError(t, err)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, request)
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "ready", rr.Body.String())
}

func TestJSONRPCServerReadyzError(t *testing.T) {
	handler := testHandler(JSONRPCHandlerOpts{
		ReadyHandler: func(w http.ResponseWriter, r *http.Request) error {
			return fmt.Errorf("not ready")
		},
	})
	httpServer := httptest.NewServer(handler)
	defer httpServer.Close()

	request, err := http.NewRequest(http.MethodGet, "/readyz", nil)
	require.NoError(t, err)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, request)
	require.Equal(t, http.StatusInternalServerError, rr.Code)
	fmt.Println(rr.Body.String())
	require.Equal(t, "not ready\n", rr.Body.String())
}

func TestURLExtraction(t *testing.T) {
	// Handler that captures URL from context
	var capturedURL string
	handlerMethod := func(ctx context.Context) (string, error) {
		url := GetURL(ctx)
		capturedURL = url.Path + "?" + url.RawQuery
		return capturedURL, nil
	}

	handler, err := NewJSONRPCHandler(map[string]interface{}{
		"test": handlerMethod,
	}, JSONRPCHandlerOpts{})
	require.NoError(t, err)

	t.Run("No headers: uses r.URL (backward compat)", func(t *testing.T) {
		body := bytes.NewReader([]byte(`{"jsonrpc":"2.0","id":1,"method":"test","params":[]}`))
		request, err := http.NewRequest(http.MethodPost, "/fast?hint=calldata", body)
		require.NoError(t, err)
		request.Header.Add("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, request)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "/fast?hint=calldata", capturedURL)
	})

	t.Run("Both headers: reconstructs URL (works with any path)", func(t *testing.T) {
		// Test with /whatever instead of /fast to prove it's not hardcoded
		body := bytes.NewReader([]byte(`{"jsonrpc":"2.0","id":1,"method":"test","params":[]}`))
		request, err := http.NewRequest(http.MethodPost, "/", body)
		require.NoError(t, err)
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("X-Original-Path", "/whatever")
		request.Header.Add("X-Original-Query", "hint=hash&builder=flashbots")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, request)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "/whatever?hint=hash&builder=flashbots", capturedURL)
	})

	t.Run("Only query header: uses r.URL.Path", func(t *testing.T) {
		// Proxyd doesn't send X-Original-Path when path is "/"
		body := bytes.NewReader([]byte(`{"jsonrpc":"2.0","id":1,"method":"test","params":[]}`))
		request, err := http.NewRequest(http.MethodPost, "/", body)
		require.NoError(t, err)
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("X-Original-Query", "hint=hash")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, request)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "/?hint=hash", capturedURL)
	})

	t.Run("Only path header: uses r.URL.RawQuery", func(t *testing.T) {
		// Proxyd doesn't send X-Original-Query when there's no query string
		body := bytes.NewReader([]byte(`{"jsonrpc":"2.0","id":1,"method":"test","params":[]}`))
		request, err := http.NewRequest(http.MethodPost, "/api", body)
		require.NoError(t, err)
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("X-Original-Path", "/fast")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, request)

		require.Equal(t, http.StatusOK, rr.Code)
		require.Equal(t, "/fast?", capturedURL)
	})
}
