package rpcserver

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flashbots/go-utils/rpcclient"
	"github.com/flashbots/go-utils/signature"
	"github.com/stretchr/testify/require"
)

func testHandler(
	handlerOpts JSONRPCHandlerOpts,
	methodOpts map[string]MethodOpts,
) *JSONRPCHandler {
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

	handler, err := NewJSONRPCHandler(map[string]any{
		"function": handlerMethod,
	}, handlerOpts, methodOpts)
	if err != nil {
		panic(err)
	}
	return handler
}

func TestHandler_ServeHTTP(t *testing.T) {
	handler := testHandler(JSONRPCHandlerOpts{}, nil)

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
			expectedResponse: `{"jsonrpc":"2.0","id":null,"error":{"code":-32700,"message":"unexpected end of JSON input"}}`,
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
			expectedResponse: `{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"json: cannot unmarshal string into Go value of type int"}}`,
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
	handler := testHandler(JSONRPCHandlerOpts{}, nil)
	httpServer := httptest.NewServer(handler)
	defer httpServer.Close()

	client := rpcclient.NewClient(httpServer.URL)

	var resp dummyStruct
	err := client.CallFor(context.Background(), &resp, "function", 123)
	require.NoError(t, err)
	require.Equal(t, 123, resp.Field)
}

func TestJSONRPCServerWithSignatureWithClient(t *testing.T) {
	methodName := "function"
	methodConfig := MethodOpts{
		VerifyRequestSignatureFromHeader: true,
	}
	handler := testHandler(JSONRPCHandlerOpts{}, map[string]MethodOpts{
		methodName: methodConfig,
	})
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
