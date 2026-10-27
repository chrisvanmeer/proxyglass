BINARY_NAME=proxyglass
BUILD_DIR=bin

VERSION?=$(shell cat VERSION 2>/dev/null || echo dev)
LDFLAGS=-ldflags="-w -s -X github.com/chrisvanmeer/proxyglass/internal/build.Version=$(VERSION)"

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

.PHONY: all build build-native build-linux-amd64 build-linux-arm64 build-all lint test clean run css assets

all: lint test build

build: assets
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/proxyglass

build-native:
	@$(MAKE) build GOOS=$$(go env GOOS) GOARCH=$$(go env GOARCH)

build-linux-amd64: assets
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/proxyglass

build-linux-arm64: assets
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/proxyglass

build-all: build-linux-amd64 build-linux-arm64

lint:
	@golangci-lint run ./... 2>/dev/null || go vet ./...

test:
	go test ./...

clean:
	rm -rf $(BUILD_DIR)

run:
	go run ./cmd/proxyglass --config config.yaml

css:
	npm run build:css

assets:
	cp assets/brand/proxyglass-logo.svg internal/server/static/img/proxyglass-logo.svg
	cp assets/brand/proxyglass-mark.svg internal/server/static/img/proxyglass-mark.svg
