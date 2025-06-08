Example RPC Server usage.

- Implement a simple RPC server
- Handle requests with different processing times and gc-heavy operations
- Use pprof for profiling
- Use [Vegeta](https://github.com/tsenart/vegeta) for load testing

Getting started:

```bash
cd examples/rpcserver

# Run the RPC server
go run main.go

# Example requests
curl 'http://localhost:8080' --header 'Content-Type: application/json' --data '{"jsonrpc":"2.0","method":"slow","params":[],"id":2}'

# Using packaged payloads
curl 'http://localhost:8080' --header 'Content-Type: application/json' --data "@rpc-payload-fast.json"
curl 'http://localhost:8080' --header 'Content-Type: application/json' --data "@rpc-payload-slow.json"

# Load testing with Vegeta
vegeta attack -rate=10000 -duration=60s -targets=targets.txt | tee results.bin | vegeta report

# Grab pprof profiles
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
go tool pprof http://localhost:6060/debug/pprof/heap
```
