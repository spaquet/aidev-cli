.PHONY: build run clean test lint help

# Build variables
BINARY_NAME=aidev
BUILD_DIR=bin
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

help:
	@echo "AIDev CLI - Makefile targets"
	@echo ""
	@echo "  make build          Build the binary"
	@echo "  make run            Build and run the TUI"
	@echo "  make run-login      Run the login command"
	@echo "  make clean          Remove build artifacts"
	@echo "  make lint           Run linters (requires golangci-lint)"
	@echo "  make test           Run tests"
	@echo "  make install        Install to GOBIN"

build:
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/aidev

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

run-login: build
	./$(BUILD_DIR)/$(BINARY_NAME) login --email alice@example.com --password test123

clean:
	rm -rf $(BUILD_DIR)
	go clean -testcache

lint:
	golangci-lint run ./...

test:
	go test -v -cover ./...

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOBIN)/$(BINARY_NAME)

# Development
dev-watch:
	@command -v entr >/dev/null || (echo "entr not installed. Install with: brew install entr"; exit 1)
	find internal cmd -name '*.go' | entr make build

cross-build:
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/aidev
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/aidev
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/aidev
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/aidev
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/aidev
	ls -lh $(BUILD_DIR)

fmt:
	go fmt ./...

vet:
	go vet ./...
