#!/bin/bash
# Homeboard KUAL Extension - Main Dashboard Launcher
# This script connects to the homeboard server and displays the assigned dashboard

EXTENSION_DIR="/mnt/us/extensions/homeboard"
CONFIG_FILE="$EXTENSION_DIR/config/device.conf"
LOG_FILE="/tmp/homeboard_kual.log"

# Source configuration
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
else
    echo "Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Logging function
log() {
    local level="$1"
    local message="$2"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    if [ "$DEBUG_MODE" = "true" ] || [ "$level" != "DEBUG" ]; then
        echo "[$timestamp] [$level] $message" | tee -a "$LOG_FILE"
    fi
}

# Check if device is configured
check_device_config() {
    if [ -z "$SERVER_URL" ] || [ -z "$DEVICE_ID" ]; then
        log "WARN" "Device not configured. Server URL: '$SERVER_URL', Device ID: '$DEVICE_ID'"
        return 1
    fi
    return 0
}

# Test network connectivity
test_connectivity() {
    log "INFO" "Testing connectivity to $SERVER_URL:$SERVER_PORT"
    
    # Try to ping the server
    if ! ping -c 1 -W 5 "$SERVER_URL" > /dev/null 2>&1; then
        log "ERROR" "Cannot reach server $SERVER_URL"
        return 1
    fi
    
    # Try to connect to the port
    if ! nc -z -w5 "$SERVER_URL" "$SERVER_PORT" 2>/dev/null; then
        log "ERROR" "Cannot connect to $SERVER_URL:$SERVER_PORT"
        return 1
    fi
    
    log "INFO" "Connectivity test successful"
    return 0
}

# Get dashboard assignment from server
get_dashboard_assignment() {
    local url="http://$SERVER_URL:$SERVER_PORT/api/device/$DEVICE_ID/dashboard"
    log "INFO" "Fetching dashboard assignment from $url"
    
    # Use wget to fetch the assignment (curl might not be available on Kindle)
    local response=$(wget -qO- --timeout="$CONNECTION_TIMEOUT" "$url" 2>/dev/null)
    local exit_code=$?
    
    if [ $exit_code -eq 0 ] && [ -n "$response" ]; then
        # Parse JSON response to get dashboard URL
        local dashboard_url=$(echo "$response" | grep -o '"dashboard_url":"[^"]*"' | cut -d'"' -f4)
        if [ -n "$dashboard_url" ]; then
            log "INFO" "Dashboard assignment received: $dashboard_url"
            echo "$dashboard_url"
            return 0
        fi
    fi
    
    log "ERROR" "Failed to get dashboard assignment (exit code: $exit_code)"
    return 1
}

# Launch browser with dashboard
launch_browser() {
    local dashboard_url="$1"
    log "INFO" "Launching browser with URL: $dashboard_url"
    
    # Kill any existing browser instances
    killall browser 2>/dev/null || true
    
    # Launch the Kindle browser in fullscreen mode
    if [ "$FULLSCREEN" = "true" ]; then
        /usr/bin/browser -u "$dashboard_url" -f &
    else
        /usr/bin/browser -u "$dashboard_url" &
    fi
    
    local browser_pid=$!
    log "INFO" "Browser launched with PID: $browser_pid"
    
    # Store the PID for later management
    echo "$browser_pid" > /tmp/homeboard_browser.pid
}

# Launch offline dashboard
launch_offline_dashboard() {
    log "INFO" "Launching offline dashboard"
    local offline_url="file://$EXTENSION_DIR/html/hello_world.html"
    launch_browser "$offline_url"
}

# Main execution
main() {
    log "INFO" "Starting Homeboard Dashboard Launcher"
    log "INFO" "Extension directory: $EXTENSION_DIR"
    log "INFO" "Configuration: Server=$SERVER_URL:$SERVER_PORT, Device=$DEVICE_ID"
    
    # Check device configuration
    if ! check_device_config; then
        log "WARN" "Device not configured, launching configuration wizard"
        "$EXTENSION_DIR/bin/configure_server.sh"
        return $?
    fi
    
    # Test connectivity
    if test_connectivity; then
        # Try to get dashboard assignment
        dashboard_url=$(get_dashboard_assignment)
        if [ $? -eq 0 ] && [ -n "$dashboard_url" ]; then
            # Launch the assigned dashboard
            if [[ "$dashboard_url" == http* ]]; then
                launch_browser "$dashboard_url"
            else
                # Construct full URL
                full_url="http://$SERVER_URL:$SERVER_PORT$dashboard_url"
                launch_browser "$full_url"
            fi
        else
            log "WARN" "Could not get dashboard assignment, launching default dashboard"
            default_url="http://$SERVER_URL:$SERVER_PORT/"
            launch_browser "$default_url"
        fi
    else
        # Network connectivity failed
        if [ "$ENABLE_OFFLINE_MODE" = "true" ]; then
            log "INFO" "Network unavailable, launching offline mode"
            launch_offline_dashboard
        else
            log "ERROR" "Network unavailable and offline mode disabled"
            eips 10 10 "Homeboard: Network Error"
            eips 10 11 "Cannot connect to server"
            eips 10 12 "Check WiFi settings"
            return 1
        fi
    fi
    
    log "INFO" "Dashboard launcher completed"
    return 0
}

# Execute main function
main "$@"