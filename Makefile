.PHONY: build run lint test clean

# Variables
BINARY=term-idle
CMD_DIR=cmd/term-idle
MAIN_FILE=$(CMD_DIR)/main.go

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY)..."
	go build -o $(BINARY) $(MAIN_FILE)
	@echo "Build complete: ./$(BINARY)"

# Run the application
run:
	@echo "Running $(BINARY)..."
	go run $(MAIN_FILE)

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