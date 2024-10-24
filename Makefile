GIT_VER := $(shell git describe --tags --always --dirty="-dev")

v:
	@echo "Version: ${GIT_VER}"

test:
	go test ./...

fmt: 
	gofmt -s -w .
	gci write .
	gofumpt -w -extra .
	go mod tidy

lint:
	gofmt -d ./
	go vet ./...
	staticcheck ./...

cover:
	go test -coverprofile=/tmp/go-sim-lb.cover.tmp ./...
	go tool cover -func /tmp/go-sim-lb.cover.tmp
	unlink /tmp/go-sim-lb.cover.tmp

cover-html:
	go test -coverprofile=/tmp/go-sim-lb.cover.tmp ./...
	go tool cover -html=/tmp/go-sim-lb.cover.tmp
	unlink /tmp/go-sim-lb.cover.tmp
