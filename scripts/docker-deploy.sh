#!/bin/bash

# Docker Deployment Script for Term Idle
# Builds and deploys Term Idle using Docker and Docker Compose

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
IMAGE_NAME=${IMAGE_NAME:-term-idle}
TAG=${TAG:-latest}
COMPOSE_FILE=${COMPOSE_FILE:-docker-compose.yml}
ENV_FILE=${ENV_FILE:-.env}

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
    echo -e "${BLUE}[DOCKER]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    if ! command_exists docker; then
        print_error "Docker is not installed or not in PATH"
        echo "Please install Docker: https://docs.docker.com/get-docker/"
        exit 1
    fi
    
    if ! command_exists docker-compose; then
        print_error "Docker Compose is not installed or not in PATH"
        echo "Please install Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi
    
    # Check if Docker daemon is running
    if ! docker info >/dev/null 2>&1; then
        print_error "Docker daemon is not running"
        echo "Please start Docker daemon"
        exit 1
    fi
    
    print_status "Prerequisites check passed"
}

# Function to create environment file if it doesn't exist
create_env_file() {
    if [ ! -f "$ENV_FILE" ]; then
        print_status "Creating environment file: $ENV_FILE"
        cat > "$ENV_FILE" << EOF
# Term Idle Environment Configuration
# Copy this file and modify values as needed

# SSH Configuration
TERMIDLE_SSH_PORT=2222
TERMIDLE_SSH_MAX_SESSIONS=100

# Game Configuration  
TERMIDLE_GAME_SAVE_INTERVAL=30s
TERMIDLE_GAME_PRODUCTION_TICK=1s
TERMIDLE_GAME_MAX_PLAYERS=1000
TERMIDLE_GAME_OFFLINE_PRODUCTION=true

# Database Configuration
TERMIDLE_DATABASE_PATH=/app/data/term_idle.db
TERMIDLE_DATABASE_MAX_CONNECTIONS=10

# HTTP API Configuration
TERMIDLE_SERVER_PORT=8080
TERMIDLE_SERVER_HOST=0.0.0.0
TERMIDLE_SERVER_READ_TIMEOUT=30s
TERMIDLE_SERVER_WRITE_TIMEOUT=30s

# Logging Configuration
TERMIDLE_LOGGING_LEVEL=info
TERMIDLE_LOGGING_FORMAT=text

# Security (for future use)
# POSTGRES_PASSWORD=securepassword
# REDIS_PASSWORD=redispassword
EOF
        print_warning "Environment file created. Review and modify as needed."
    fi
}

# Function to build Docker image
build_image() {
    print_header "Building Docker Image"
    print_status "Building image: $IMAGE_NAME:$TAG"
    
    docker build -t "$IMAGE_NAME:$TAG" .
    
    # Also tag as latest if tag is different
    if [ "$TAG" != "latest" ]; then
        docker tag "$IMAGE_NAME:$TAG" "$IMAGE_NAME:latest"
    fi
    
    print_status "Docker image built successfully"
}

# Function to start services
start_services() {
    print_header "Starting Services"
    
    # Ensure environment file exists
    create_env_file
    
    # Start services using docker-compose
    if [ -f "$COMPOSE_FILE" ]; then
        print_status "Starting services with docker-compose..."
        docker-compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d
        
        print_status "Waiting for services to be healthy..."
        sleep 10
        
        # Check service status
        print_status "Service status:"
        docker-compose -f "$COMPOSE_FILE" ps
    else
        print_error "Docker Compose file not found: $COMPOSE_FILE"
        exit 1
    fi
}

# Function to stop services
stop_services() {
    print_header "Stopping Services"
    
    if [ -f "$COMPOSE_FILE" ]; then
        docker-compose -f "$COMPOSE_FILE" down
        print_status "Services stopped successfully"
    else
        print_warning "No docker-compose file found, stopping containers manually..."
        docker stop $(docker ps -q --filter "ancestor=$IMAGE_NAME") 2>/dev/null || true
    fi
}

# Function to restart services
restart_services() {
    print_header "Restarting Services"
    stop_services
    sleep 2
    start_services
}

# Function to show logs
show_logs() {
    print_header "Showing Logs"
    
    if [ -f "$COMPOSE_FILE" ]; then
        docker-compose -f "$COMPOSE_FILE" logs -f --tail=100
    else
        docker logs -f $(docker ps -q --filter "ancestor=$IMAGE_NAME")
    fi
}

# Function to show status
show_status() {
    print_header "Service Status"
    
    if [ -f "$COMPOSE_FILE" ]; then
        docker-compose -f "$COMPOSE_FILE" ps
    else
        docker ps --filter "ancestor=$IMAGE_NAME"
    fi
    
    echo ""
    print_status "Container resource usage:"
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}" \
        $(docker ps -q --filter "ancestor=$IMAGE_NAME") 2>/dev/null || \
        print_warning "No running containers found"
}

# Function to clean up
cleanup() {
    print_header "Cleaning Up"
    
    # Remove stopped containers
    docker container prune -f
    
    # Remove unused images
    docker image prune -f
    
    # Remove unused volumes (be careful with this!)
    # docker volume prune -f
    
    print_status "Cleanup completed"
}

# Function to update services
update_services() {
    print_header "Updating Services"
    
    # Pull latest code
    if [ -d ".git" ]; then
        print_status "Pulling latest code..."
        git pull
    fi
    
    # Rebuild image
    build_image
    
    # Restart services
    restart_services
    
    print_status "Services updated successfully"
}

# Function to backup data
backup_data() {
    print_header "Backing Up Data"
    
    local backup_dir="backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    # Backup Docker volumes
    if docker volume inspect term-idle_term_idle_data >/dev/null 2>&1; then
        print_status "Backing up database volume..."
        docker run --rm -v term-idle_term_idle_data:/data -v "$(pwd)/$backup_dir":/backup \
            alpine tar czf /backup/data.tar.gz -C /data .
    fi
    
    if docker volume inspect term-idle_term_idle_logs >/dev/null 2>&1; then
        print_status "Backing up logs volume..."
        docker run --rm -v term-idle_term_idle_logs:/logs -v "$(pwd)/$backup_dir":/backup \
            alpine tar czf /backup/logs.tar.gz -C /logs .
    fi
    
    print_status "Backup completed: $backup_dir"
}

# Function to restore data
restore_data() {
    local backup_dir=$1
    
    if [ -z "$backup_dir" ] || [ ! -d "$backup_dir" ]; then
        print_error "Please specify a valid backup directory"
        echo "Usage: $0 restore /path/to/backup"
        exit 1
    fi
    
    print_header "Restoring Data from $backup_dir"
    
    # Stop services first
    stop_services
    
    # Restore data
    if [ -f "$backup_dir/data.tar.gz" ]; then
        print_status "Restoring database volume..."
        docker run --rm -v term-idle_term_idle_data:/data -v "$(pwd)/$backup_dir":/backup \
            alpine tar xzf /backup/data.tar.gz -C /data
    fi
    
    if [ -f "$backup_dir/logs.tar.gz" ]; then
        print_status "Restoring logs volume..."
        docker run --rm -v term-idle_term_idle_logs:/logs -v "$(pwd)/$backup_dir":/backup \
            alpine tar xzf /backup/logs.tar.gz -C /logs
    fi
    
    # Start services
    start_services
    
    print_status "Data restore completed"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  build         Build Docker image"
    echo "  start         Start services"
    echo "  stop          Stop services"
    echo "  restart       Restart services"
    echo "  logs          Show service logs"
    echo "  status        Show service status"
    echo "  update        Pull latest code and restart services"
    echo "  backup        Backup data volumes"
    echo "  restore DIR   Restore data from backup directory"
    echo "  cleanup       Clean up unused Docker resources"
    echo "  help          Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  IMAGE_NAME    Docker image name (default: term-idle)"
    echo "  TAG           Docker image tag (default: latest)"
    echo "  COMPOSE_FILE  Docker Compose file (default: docker-compose.yml)"
    echo "  ENV_FILE      Environment file (default: .env)"
    echo ""
    echo "Examples:"
    echo "  $0 build                              # Build image"
    echo "  $0 start                              # Start services"
    echo "  TAG=v1.0.0 $0 build                   # Build with specific tag"
    echo "  $0 backup                             # Backup data"
    echo "  $0 restore backups/20231201_120000    # Restore from backup"
}

# Main script logic
main() {
    local command=${1:-help}
    
    # Check prerequisites for most commands
    case "$command" in
        build|start|stop|restart|logs|status|update|backup|restore|cleanup)
            check_prerequisites
            ;;
    esac
    
    case "$command" in
        "build")
            build_image
            ;;
        "start")
            start_services
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            restart_services
            ;;
        "logs")
            show_logs
            ;;
        "status")
            show_status
            ;;
        "update")
            update_services
            ;;
        "backup")
            backup_data
            ;;
        "restore")
            restore_data "$2"
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

# Set up signal handlers
trap cleanup EXIT INT TERM

# Parse command line arguments
for arg in "$@"; do
    case $arg in
        IMAGE_NAME=*)
            IMAGE_NAME="${arg#*=}"
            ;;
        TAG=*)
            TAG="${arg#*=}"
            ;;
        COMPOSE_FILE=*)
            COMPOSE_FILE="${arg#*=}"
            ;;
        ENV_FILE=*)
            ENV_FILE="${arg#*=}"
            ;;
    esac
done

# Run main function with all arguments
main "$@"