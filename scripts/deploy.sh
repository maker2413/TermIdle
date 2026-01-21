#!/bin/bash

# Term Idle Deployment Script
# Builds and deploys the term-idle game and SSH server

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BINARY_DIR="./bin"
DATA_DIR="./data"
LOGS_DIR="./logs"
CONFIG_DIR="./configs"
SSH_KEY_FILE="./ssh_host_key"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to create directory if it doesn't exist
ensure_directory() {
    if [ ! -d "$1" ]; then
        print_status "Creating directory: $1"
        mkdir -p "$1"
    fi
}

# Function to generate SSH host key if it doesn't exist
generate_ssh_key() {
    if [ ! -f "$SSH_KEY_FILE" ]; then
        print_status "Generating SSH host key..."
        ssh-keygen -t rsa -b 2048 -f "$SSH_KEY_FILE" -N "" -C "term-idle-host-key"
        chmod 600 "$SSH_KEY_FILE"
        chmod 644 "${SSH_KEY_FILE}.pub"
    else
        print_status "SSH host key already exists"
    fi
}

# Function to stop running services
stop_services() {
    print_status "Stopping any running services..."
    
    # Kill term-idle processes
    if pgrep -f "term-idle" > /dev/null; then
        pkill -f "term-idle" || true
        print_status "Stopped term-idle processes"
    fi
    
    # Kill ssh-server processes
    if pgrep -f "ssh-server" > /dev/null; then
        pkill -f "ssh-server" || true
        print_status "Stopped ssh-server processes"
    fi
    
    # Wait a moment for processes to stop
    sleep 2
}

# Function to build binaries
build_binaries() {
    print_status "Building Term Idle binaries..."
    
    # Ensure binary directory exists
    ensure_directory "$BINARY_DIR"
    
    # Build term-idle binary
    print_status "Building term-idle..."
    if ! go build -o "${BINARY_DIR}/term-idle" cmd/term-idle/main.go; then
        print_error "Failed to build term-idle binary"
        exit 1
    fi
    
    # Build ssh-server binary
    print_status "Building ssh-server..."
    if ! go build -o "${BINARY_DIR}/ssh-server" cmd/ssh-server/main.go; then
        print_error "Failed to build ssh-server binary"
        exit 1
    fi
    
    # Make binaries executable
    chmod +x "${BINARY_DIR}/term-idle"
    chmod +x "${BINARY_DIR}/ssh-server"
    
    print_status "Build completed successfully"
    ls -la "${BINARY_DIR}/"
}

# Function to initialize database
initialize_database() {
    print_status "Initializing database..."
    
    # Ensure data directory exists
    ensure_directory "$DATA_DIR"
    
    # Run database migration if term-idle binary supports it
    if [ -f "${BINARY_DIR}/term-idle" ]; then
        # Check if migrate command exists (we'll need to implement this)
        print_status "Running database setup..."
        # For now, just ensure the database file can be created
        if [ ! -f "./term_idle.db" ]; then
            print_status "Creating new database..."
            touch "./term_idle.db"
        fi
    fi
}

# Function to start services
start_services() {
    print_status "Starting Term Idle services..."
    
    # Start HTTP API server in background
    print_status "Starting HTTP API server on port 8080..."
    nohup "${BINARY_DIR}/term-idle" > "${LOGS_DIR}/term-idle.log" 2>&1 &
    HTTP_PID=$!
    echo $HTTP_PID > "${DATA_DIR}/term-idle.pid"
    
    # Start SSH server in background
    print_status "Starting SSH server on port 2222..."
    nohup "${BINARY_DIR}/ssh-server" > "${LOGS_DIR}/ssh-server.log" 2>&1 &
    SSH_PID=$!
    echo $SSH_PID > "${DATA_DIR}/ssh-server.pid"
    
    print_status "Services started successfully"
    print_status "HTTP API: http://localhost:8080"
    print_status "SSH: ssh username@localhost -p 2222"
    print_status "HTTP PID: $HTTP_PID"
    print_status "SSH PID: $SSH_PID"
}

# Function to validate deployment
validate_deployment() {
    print_status "Validating deployment..."
    
    # Wait a moment for services to start
    sleep 3
    
    # Check if HTTP server is responding
    if curl -s http://localhost:8080/api/health > /dev/null 2>&1; then
        print_status "✓ HTTP API server is responding"
    else
        print_warning "⚠ HTTP API server is not responding yet (may still be starting)"
    fi
    
    # Check if SSH server is listening
    if nc -z localhost 2222 2>/dev/null; then
        print_status "✓ SSH server is listening on port 2222"
    else
        print_warning "⚠ SSH server is not listening on port 2222 (may still be starting)"
    fi
    
    # Check if processes are running
    if pgrep -f "term-idle" > /dev/null; then
        print_status "✓ term-idle process is running"
    else
        print_error "✗ term-idle process is not running"
        return 1
    fi
    
    if pgrep -f "ssh-server" > /dev/null; then
        print_status "✓ ssh-server process is running"
    else
        print_error "✗ ssh-server process is not running"
        return 1
    fi
    
    print_status "Deployment validation completed"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  deploy    Full deployment (stop, build, initialize, start, validate)"
    echo "  stop      Stop running services"
    echo "  build     Build binaries only"
    echo "  start     Start services only"
    echo "  validate  Validate running services"
    echo "  logs      Show service logs"
    echo "  status    Show service status"
    echo "  help      Show this help message"
}

# Function to show logs
show_logs() {
    print_status "Showing recent logs..."
    
    if [ -f "${LOGS_DIR}/term-idle.log" ]; then
        echo "=== Term Idle Logs ==="
        tail -20 "${LOGS_DIR}/term-idle.log"
    fi
    
    if [ -f "${LOGS_DIR}/ssh-server.log" ]; then
        echo ""
        echo "=== SSH Server Logs ==="
        tail -20 "${LOGS_DIR}/ssh-server.log"
    fi
}

# Function to show status
show_status() {
    print_status "Service status:"
    
    # Check processes
    if pgrep -f "term-idle" > /dev/null; then
        echo "✓ term-idle: RUNNING (PID: $(pgrep -f term-idle))"
    else
        echo "✗ term-idle: STOPPED"
    fi
    
    if pgrep -f "ssh-server" > /dev/null; then
        echo "✓ ssh-server: RUNNING (PID: $(pgrep -f ssh-server))"
    else
        echo "✗ ssh-server: STOPPED"
    fi
    
    # Check ports
    if nc -z localhost 8080 2>/dev/null; then
        echo "✓ HTTP port 8080: OPEN"
    else
        echo "✗ HTTP port 8080: CLOSED"
    fi
    
    if nc -z localhost 2222 2>/dev/null; then
        echo "✓ SSH port 2222: OPEN"
    else
        echo "✗ SSH port 2222: CLOSED"
    fi
}

# Main script logic
main() {
    local command=${1:-deploy}
    
    # Ensure required directories exist
    ensure_directory "$BINARY_DIR"
    ensure_directory "$DATA_DIR"
    ensure_directory "$LOGS_DIR"
    ensure_directory "$CONFIG_DIR"
    
    case "$command" in
        "deploy")
            print_status "Starting full deployment..."
            stop_services
            generate_ssh_key
            build_binaries
            initialize_database
            start_services
            validate_deployment
            ;;
        "stop")
            stop_services
            ;;
        "build")
            build_binaries
            ;;
        "start")
            start_services
            ;;
        "validate")
            validate_deployment
            ;;
        "logs")
            show_logs
            ;;
        "status")
            show_status
            ;;
        "help"|"-h"|"--help")
            show_usage
            ;;
        *)
            print_error "Unknown command: $command"
            show_usage
            exit 1
            ;;
    esac
}

# Check prerequisites
print_status "Checking prerequisites..."

if ! command_exists go; then
    print_error "Go is not installed. Please install Go 1.21+"
    exit 1
fi

if ! command_exists nc; then
    print_warning "netcat is not installed. Some validation features may not work."
fi

if ! command_exists curl; then
    print_warning "curl is not installed. Some validation features may not work."
fi

# Run main function with all arguments
main "$@"