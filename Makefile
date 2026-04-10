APP_CLI := blockgo
APP_NODE := blockgo-node
BIN_DIR := bin

VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X 'blockgo/internal/version.Version=$(VERSION)' \
           -X 'blockgo/internal/version.Commit=$(COMMIT)' \
           -X 'blockgo/internal/version.Date=$(DATE)'

.PHONY: build build-cli build-node build-release build-release-cli build-release-node fmt fmt-check vet test tidy clean ci run-cli run-node release-check

build: build-cli build-node

build-release: build-release-cli build-release-node

build-cli:
	mkdir -p $(BIN_DIR)
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_CLI) ./cmd/blockgo

build-node:
	mkdir -p $(BIN_DIR)
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NODE) ./cmd/blockgo-node

build-release-cli:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS) -s -w" -o $(BIN_DIR)/$(APP_CLI) ./cmd/blockgo

build-release-node:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS) -s -w" -o $(BIN_DIR)/$(APP_NODE) ./cmd/blockgo-node

fmt:
	gofmt -w .

fmt-check:
	@files=$$(gofmt -l .); \
	if [ -n "$$files" ]; then \
		echo "The following files are not formatted:"; \
		echo "$$files"; \
		exit 1; \
	fi

vet:
	go vet ./...

test:
	go test ./...

tidy:
	go mod tidy

clean:
	rm -rf $(BIN_DIR)

ci: fmt-check vet test build

release-check: fmt-check vet test build-release

run-cli:
	go run ./cmd/blockgo version

run-node:
	go run ./cmd/blockgo-node -config ./configs/node.example.json
