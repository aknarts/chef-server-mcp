# Makefile for chef-server-mcp (MCP-only after HTTP removal)
SHELL := /bin/bash
GO    ?= go
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LD_FLAGS := -s -w -X github.com/aknarts/chef-server-mcp/internal/version.Version=$(VERSION)
BINARY_MCP  := mcp-chef
BUILD_DIR := dist

.PHONY: all
all: build

## ---------- Build Targets ----------
.PHONY: build
build: build-mcp ## Build MCP binary only

.PHONY: build-mcp
build-mcp: ## Build MCP (stdio JSON-RPC) binary
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags "$(LD_FLAGS)" -o $(BUILD_DIR)/$(BINARY_MCP) ./cmd/mcp-chef
	@echo "Built $(BUILD_DIR)/$(BINARY_MCP) (version $(VERSION))"

## ---------- Run / Dev ----------
.PHONY: run-mcp
run-mcp: build-mcp ## Run MCP server (stdin/stdout JSON-RPC)
	@echo "Starting MCP (press Ctrl+D or send exit method to quit)" >&2
	./$(BUILD_DIR)/$(BINARY_MCP)

## ---------- Testing & Lint ----------
.PHONY: test
test: ## Run unit tests (none currently after HTTP removal except package-level)
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
docker-build: ## Build docker image (single-arch) containing only mcp-chef
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
