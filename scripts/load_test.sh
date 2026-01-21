#!/bin/bash

# Term Idle Load Testing Script
# Tests concurrent SSH connections to simulate multiple players

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SSH_PORT=${SSH_PORT:-2222}
SSH_HOST=${SSH_HOST:-localhost}
CONNECTIONS=${CONNECTIONS:-10}
TEST_DURATION=${TEST_DURATION:-30}
SSH_USER=${SSH_USER:-testuser}
SSH_KEY_FILE=${SSH_KEY_FILE:-"./load_test_key"}

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

print_header() {
    echo -e "${BLUE}[LOAD TEST]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to generate SSH key pair for testing
generate_ssh_keys() {
    print_status "Generating SSH keys for load testing..."
    
    if [ -f "$SSH_KEY_FILE" ]; then
        print_warning "SSH key file already exists, reusing it"
        return
    fi
    
    # Generate SSH key pair
    ssh-keygen -t rsa -b 2048 -f "$SSH_KEY_FILE" -N "" -C "load-test-key"
    chmod 600 "$SSH_KEY_FILE"
    chmod 644 "${SSH_KEY_FILE}.pub"
    
    print_status "SSH keys generated successfully"
}

# Function to start SSH server if not running
ensure_ssh_server() {
    print_status "Checking SSH server availability..."
    
    # Check if SSH server is running on the specified port
    if ! nc -z "$SSH_HOST" "$SSH_PORT" 2>/dev/null; then
        print_warning "SSH server not running on $SSH_HOST:$SSH_PORT"
        print_status "Attempting to start SSH server..."
        
        # Try to start SSH server if possible
        if [ -f "./bin/ssh-server" ]; then
            print_status "Starting SSH server from ./bin/ssh-server"
            nohup ./bin/ssh-server -port="$SSH_PORT" -host-key="./test_ssh_host_key" > ssh_server_load_test.log 2>&1 &
            SSH_SERVER_PID=$!
            echo $SSH_SERVER_PID > ssh_server.pid
            
            # Wait for server to start
            for i in {1..10}; do
                if nc -z "$SSH_HOST" "$SSH_PORT" 2>/dev/null; then
                    print_status "SSH server started successfully (PID: $SSH_SERVER_PID)"
                    break
                fi
                sleep 1
            done
        else
            print_error "SSH server binary not found at ./bin/ssh-server"
            print_error "Please build the SSH server first: make build-ssh"
            exit 1
        fi
    else
        print_status "SSH server is already running on $SSH_HOST:$SSH_PORT"
    fi
    
    # Final check
    if ! nc -z "$SSH_HOST" "$SSH_PORT" 2>/dev/null; then
        print_error "Failed to start SSH server on $SSH_HOST:$SSH_PORT"
        exit 1
    fi
}

# Function to simulate a single SSH connection
simulate_ssh_connection() {
    local connection_id=$1
    local duration=$2
    local temp_output="/tmp/ssh_test_${connection_id}.log"
    
    # Connect and stay connected for the specified duration
    timeout "$duration" ssh -i "$SSH_KEY_FILE" \
        -o StrictHostKeyChecking=no \
        -o UserKnownHostsFile=/dev/null \
        -o ConnectTimeout=10 \
        -o ServerAliveInterval=30 \
        -o ServerAliveCountMax=3 \
        "${SSH_USER}@${SSH_HOST}" \
        "echo 'Connected from connection $connection_id'; sleep $duration" \
        > "$temp_output" 2>&1
    
    local exit_code=$?
    
    if [ $exit_code -eq 124 ]; then
        echo -e "${GREEN}✓${NC} Connection $connection_id: Success (completed ${duration}s)"
    elif [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}✓${NC} Connection $connection_id: Success"
    else
        echo -e "${RED}✗${NC} Connection $connection_id: Failed (exit code: $exit_code)"
        if [ -f "$temp_output" ]; then
            echo "  Error details: $(tail -1 "$temp_output")"
        fi
    fi
    
    # Clean up temp file
    rm -f "$temp_output"
}

# Function to run concurrent SSH connections
run_load_test() {
    local connections=$1
    local duration=$2
    
    print_header "Starting SSH Load Test"
    print_status "Target: $SSH_HOST:$SSH_PORT"
    print_status "Connections: $connections"
    print_status "Duration per connection: ${duration}s"
    echo ""
    
    # Start connections in background
    local start_time=$(date +%s)
    local pids=()
    
    for i in $(seq 1 $connections); do
        simulate_ssh_connection $i $duration &
        pids+=($!)
        
        # Stagger connections slightly to avoid overwhelming the server
        sleep 0.1
    done
    
    print_status "Started $connections SSH connections (PIDs: ${pids[*]})"
    
    # Wait for all connections to complete
    local completed=0
    local failed=0
    
    for pid in "${pids[@]}"; do
        if wait $pid; then
            ((completed++))
        else
            ((failed++))
        fi
    done
    
    local end_time=$(date +%s)
    local total_time=$((end_time - start_time))
    
    echo ""
    print_header "Load Test Results"
    echo "Total connections attempted: $connections"
    echo "Successful connections: $completed"
    echo "Failed connections: $failed"
    echo "Total test time: ${total_time}s"
    
    if [ $failed -eq 0 ]; then
        echo -e "${GREEN}✓ All connections successful!${NC}"
        return 0
    else
        local failure_rate=$((failed * 100 / connections))
        echo -e "${YELLOW}⚠ Failure rate: ${failure_rate}%${NC}"
        if [ $failure_rate -gt 50 ]; then
            echo -e "${RED}✗ High failure rate detected!${NC}"
            return 1
        fi
        return 0
    fi
}

# Function to run incremental load test
run_incremental_test() {
    local max_connections=$1
    local step=${2:-5}
    local duration=${3:-10}
    
    print_header "Running Incremental Load Test"
    
    for i in $(seq $step $step $max_connections); do
        echo ""
        print_status "Testing with $i connections..."
        if run_load_test $i $duration; then
            print_status "✓ $i connections: PASSED"
        else
            print_warning "⚠ $i connections: FAILED (stopping incremental test)"
            break
        fi
        
        # Brief pause between tests
        sleep 3
    done
}

# Function to run stress test
run_stress_test() {
    local max_connections=$1
    local duration=${2:-60}
    
    print_header "Running Stress Test"
    print_warning "This test will push the server to its limits!"
    
    run_load_test $max_connections $duration
}

# Function to cleanup
cleanup() {
    print_status "Cleaning up..."
    
    # Kill SSH server if we started it
    if [ -f "ssh_server.pid" ]; then
        local pid=$(cat ssh_server.pid)
        if kill -0 "$pid" 2>/dev/null; then
            print_status "Stopping SSH server (PID: $pid)"
            kill "$pid"
            rm -f ssh_server.pid
        fi
    fi
    
    # Clean up SSH keys
    rm -f "$SSH_KEY_FILE" "${SSH_KEY_FILE}.pub"
    
    # Clean up temporary files
    rm -f /tmp/ssh_test_*.log
    
    print_status "Cleanup completed"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  quick         Run quick load test (10 connections, 30s)"
    echo "  incremental   Run incremental test (5, 10, 15... connections)"
    echo "  stress        Run stress test (50 connections, 60s)"
    echo "  custom        Run custom test with specified parameters"
    echo "  cleanup       Clean up test files and processes"
    echo "  help          Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  SSH_PORT      SSH server port (default: 2222)"
    echo "  SSH_HOST      SSH server host (default: localhost)"
    echo "  CONNECTIONS   Number of connections (default: 10)"
    echo "  TEST_DURATION Test duration per connection in seconds (default: 30)"
    echo "  SSH_USER      SSH username (default: testuser)"
    echo "  SSH_KEY_FILE  Path to SSH private key (default: ./load_test_key)"
    echo ""
    echo "Examples:"
    echo "  $0 quick                              # Quick test with defaults"
    echo "  $0 custom CONNECTIONS=20 DURATION=60  # Custom test"
    echo "  $0 incremental                        # Incremental test"
    echo "  CONNECTIONS=100 $0 stress             # Stress test with 100 connections"
}

# Main script logic
main() {
    local command=${1:-quick}
    
    # Check prerequisites
    print_status "Checking prerequisites..."
    
    if ! command_exists ssh; then
        print_error "ssh client is not installed"
        exit 1
    fi
    
    if ! command_exists nc; then
        print_warning "netcat is not installed, some features may not work"
    fi
    
    # Set up signal handlers for cleanup
    trap cleanup EXIT INT TERM
    
    case "$command" in
        "quick")
            ensure_ssh_server
            generate_ssh_keys
            run_load_test $CONNECTIONS $TEST_DURATION
            ;;
        "incremental")
            local max_conn=${MAX_CONNECTIONS:-50}
            ensure_ssh_server
            generate_ssh_keys
            run_incremental_test $max_conn
            ;;
        "stress")
            local max_conn=${MAX_CONNECTIONS:-50}
            ensure_ssh_server
            generate_ssh_keys
            run_stress_test $max_conn $TEST_DURATION
            ;;
        "custom")
            ensure_ssh_server
            generate_ssh_keys
            run_load_test $CONNECTIONS $TEST_DURATION
            ;;
        "cleanup")
            cleanup
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

# Parse command line arguments
for arg in "$@"; do
    case $arg in
        CONNECTIONS=*)
            CONNECTIONS="${arg#*=}"
            ;;
        DURATION=*)
            TEST_DURATION="${arg#*=}"
            ;;
        MAX_CONNECTIONS=*)
            MAX_CONNECTIONS="${arg#*=}"
            ;;
        SSH_PORT=*)
            SSH_PORT="${arg#*=}"
            ;;
        SSH_HOST=*)
            SSH_HOST="${arg#*=}"
            ;;
        SSH_USER=*)
            SSH_USER="${arg#*=}"
            ;;
        SSH_KEY_FILE=*)
            SSH_KEY_FILE="${arg#*=}"
            ;;
    esac
done

# Run main function with all arguments
main "$@"