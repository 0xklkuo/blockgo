APP_CLI := blockgo
APP_NODE := blockgo-node
BIN_DIR := bin
DIST_DIR ?= dist

VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X 'blockgo/internal/version.Version=$(VERSION)' \
           -X 'blockgo/internal/version.Commit=$(COMMIT)' \
           -X 'blockgo/internal/version.Date=$(DATE)'
RELEASE_LDFLAGS := $(LDFLAGS) -s -w

.PHONY: build build-cli build-node build-release build-release-cli build-release-node release-dist release-dist-cli release-dist-node ensure-release-target-env fmt fmt-check vet test tidy clean ci run-cli run-node release-check

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
	CGO_ENABLED=0 go build -trimpath -ldflags "$(RELEASE_LDFLAGS)" -o $(BIN_DIR)/$(APP_CLI) ./cmd/blockgo

build-release-node:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags "$(RELEASE_LDFLAGS)" -o $(BIN_DIR)/$(APP_NODE) ./cmd/blockgo-node

release-dist: release-dist-cli release-dist-node

ensure-release-target-env:
	@test -n "$(GOOS)" || (echo "GOOS is required for release-dist targets" >&2; exit 1)
	@test -n "$(GOARCH)" || (echo "GOARCH is required for release-dist targets" >&2; exit 1)

release-dist-cli: ensure-release-target-env
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -ldflags "$(RELEASE_LDFLAGS)" -o $(DIST_DIR)/$(APP_CLI)-$(GOOS)-$(GOARCH) ./cmd/blockgo

release-dist-node: ensure-release-target-env
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -ldflags "$(RELEASE_LDFLAGS)" -o $(DIST_DIR)/$(APP_NODE)-$(GOOS)-$(GOARCH) ./cmd/blockgo-node

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
	go run ./cmd/blockgo gen-localnet -mode local -nodes 1 -out ./configs/run-node
	go run ./cmd/blockgo-node -config ./configs/run-node/node1.json
