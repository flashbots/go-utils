package rpcclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flashbots/go-utils/signature"
)

// needed to retrieve requests that arrived at httpServer for further investigation
var requestChan = make(chan *RequestData, 1)

// the request datastructure that can be retrieved for test assertions
type RequestData struct {
	request *http.Request
	body    string
}

// set the response body the httpServer should return for the next request
var responseBody = ""

var (
	httpStatusCode = http.StatusOK
	httpServer     *httptest.Server
)

// start the test-http server and stop it when tests are finished
func TestMain(m *testing.M) {
	httpServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		defer r.Body.Close()
		// put request and body to channel for the client to investigate them
		requestChan <- &RequestData{r, string(data)}

		w.WriteHeader(httpStatusCode)
		fmt.Fprint(w, responseBody)
	}))
	defer httpServer.Close()

	os.Exit(m.Run())
}

func TestSimpleRpcCallHeaderCorrect(t *testing.T) {
	check := assert.New(t)

	responseBody = `{"result": null}`
	rpcClient := NewClient(httpServer.URL)
	_, err := rpcClient.Call(context.Background(), "add", 1, 2)
	check.Nil(err)

	req := (<-requestChan).request

	check.Equal("POST", req.Method)
	check.Equal("application/json", req.Header.Get("Content-Type"))
	check.Equal("application/json", req.Header.Get("Accept"))
}

// test if the structure of a rpc request is built correctly by validating the data that arrived at the test server
func TestRpcClient_Call(t *testing.T) {
	check := assert.New(t)

	rpcClient := NewClient(httpServer.URL)

	person := Person{
		Name:    "Alex",
		Age:     35,
		Country: "Germany",
	}

	drink := Drink{
		Name:        "Cuba Libre",
		Ingredients: []string{"rum", "cola"},
	}

	_, err := rpcClient.Call(context.Background(), "missingParam")
	check.Nil(err)
	check.Equal(`{"method":"missingParam","id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "nullParam", nil)
	check.Nil(err)
	check.Equal(`{"method":"nullParam","params":[null],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "nullParams", nil, nil)
	check.Nil(err)
	check.Equal(`{"method":"nullParams","params":[null,null],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "emptyParams", []interface{}{})
	check.Nil(err)
	check.Equal(`{"method":"emptyParams","params":[[]],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "emptyAnyParams", []string{})
	check.Nil(err)
	check.Equal(`{"method":"emptyAnyParams","params":[[]],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "emptyObject", struct{}{})
	check.Nil(err)
	check.Equal(`{"method":"emptyObject","params":[{}],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "emptyObjectList", []struct{}{{}, {}})
	check.Nil(err)
	check.Equal(`{"method":"emptyObjectList","params":[[{},{}]],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "boolParam", true)
	check.Nil(err)
	check.Equal(`{"method":"boolParam","params":[true],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "boolParams", true, false, true)
	check.Nil(err)
	check.Equal(`{"method":"boolParams","params":[true,false,true],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "stringParam", "Alex")
	check.Nil(err)
	check.Equal(`{"method":"stringParam","params":["Alex"],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "stringParams", "JSON", "RPC")
	check.Nil(err)
	check.Equal(`{"method":"stringParams","params":["JSON","RPC"],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "numberParam", 123)
	check.Nil(err)
	check.Equal(`{"method":"numberParam","params":[123],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "numberParams", 123, 321)
	check.Nil(err)
	check.Equal(`{"method":"numberParams","params":[123,321],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "floatParam", 1.23)
	check.Nil(err)
	check.Equal(`{"method":"floatParam","params":[1.23],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "floatParams", 1.23, 3.21)
	check.Nil(err)
	check.Equal(`{"method":"floatParams","params":[1.23,3.21],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "manyParams", "Alex", 35, true, nil, 2.34)
	check.Nil(err)
	check.Equal(`{"method":"manyParams","params":["Alex",35,true,null,2.34],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "emptyMissingPublicFieldObject", struct{ name string }{name: "Alex"})
	check.Nil(err)
	check.Equal(`{"method":"emptyMissingPublicFieldObject","params":[{}],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "singleStruct", person)
	check.Nil(err)
	check.Equal(`{"method":"singleStruct","params":[{"name":"Alex","age":35,"country":"Germany"}],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "singlePointerToStruct", &person)
	check.Nil(err)
	check.Equal(`{"method":"singlePointerToStruct","params":[{"name":"Alex","age":35,"country":"Germany"}],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	pp := &person
	_, err = rpcClient.Call(context.Background(), "doublePointerStruct", &pp)
	check.Nil(err)
	check.Equal(`{"method":"doublePointerStruct","params":[{"name":"Alex","age":35,"country":"Germany"}],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "multipleStructs", person, &drink)
	check.Nil(err)
	check.Equal(`{"method":"multipleStructs","params":[{"name":"Alex","age":35,"country":"Germany"},{"name":"Cuba Libre","ingredients":["rum","cola"]}],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "singleStructInArray", []interface{}{person})
	check.Nil(err)
	check.Equal(`{"method":"singleStructInArray","params":[[{"name":"Alex","age":35,"country":"Germany"}]],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "namedParameters", map[string]interface{}{
		"name": "Alex",
		"age":  35,
	})
	check.Nil(err)
	check.Equal(`{"method":"namedParameters","params":[{"age":35,"name":"Alex"}],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "anonymousStructNoTags", struct {
		Name string
		Age  int
	}{"Alex", 33})
	check.Nil(err)
	check.Equal(`{"method":"anonymousStructNoTags","params":[{"Name":"Alex","Age":33}],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "anonymousStructWithTags", struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{"Alex", 33})
	check.Nil(err)
	check.Equal(`{"method":"anonymousStructWithTags","params":[{"name":"Alex","age":33}],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "structWithNullField", struct {
		Name    string  `json:"name"`
		Address *string `json:"address"`
	}{"Alex", nil})
	check.Nil(err)
	check.Equal(`{"method":"structWithNullField","params":[{"name":"Alex","address":null}],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	_, err = rpcClient.Call(context.Background(), "nestedStruct",
		Planet{
			Name: "Mars",
			Properties: Properties{
				Distance: 54600000,
				Color:    "red",
			},
		})

	check.Nil(err)
	check.Equal(`{"method":"nestedStruct","params":[{"name":"Mars","properties":{"distance":54600000,"color":"red"}}],"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)

	request := NewRequestWithObjectParam(0, "singleStructRawObjectRequest", person)
	_, err = rpcClient.CallRaw(context.Background(), request)
	check.Nil(err)
	check.Equal(`{"method":"singleStructRawObjectRequest","params":{"name":"Alex","age":35,"country":"Germany"},"id":0,"jsonrpc":"2.0"}`, (<-requestChan).body)
}

func TestRpcClient_CallBatch(t *testing.T) {
	responseBody = `[{"result": null}]`
	check := assert.New(t)

	rpcClient := NewClient(httpServer.URL)

	person := Person{
		Name:    "Alex",
		Age:     35,
		Country: "Germany",
	}

	drink := Drink{
		Name:        "Cuba Libre",
		Ingredients: []string{"rum", "cola"},
	}

	// invalid parameters are possible by manually defining *RPCRequest
	_, err := rpcClient.CallBatch(context.Background(), RPCRequests{
		{
			Method: "singleRequest",
			Params: 3, // invalid, should be []int{3}
		},
	})
	check.Nil(err)
	check.Equal(`[{"method":"singleRequest","params":3,"id":0,"jsonrpc":"2.0"}]`, (<-requestChan).body)

	// better use Params() unless you know what you are doing
	_, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		{
			Method: "singleRequest",
			Params: []int{3}, // always valid json rpc
		},
	})
	check.Nil(err)
	check.Equal(`[{"method":"singleRequest","params":[3],"id":0,"jsonrpc":"2.0"}]`, (<-requestChan).body)

	// even better, use NewRequest()
	_, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("multipleRequests1", 1),
		NewRequest("multipleRequests2", 2),
		NewRequest("multipleRequests3", 3),
	})
	check.Nil(err)
	check.Equal(`[{"method":"multipleRequests1","params":[1],"id":0,"jsonrpc":"2.0"},{"method":"multipleRequests2","params":[2],"id":1,"jsonrpc":"2.0"},{"method":"multipleRequests3","params":[3],"id":2,"jsonrpc":"2.0"}]`, (<-requestChan).body)

	// test a huge batch request
	requests := RPCRequests{
		NewRequest("nullParam", nil),
		NewRequest("nullParams", nil, nil),
		NewRequest("emptyParams", []interface{}{}),
		NewRequest("emptyAnyParams", []string{}),
		NewRequest("emptyObject", struct{}{}),
		NewRequest("emptyObjectList", []struct{}{{}, {}}),
		NewRequest("boolParam", true),
		NewRequest("boolParams", true, false, true),
		NewRequest("stringParam", "Alex"),
		NewRequest("stringParams", "JSON", "RPC"),
		NewRequest("numberParam", 123),
		NewRequest("numberParams", 123, 321),
		NewRequest("floatParam", 1.23),
		NewRequest("floatParams", 1.23, 3.21),
		NewRequest("manyParams", "Alex", 35, true, nil, 2.34),
		NewRequest("emptyMissingPublicFieldObject", struct{ name string }{name: "Alex"}),
		NewRequest("singleStruct", person),
		NewRequest("singlePointerToStruct", &person),
		NewRequest("multipleStructs", person, &drink),
		NewRequest("singleStructInArray", []interface{}{person}),
		NewRequest("namedParameters", map[string]interface{}{
			"name": "Alex",
			"age":  35,
		}),
		NewRequest("anonymousStructNoTags", struct {
			Name string
			Age  int
		}{"Alex", 33}),
		NewRequest("anonymousStructWithTags", struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}{"Alex", 33}),
		NewRequest("structWithNullField", struct {
			Name    string  `json:"name"`
			Address *string `json:"address"`
		}{"Alex", nil}),
	}
	_, err = rpcClient.CallBatch(context.Background(), requests)
	check.Nil(err)

	check.Equal(`[{"method":"nullParam","params":[null],"id":0,"jsonrpc":"2.0"},`+
		`{"method":"nullParams","params":[null,null],"id":1,"jsonrpc":"2.0"},`+
		`{"method":"emptyParams","params":[[]],"id":2,"jsonrpc":"2.0"},`+
		`{"method":"emptyAnyParams","params":[[]],"id":3,"jsonrpc":"2.0"},`+
		`{"method":"emptyObject","params":[{}],"id":4,"jsonrpc":"2.0"},`+
		`{"method":"emptyObjectList","params":[[{},{}]],"id":5,"jsonrpc":"2.0"},`+
		`{"method":"boolParam","params":[true],"id":6,"jsonrpc":"2.0"},`+
		`{"method":"boolParams","params":[true,false,true],"id":7,"jsonrpc":"2.0"},`+
		`{"method":"stringParam","params":["Alex"],"id":8,"jsonrpc":"2.0"},`+
		`{"method":"stringParams","params":["JSON","RPC"],"id":9,"jsonrpc":"2.0"},`+
		`{"method":"numberParam","params":[123],"id":10,"jsonrpc":"2.0"},`+
		`{"method":"numberParams","params":[123,321],"id":11,"jsonrpc":"2.0"},`+
		`{"method":"floatParam","params":[1.23],"id":12,"jsonrpc":"2.0"},`+
		`{"method":"floatParams","params":[1.23,3.21],"id":13,"jsonrpc":"2.0"},`+
		`{"method":"manyParams","params":["Alex",35,true,null,2.34],"id":14,"jsonrpc":"2.0"},`+
		`{"method":"emptyMissingPublicFieldObject","params":[{}],"id":15,"jsonrpc":"2.0"},`+
		`{"method":"singleStruct","params":[{"name":"Alex","age":35,"country":"Germany"}],"id":16,"jsonrpc":"2.0"},`+
		`{"method":"singlePointerToStruct","params":[{"name":"Alex","age":35,"country":"Germany"}],"id":17,"jsonrpc":"2.0"},`+
		`{"method":"multipleStructs","params":[{"name":"Alex","age":35,"country":"Germany"},{"name":"Cuba Libre","ingredients":["rum","cola"]}],"id":18,"jsonrpc":"2.0"},`+
		`{"method":"singleStructInArray","params":[[{"name":"Alex","age":35,"country":"Germany"}]],"id":19,"jsonrpc":"2.0"},`+
		`{"method":"namedParameters","params":[{"age":35,"name":"Alex"}],"id":20,"jsonrpc":"2.0"},`+
		`{"method":"anonymousStructNoTags","params":[{"Name":"Alex","Age":33}],"id":21,"jsonrpc":"2.0"},`+
		`{"method":"anonymousStructWithTags","params":[{"name":"Alex","age":33}],"id":22,"jsonrpc":"2.0"},`+
		`{"method":"structWithNullField","params":[{"name":"Alex","address":null}],"id":23,"jsonrpc":"2.0"}]`, (<-requestChan).body)

	// create batch manually
	requests = []*RPCRequest{
		{
			Method:  "myMethod1",
			Params:  []int{1},
			ID:      123,   // will be forced to requests[i].ID == i unless you use CallBatchRaw
			JSONRPC: "7.0", // will be forced to "2.0"  unless you use CallBatchRaw
		},
		{
			Method:  "myMethod2",
			Params:  &person,
			ID:      321,     // will be forced to requests[i].ID == i unless you use CallBatchRaw
			JSONRPC: "wrong", // will be forced to "2.0" unless you use CallBatchRaw
		},
	}
	_, err = rpcClient.CallBatch(context.Background(), requests)
	check.Nil(err)

	check.Equal(`[{"method":"myMethod1","params":[1],"id":0,"jsonrpc":"2.0"},`+
		`{"method":"myMethod2","params":{"name":"Alex","age":35,"country":"Germany"},"id":1,"jsonrpc":"2.0"}]`, (<-requestChan).body)

	// use raw batch
	requests = []*RPCRequest{
		{
			Method:  "myMethod1",
			Params:  []int{1},
			ID:      123,
			JSONRPC: "7.0",
		},
		{
			Method:  "myMethod2",
			Params:  &person,
			ID:      321,
			JSONRPC: "wrong",
		},
	}
	_, err = rpcClient.CallBatchRaw(context.Background(), requests)
	check.Nil(err)

	check.Equal(`[{"method":"myMethod1","params":[1],"id":123,"jsonrpc":"7.0"},`+
		`{"method":"myMethod2","params":{"name":"Alex","age":35,"country":"Germany"},"id":321,"jsonrpc":"wrong"}]`, (<-requestChan).body)
}

// test if the result of a rpc request is parsed correctly and if errors are thrown correctly
func TestRpcJsonResponseStruct(t *testing.T) {
	check := assert.New(t)

	rpcClient := NewClient(httpServer.URL)

	// empty return body is an error
	responseBody = ``
	res, err := rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.NotNil(err)
	check.Nil(res)

	// not a json body is an error
	responseBody = `{ "not": "a", "json": "object"`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.NotNil(err)
	check.Nil(res)

	// field "anotherField" not allowed in rpc response is an error
	responseBody = `{ "anotherField": "norpc"}`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.NotNil(err)
	check.Nil(res)

	// result null is ok
	responseBody = `{"result": null}`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Result)
	check.Nil(res.Error)

	// error null is ok
	responseBody = `{"error": null}`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Result)
	check.Nil(res.Error)

	// result and error null is ok
	responseBody = `{"result": null, "error": null}`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Result)
	check.Nil(res.Error)

	// result string is ok
	responseBody = `{"result": "ok"}`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Equal("ok", res.Result)

	// result with error null is ok
	responseBody = `{"result": "ok", "error": null}`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Equal("ok", res.Result)

	// error with result null is ok
	responseBody = `{"error": {"code": 123, "message": "something wrong"}, "result": null}`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Result)
	check.Equal(123, res.Error.Code)
	check.Equal("something wrong", res.Error.Message)

	// error with code and message is ok
	responseBody = `{ "error": {"code": 123, "message": "something wrong"}}`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Result)
	check.Equal(123, res.Error.Code)
	check.Equal("something wrong", res.Error.Message)

	// check results

	// should return int correctly
	responseBody = `{ "result": 1 }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	i, err := res.GetInt()
	check.Nil(err)
	check.Equal(int64(1), i)

	// error on not int
	responseBody = `{ "result": "notAnInt" }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	i, err = res.GetInt()
	check.NotNil(err)
	check.Equal(int64(0), i)

	// error on not int but float
	responseBody = `{ "result": 1.234 }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	i, err = res.GetInt()
	check.NotNil(err)
	check.Equal(int64(0), i)

	// error on result null
	responseBody = `{ "result": null }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	i, err = res.GetInt()
	check.NotNil(err)
	check.Equal(int64(0), i)

	responseBody = `{ "result": true }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	b, err := res.GetBool()
	check.Nil(err)
	check.Equal(true, b)

	responseBody = `{ "result": 123 }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	b, err = res.GetBool()
	check.NotNil(err)
	check.Equal(false, b)

	responseBody = `{ "result": "string" }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	str, err := res.GetString()
	check.Nil(err)
	check.Equal("string", str)

	responseBody = `{ "result": 1.234 }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	str, err = res.GetString()
	check.NotNil(err)
	check.Equal("", str)

	responseBody = `{ "result": 1.234 }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	f, err := res.GetFloat()
	check.Nil(err)
	check.Equal(1.234, f)

	responseBody = `{ "result": "notfloat" }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	f, err = res.GetFloat()
	check.NotNil(err)
	check.Equal(0.0, f)

	var p *Person
	responseBody = `{ "result": {"name": "Alex", "age": 35, "anotherField": "something"} }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	err = res.GetObject(&p)
	check.Nil(err)
	check.Equal("Alex", p.Name)
	check.Equal(35, p.Age)
	check.Equal("", p.Country)

	// TODO: How to check if result could be parsed or if it is default?
	p = nil
	responseBody = `{ "result": {"anotherField": "something"} }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	err = res.GetObject(&p)
	check.Nil(err)
	check.NotNil(p)

	var pp *PointerFieldPerson
	responseBody = `{ "result": {"anotherField": "something", "country": "Germany"} }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	err = res.GetObject(&pp)
	check.Nil(err)
	check.Nil(pp.Name)
	check.Nil(pp.Age)
	check.Equal("Germany", *pp.Country)

	p = nil
	responseBody = `{ "result": null }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	err = res.GetObject(&p)
	check.Nil(err)
	check.Nil(p)

	// passing nil is an error
	p = nil
	responseBody = `{ "result": null }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	err = res.GetObject(p)
	check.NotNil(err)
	check.Nil(p)

	p2 := &Person{
		Name: "Alex",
	}
	responseBody = `{ "result": null }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	err = res.GetObject(&p2)
	check.Nil(err)
	check.Nil(p2)

	p2 = &Person{
		Name: "Alex",
	}
	responseBody = `{ "result": {"age": 35} }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	err = res.GetObject(p2)
	check.Nil(err)
	check.Equal("Alex", p2.Name)
	check.Equal(35, p2.Age)

	// prefilled struct is kept on no result
	p3 := Person{
		Name: "Alex",
	}
	responseBody = `{ "result": null }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	err = res.GetObject(&p3)
	check.Nil(err)
	check.Equal("Alex", p3.Name)

	// prefilled struct is extended / overwritten
	p3 = Person{
		Name: "Alex",
		Age:  123,
	}
	responseBody = `{ "result": {"age": 35, "country": "Germany"} }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	err = res.GetObject(&p3)
	check.Nil(err)
	check.Equal("Alex", p3.Name)
	check.Equal(35, p3.Age)
	check.Equal("Germany", p3.Country)

	// nil is an error
	responseBody = `{ "result": {"age": 35} }`
	res, err = rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Nil(res.Error)
	err = res.GetObject(nil)
	check.NotNil(err)
}

func TestRpcClientOptions(t *testing.T) {
	check := assert.New(t)

	t.Run("allowUnknownFields false should return error on unknown field", func(t *testing.T) {
		rpcClient := NewClientWithOpts(httpServer.URL, &RPCClientOpts{AllowUnknownFields: false})

		// unknown field should cause error
		responseBody = `{ "result": 1, "unknown_field": 2 }`
		res, err := rpcClient.Call(context.Background(), "something", 1, 2, 3)
		<-requestChan
		check.NotNil(err)
		check.Nil(res)
	})

	t.Run("allowUnknownFields true should not return error on unknown field", func(t *testing.T) {
		rpcClient := NewClientWithOpts(httpServer.URL, &RPCClientOpts{AllowUnknownFields: true})

		// unknown field should not cause error now
		responseBody = `{ "result": 1, "unknown_field": 2 }`
		res, err := rpcClient.Call(context.Background(), "something", 1, 2, 3)
		<-requestChan
		check.Nil(err)
		check.NotNil(res)
	})

	t.Run("customheaders should be added to request", func(t *testing.T) {
		rpcClient := NewClientWithOpts(httpServer.URL, &RPCClientOpts{
			CustomHeaders: map[string]string{
				"X-Custom-Header":  "custom-value",
				"X-Custom-Header2": "custom-value2",
			},
		})

		responseBody = `{"result": 1}`
		res, err := rpcClient.Call(context.Background(), "something", 1, 2, 3)
		reqObject := <-requestChan
		check.Nil(err)
		check.NotNil(res)
		check.Equal("custom-value", reqObject.request.Header.Get("X-Custom-Header"))
		check.Equal("custom-value2", reqObject.request.Header.Get("X-Custom-Header2"))
	})

	t.Run("host header should be added to request", func(t *testing.T) {
		rpcClient := NewClientWithOpts(httpServer.URL, &RPCClientOpts{
			CustomHeaders: map[string]string{
				"X-Custom-Header1": "custom-value1",
				"Host":             "my-host.com",
				"X-Custom-Header2": "custom-value2",
			},
		})

		responseBody = `{"result": 1}`
		res, err := rpcClient.Call(context.Background(), "something", 1, 2, 3)
		reqObject := <-requestChan
		check.Nil(err)
		check.NotNil(res)
		check.Equal("custom-value1", reqObject.request.Header.Get("X-Custom-Header1"))
		check.Equal("my-host.com", reqObject.request.Host)
		check.Equal("custom-value2", reqObject.request.Header.Get("X-Custom-Header2"))
	})

	t.Run("default rpcrequest id should be customized", func(t *testing.T) {
		rpcClient := NewClientWithOpts(httpServer.URL, &RPCClientOpts{
			DefaultRequestID: 123,
		})

		_, err := rpcClient.Call(context.Background(), "myMethod", 1, 2, 3)
		check.Nil(err)
		check.Equal(`{"method":"myMethod","params":[1,2,3],"id":123,"jsonrpc":"2.0"}`, (<-requestChan).body)
	})
}

func TestRpcBatchJsonResponseStruct(t *testing.T) {
	check := assert.New(t)

	rpcClient := NewClient(httpServer.URL)

	// empty return body is an error
	responseBody = ``
	res, err := rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	check.NotNil(err)
	check.Nil(res)

	// not a json body is an error
	responseBody = `{ "not": "a", "json": "object"`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	check.NotNil(err)
	check.Nil(res)

	// field "anotherField" not allowed in rpc response is an error
	responseBody = `{ "anotherField": "norpc"}`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	check.NotNil(err)
	check.Nil(res)

	// result must be wrapped in array on batch request
	responseBody = `{"result": null}`
	_, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	check.NotNil(err.Error())

	// result ok since in array
	responseBody = `[{"result": null}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	check.Nil(err)
	check.Equal(1, len(res))
	check.Nil(res[0].Result)

	// error null is ok
	responseBody = `[{"error": null}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	check.Nil(err)
	check.Nil(res[0].Result)
	check.Nil(res[0].Error)

	// result and error null is ok
	responseBody = `[{"result": null, "error": null}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	check.Nil(err)
	check.Nil(res[0].Result)
	check.Nil(res[0].Error)

	// result string is ok
	responseBody = `[{"result": "ok","id":0}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	check.Nil(err)
	check.Equal("ok", res[0].Result)
	check.Equal(0, res[0].ID)

	// result with error null is ok
	responseBody = `[{"result": "ok", "error": null}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	check.Nil(err)
	check.Equal("ok", res[0].Result)

	// error with result null is ok
	responseBody = `[{"error": {"code": 123, "message": "something wrong"}, "result": null}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	check.Nil(err)
	check.Nil(res[0].Result)
	check.Equal(123, res[0].Error.Code)
	check.Equal("something wrong", res[0].Error.Message)

	// error with code and message is ok
	responseBody = `[{ "error": {"code": 123, "message": "something wrong"}}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	check.Nil(err)
	check.Nil(res[0].Result)
	check.Equal(123, res[0].Error.Code)
	check.Equal("something wrong", res[0].Error.Message)

	// check results

	// should return int correctly
	responseBody = `[{ "result": 1 }]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	check.Nil(err)
	check.Nil(res[0].Error)
	i, err := res[0].GetInt()
	check.Nil(err)
	check.Equal(int64(1), i)

	// error on wrong type
	responseBody = `[{ "result": "notAnInt" }]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	check.Nil(err)
	check.Nil(res[0].Error)
	i, err = res[0].GetInt()
	check.NotNil(err)
	check.Equal(int64(0), i)

	var p *Person
	responseBody = `[{"id":0, "result": {"name": "Alex", "age": 35}}, {"id":2, "result": {"name": "Lena", "age": 2}}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})

	<-requestChan
	check.Nil(err)

	check.Nil(res[0].Error)
	check.Equal(0, res[0].ID)

	check.Nil(res[1].Error)
	check.Equal(2, res[1].ID)

	err = res[0].GetObject(&p)
	check.Nil(err)
	check.Equal("Alex", p.Name)
	check.Equal(35, p.Age)

	err = res[1].GetObject(&p)
	check.Nil(err)
	check.Equal("Lena", p.Name)
	check.Equal(2, p.Age)

	// check if error occurred
	responseBody = `[{ "result": "someresult", "error": null}, { "result": null, "error": {"code": 123, "message": "something wrong"}}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	check.Nil(err)
	check.True(res.HasError())

	// check if error occurred
	responseBody = `[{ "result": null, "error": {"code": 123, "message": "something wrong"}}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	check.Nil(err)
	check.True(res.HasError())
	// check if error occurred
	responseBody = `[{ "result": null, "error": {"code": 123, "message": "something wrong"}}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	check.Nil(err)
	check.True(res.HasError())

	// check if response mapping works
	responseBody = `[{ "id":123,"result": 123},{ "id":1,"result": 1}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	check.Nil(err)
	check.False(res.HasError())
	resMap := res.AsMap()

	int1, _ := resMap[1].GetInt()
	int123, _ := resMap[123].GetInt()
	check.Equal(int64(1), int1)
	check.Equal(int64(123), int123)

	// check if getByID works
	int123, _ = res.GetByID(123).GetInt()
	check.Equal(int64(123), int123)

	// check if missing id returns nil
	missingIDRes := res.GetByID(124)
	check.Nil(missingIDRes)

	// check if error occurred
	responseBody = `[{ "result": null, "error": {"code": 123, "message": "something wrong"}}]`
	res, err = rpcClient.CallBatch(context.Background(), RPCRequests{
		NewRequest("something", 1, 2, 3),
	})
	<-requestChan
	check.Nil(err)
	check.True(res.HasError())
}

func TestRpcClient_CallFor(t *testing.T) {
	check := assert.New(t)

	rpcClient := NewClient(httpServer.URL)

	i := 0
	responseBody = `{"result":3,"id":0,"jsonrpc":"2.0"}`
	err := rpcClient.CallFor(context.Background(), &i, "something", 1, 2, 3)
	<-requestChan
	check.Nil(err)
	check.Equal(3, i)
}

func TestErrorHandling(t *testing.T) {
	check := assert.New(t)
	rpcClient := NewClient(httpServer.URL)

	oldStatusCode := httpStatusCode
	oldResponseBody := responseBody
	defer func() {
		httpStatusCode = oldStatusCode
		responseBody = oldResponseBody
	}()

	t.Run("check returned rpcerror", func(t *testing.T) {
		responseBody = `{"error":{"code":123,"message":"something wrong"}}`
		call, err := rpcClient.Call(context.Background(), "something", 1, 2, 3)
		<-requestChan
		check.Nil(err)
		check.NotNil(call.Error)
		check.Equal("123: something wrong", call.Error.Error())
	})

	t.Run("check returned httperror", func(t *testing.T) {
		responseBody = `{"error":{"code":123,"message":"something wrong"}}`
		httpStatusCode = http.StatusInternalServerError
		call, err := rpcClient.Call(context.Background(), "something", 1, 2, 3)
		<-requestChan
		check.Nil(err)
		check.NotNil(call)
		check.NotNil(call.Error)
		check.Equal("123: something wrong", call.Error.Error())
	})
}

func TestSignedRequest(t *testing.T) {
	check := assert.New(t)
	signer, _ := signature.NewRandomSigner()
	rpcClient := NewClientWithOpts(httpServer.URL, &RPCClientOpts{
		Signer: signer,
	})

	res, err := rpcClient.Call(context.Background(), "something", 1, 2, 3)
	reqObject := <-requestChan
	check.Nil(err)
	check.NotNil(res)
	header := reqObject.request.Header.Get(signature.HTTPHeader)
	recoveredAddress, err := signature.Verify(header, []byte(reqObject.body))
	check.Nil(err)
	check.Equal(signer.Address(), recoveredAddress)
}

func TestUnsignedRequest(t *testing.T) {
	check := assert.New(t)
	rpcClient := NewClient(httpServer.URL)

	res, err := rpcClient.Call(context.Background(), "something", 1, 2, 3)
	reqObject := <-requestChan
	check.Nil(err)
	check.NotNil(res)
	header := reqObject.request.Header.Get(signature.HTTPHeader)
	check.Equal("", header)
}

func TestCallFlashbots(t *testing.T) {
	check := assert.New(t)
	signer, _ := signature.NewRandomSigner()
	rpcClient := NewClientWithOpts("https://relay.flashbots.net", &RPCClientOpts{
		Signer: signer,
	})

	res, _ := rpcClient.Call(context.Background(), "eth_sendBundle", struct{}{})
	// Disabled the following two lines because they work locally, but reliably fail in Github CI!
	// See also https://github.com/flashbots/go-utils/actions/runs/13905273154/job/38919059341?pr=37
	// check.NotNil(err, res)
	// check.Contains(err.Error(), "rpc response error")
	check.NotNil(res)
	check.NotNil(res.Error)
	check.Equal("missing block param", res.Error.Message)
	check.Equal(-32602, res.Error.Code)
}

func TestBrokenFlashbotsErrorResponse(t *testing.T) {
	oldStatusCode := httpStatusCode
	oldResponseBody := responseBody
	defer func() {
		httpStatusCode = oldStatusCode
		responseBody = oldResponseBody
	}()

	check := assert.New(t)
	rpcClient := NewClient(httpServer.URL)

	responseBody = `{"error":"unknown method: something"}`
	httpStatusCode = 400
	res, err := rpcClient.Call(context.Background(), "something", 1, 2, 3)
	<-requestChan
	check.NotNil(err)
	check.Nil(res)
}

type Person struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Country string `json:"country"`
}

type PointerFieldPerson struct {
	Name    *string `json:"name"`
	Age     *int    `json:"age"`
	Country *string `json:"country"`
}

type Drink struct {
	Name        string   `json:"name"`
	Ingredients []string `json:"ingredients"`
}

type Planet struct {
	Name       string     `json:"name"`
	Properties Properties `json:"properties"`
}

type Properties struct {
	Distance int    `json:"distance"`
	Color    string `json:"color"`
}

// benchmarks

func BenchmarkJSONRPCClientNoSignatures(b *testing.B) {
	benchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		defer r.Body.Close()
		w.WriteHeader(httpStatusCode)
		_, _ = w.Write([]byte(`{"result": null}`))
	}))
	defer benchServer.Close()

	rpcClient := NewClient(benchServer.URL)
	responseBody = `{"result": null}`
	for i := 0; i < b.N; i++ {
		_, err := rpcClient.Call(context.Background(), "something", 1, 2, 3)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkJSONRPCClientWithSignatures(b *testing.B) {
	benchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		defer r.Body.Close()
		w.WriteHeader(httpStatusCode)
		_, _ = w.Write([]byte(`{"result": null}`))
	}))
	defer benchServer.Close()

	signer, _ := signature.NewRandomSigner()
	rpcClient := NewClientWithOpts(benchServer.URL, &RPCClientOpts{
		Signer: signer,
	})

	responseBody = `{"result": null}`
	for i := 0; i < b.N; i++ {
		_, err := rpcClient.Call(context.Background(), "something", 1, 2, 3)
		if err != nil {
			panic(err)
		}
	}
}
