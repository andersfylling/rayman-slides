.PHONY: build run server lookup test clean fmt lint sprites-debug sprite-editor

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

# Build all binaries
build:
	go build $(LDFLAGS) -tags gio -o bin/rayman ./cmd/rayman-gui
	go build $(LDFLAGS) -o bin/rayserver ./cmd/rayserver
	go build $(LDFLAGS) -o bin/lookup ./cmd/lookup

# Build and run game client (Gio)
run: build
	./bin/rayman

# Build and run server
server: build
	./bin/rayserver

# Build and run lookup service
lookup: build
	./bin/lookup

# Run tests with race detection
test:
	go test -race -v ./...

# Run tests with coverage
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod tidy
	go mod download

# Generate sprite debug GIF
sprites-debug:
	go run ./cmd/sprite-debug
	@echo "Open sprites.debug.gif in browser to inspect sprite regions"

# Run sprite editor GUI
sprite-editor:
	go run -tags gio ./cmd/sprite-editor
