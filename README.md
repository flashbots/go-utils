# goutils

[![Test status](https://github.com/flashbots/goutils/workflows/Checks/badge.svg)](https://github.com/flashbots/goutils/actions?query=workflow%3A%22Checks%22)

Various reusable Go utilities and modules


### httplogger

Logging middleware for HTTP requests using `go-ethereum/log`.

See `examples/httplogger/main.go`

```go
mux := http.NewServeMux()
mux.HandleFunc("/v1/hello", HelloHandler)
loggedRouter := httplogger.LoggingMiddleware(r)
```
