PROTO     := api/keeper.proto
VERSION   := $(shell git describe --tags --always 2>/dev/null || echo dev)
DATE      := $(shell date -u +%Y-%m-%d)
COMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS   := -X main.buildVersion=$(VERSION) -X main.buildDate=$(DATE) -X main.buildCommit=$(COMMIT)

.PHONY: proto mocks build build-server build-client test cover

proto:
	protoc --go_out=. --go_opt=module=github.com/iliaonishchenko/gophkeeper \
	       --go-grpc_out=. --go-grpc_opt=module=github.com/iliaonishchenko/gophkeeper $(PROTO)

mocks:
	go generate ./...

build: build-server build-client

build-server:
	go build -ldflags "$(LDFLAGS)" -o bin/gophkeeper-server ./cmd/server

build-client:
	go build -ldflags "$(LDFLAGS)" -o bin/gophkeeper-client ./cmd/client

test:
	go test ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
