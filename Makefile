GIT_VER := $(shell git describe --tags --always --dirty="-dev")
# ECR_URI := 223847889945.dkr.ecr.us-east-2.amazonaws.com/your-project-name

all: clean build

v:
	@echo "Version: ${GIT_VER}"

test:
	go test ./...

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
