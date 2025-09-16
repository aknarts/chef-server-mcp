# Makefile for chef-server-mcp
# Provides common developer workflows: build, test, lint, docker, etc.
# Auto-detects a VERSION from git; falls back to 'dev' when not available.

SHELL := /bin/bash
GO    ?= go
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LD_FLAGS := -s -w -X github.com/aknarts/chef-server-mcp/internal/version.Version=$(VERSION)
BINARY_HTTP := chef-mcp
BINARY_MCP  := mcp-chef
BUILD_DIR := dist

# Default target
.PHONY: all
all: build

## ---------- Build Targets ----------
.PHONY: build
build: build-http build-mcp ## Build both binaries

.PHONY: build-http
build-http: ## Build HTTP server binary
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags "$(LD_FLAGS)" -o $(BUILD_DIR)/$(BINARY_HTTP) ./cmd/chef-mcp
	@echo "Built $(BUILD_DIR)/$(BINARY_HTTP) (version $(VERSION))"

.PHONY: build-mcp
build-mcp: ## Build MCP (stdio JSON-RPC) binary
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags "$(LD_FLAGS)" -o $(BUILD_DIR)/$(BINARY_MCP) ./cmd/mcp-chef
	@echo "Built $(BUILD_DIR)/$(BINARY_MCP) (version $(VERSION))"

## ---------- Run / Dev ----------
.PHONY: run
run: build-http ## Run HTTP server (uses built binary)
	./$(BUILD_DIR)/$(BINARY_HTTP)

.PHONY: run-dev
run-dev: ## Run HTTP server via 'go run'
	$(GO) run ./cmd/chef-mcp

.PHONY: run-mcp
run-mcp: build-mcp ## Run MCP server (stdin/stdout JSON-RPC)
	@echo "Starting MCP (press Ctrl+D or send exit method to quit)" >&2
	./$(BUILD_DIR)/$(BINARY_MCP)

## ---------- Testing & Lint ----------
.PHONY: test
test: ## Run unit tests with race detector
	$(GO) test -race -count=1 ./...

.PHONY: cover
cover: ## Generate coverage profile & summary
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out | tail -n 1

.PHONY: vet
vet: ## Run go vet static analysis
	$(GO) vet ./...

.PHONY: tidy
tidy: ## Run go mod tidy
	$(GO) mod tidy

.PHONY: lint
lint: vet ## Placeholder (extend with staticcheck / gosec)

.PHONY: ci
ci: tidy lint test build ## Run typical CI pipeline locally

## ---------- Docker ----------
IMAGE ?= ghcr.io/aknarts/chef-server-mcp
PLATFORMS ?= linux/amd64

.PHONY: docker-build
docker-build: ## Build docker image (single-arch)
	docker build --build-arg VERSION=$(VERSION) -t $(IMAGE):$(VERSION) -t $(IMAGE):latest .

.PHONY: docker-push
docker-push: docker-build ## Push docker image (requires auth)
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest

## ---------- Utilities ----------
.PHONY: mcp-smoke
mcp-smoke: build-mcp ## Simple MCP newline JSON smoke test
	@echo 'Running MCP smoke test...' >&2
	@{ \
	  echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}'; \
	  echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"listNodes","arguments":{}}}'; \
	  echo '{"jsonrpc":"2.0","id":3,"method":"shutdown"}'; \
	  echo '{"jsonrpc":"2.0","method":"exit"}'; \
	} | ./$(BUILD_DIR)/$(BINARY_MCP)

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR) coverage.out

.PHONY: help
help: ## Show this help
	@echo "Makefile targets:"; \
	grep -E '^[a-zA-Z0-9_-]+:.*?##' $(MAKEFILE_LIST) | sed -E 's/:.*?##/\t/' | sort

# End of Makefile

