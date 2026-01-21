.PHONY: build run lint test clean build-ssh run-ssh

# Variables
BINARY=term-idle
SSH_BINARY=ssh-server
CMD_DIR=cmd/term-idle
SSH_CMD_DIR=cmd/ssh-server
MAIN_FILE=$(CMD_DIR)/main.go
SSH_MAIN_FILE=$(SSH_CMD_DIR)/main.go

# Default target
all: build build-ssh

# Build the game binary
build:
	@echo "Building $(BINARY)..."
	go build -o $(BINARY) $(MAIN_FILE)
	@echo "Build complete: ./$(BINARY)"

# Build the SSH server binary
build-ssh:
	@echo "Building $(SSH_BINARY)..."
	go build -o $(SSH_BINARY) $(SSH_MAIN_FILE)
	@echo "Build complete: ./$(SSH_BINARY)"

# Run the game application
run:
	@echo "Running $(BINARY)..."
	go run $(MAIN_FILE)

# Run the SSH server
run-ssh:
	@echo "Running $(SSH_BINARY)..."
	go run $(SSH_MAIN_FILE)

# Lint the code
lint:
	@echo "Running linter..."
	golangci-lint run

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY)
	go clean

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Download dependencies
vendor:
	@echo "Vendoring dependencies..."
	go mod vendor

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html