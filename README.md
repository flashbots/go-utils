# go-utils

[![Test status](https://github.com/flashbots/go-utils/workflows/Checks/badge.svg)](https://github.com/flashbots/go-utils/actions?query=workflow%3A%22Checks%22)

Various reusable Go utilities and modules


### `cli`

Various minor command-line interface helpers: [`cli.go`](https://github.com/flashbots/go-utils/blob/main/cli/cli.go)

### httplogger

Logging middleware for HTTP requests using [`go-ethereum/log`](https://github.com/ethereum/go-ethereum/tree/master/log).
See [`examples/httplogger/main.go`](https://github.com/flashbots/go-utils/blob/main/examples/httplogger/main.go)

Install:

```bash
go get github.com/flashbots/go-utils/httplogger
```

Use:

```go
mux := http.NewServeMux()
mux.HandleFunc("/v1/hello", HelloHandler)
loggedRouter := httplogger.LoggingMiddleware(r)
```
