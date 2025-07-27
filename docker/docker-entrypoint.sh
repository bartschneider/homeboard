#!/bin/bash
# Docker entrypoint script for E-Paper Dashboard
# Handles initialization, health checks, and graceful startup

set -e

# Default values
CONFIG_PATH=${CONFIG_PATH:-/app/config.json}
PYTHONPATH=${PYTHONPATH:-/app/widgets}
LOG_LEVEL=${LOG_LEVEL:-info}
WIDGET_TIMEOUT=${WIDGET_TIMEOUT:-30}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_debug() {
    if [ "$DEBUG" = "true" ]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

# Pre-flight checks
preflight_checks() {
    log_info "Running pre-flight checks..."
    
    # Check if configuration file exists
    if [ ! -f "$CONFIG_PATH" ]; then
        log_error "Configuration file not found: $CONFIG_PATH"
        exit 1
    fi
    log_info "Configuration file found: $CONFIG_PATH"
    
    # Check if Python is available
    if ! command -v python3 &> /dev/null; then
        log_error "Python3 not found in PATH"
        exit 1
    fi
    log_info "Python3 available: $(python3 --version)"
    
    # Check Python dependencies
    log_info "Checking Python dependencies..."
    python3 -c "import psutil, pytz" 2>/dev/null || {
        log_error "Required Python dependencies not installed"
        exit 1
    }
    log_info "Python dependencies OK"
    
    # Check widgets directory
    if [ ! -d "/app/widgets" ]; then
        log_warn "Widgets directory not found, creating..."
        mkdir -p /app/widgets
    fi
    
    # Validate configuration
    log_info "Validating configuration..."
    python3 -c "
import json
try:
    with open('$CONFIG_PATH', 'r') as f:
        config = json.load(f)
    print('Configuration is valid JSON')
    print(f'Title: {config.get(\"title\", \"Unknown\")}')
    print(f'Port: {config.get(\"server_port\", \"Unknown\")}')
    print(f'Widgets: {len(config.get(\"widgets\", []))}')
except Exception as e:
    print(f'Configuration validation failed: {e}')
    exit(1)
"
    
    log_info "Pre-flight checks completed successfully"
}

# Widget validation
validate_widgets() {
    log_info "Validating widgets..."
    
    local widget_count=0
    local valid_widgets=0
    
    if [ -d "/app/widgets" ]; then
        for widget in /app/widgets/*.py; do
            if [ -f "$widget" ]; then
                widget_count=$((widget_count + 1))
                log_debug "Checking widget: $(basename "$widget")"
                
                # Basic syntax check
                if python3 -m py_compile "$widget" 2>/dev/null; then
                    valid_widgets=$((valid_widgets + 1))
                    log_debug "Widget $(basename "$widget") syntax OK"
                else
                    log_warn "Widget $(basename "$widget") has syntax errors"
                fi
            fi
        done
    fi
    
    log_info "Widget validation complete: $valid_widgets/$widget_count widgets valid"
}

# Initialize application directories
init_directories() {
    log_info "Initializing application directories..."
    
    # Create necessary directories
    mkdir -p /app/data
    mkdir -p /app/logs
    mkdir -p /app/backups
    mkdir -p /app/widgets
    
    # Set permissions
    chmod 755 /app/data /app/logs /app/backups /app/widgets
    
    log_info "Directories initialized"
}

# Wait for dependencies
wait_for_dependencies() {
    log_info "Waiting for dependencies..."
    
    # If Redis is configured, wait for it
    if [ "$REDIS_ENABLED" = "true" ] && [ -n "$REDIS_HOST" ]; then
        log_info "Waiting for Redis at $REDIS_HOST:${REDIS_PORT:-6379}..."
        timeout 30 bash -c "until nc -z $REDIS_HOST ${REDIS_PORT:-6379}; do sleep 1; done"
        log_info "Redis is ready"
    fi
    
    # If database is configured, wait for it
    if [ "$DATABASE_ENABLED" = "true" ] && [ -n "$DATABASE_HOST" ]; then
        log_info "Waiting for database at $DATABASE_HOST:${DATABASE_PORT:-5432}..."
        timeout 30 bash -c "until nc -z $DATABASE_HOST ${DATABASE_PORT:-5432}; do sleep 1; done"
        log_info "Database is ready"
    fi
}

# Signal handlers for graceful shutdown
cleanup() {
    log_info "Received shutdown signal, cleaning up..."
    
    # Send SIGTERM to homeboard process
    if [ -n "$HOMEBOARD_PID" ]; then
        log_info "Stopping homeboard process (PID: $HOMEBOARD_PID)..."
        kill -TERM "$HOMEBOARD_PID" 2>/dev/null || true
        wait "$HOMEBOARD_PID" 2>/dev/null || true
    fi
    
    log_info "Cleanup completed"
    exit 0
}

# Set up signal handlers
trap cleanup SIGTERM SIGINT SIGQUIT

# Main execution
main() {
    log_info "Starting E-Paper Dashboard container..."
    log_info "Version: ${VERSION:-unknown}"
    log_info "Environment: ${GO_ENV:-production}"
    
    # Run initialization
    preflight_checks
    init_directories
    validate_widgets
    wait_for_dependencies
    
    log_info "Starting homeboard application..."
    
    # Start the application
    if [ "$1" = "homeboard" ] || [ "$1" = "./homeboard" ]; then
        # Run the homeboard binary
        exec "$@" &
        HOMEBOARD_PID=$!
        log_info "Homeboard started with PID: $HOMEBOARD_PID"
        
        # Wait for the process
        wait "$HOMEBOARD_PID"
    else
        # Run custom command
        log_info "Running custom command: $*"
        exec "$@"
    fi
}

# Execute main function with all arguments
main "$@"