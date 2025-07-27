#!/bin/bash
# Homeboard KUAL Extension - Hello World Dashboard Launcher
# This script launches the offline hello world dashboard

EXTENSION_DIR="/mnt/us/extensions/homeboard"
LOG_FILE="/tmp/homeboard_kual.log"

# Logging function
log() {
    local level="$1"
    local message="$2"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $message" | tee -a "$LOG_FILE"
}

# Launch hello world dashboard
launch_hello_world() {
    log "INFO" "Launching Hello World dashboard (offline mode)"
    
    local html_file="$EXTENSION_DIR/html/hello_world.html"
    
    # Check if HTML file exists
    if [ ! -f "$html_file" ]; then
        log "ERROR" "Hello World HTML file not found: $html_file"
        eips 10 10 "Homeboard: File Not Found"
        eips 10 11 "hello_world.html missing"
        return 1
    fi
    
    # Kill any existing browser instances
    killall browser 2>/dev/null || true
    
    # Launch the browser with the local HTML file
    local file_url="file://$html_file"
    log "INFO" "Opening browser with URL: $file_url"
    
    # Launch browser in fullscreen mode
    /usr/bin/browser -u "$file_url" -f &
    local browser_pid=$!
    
    if [ $? -eq 0 ]; then
        log "INFO" "Hello World dashboard launched successfully (PID: $browser_pid)"
        echo "$browser_pid" > /tmp/homeboard_browser.pid
        
        # Show success message on e-ink display
        eips 10 10 "Homeboard: Offline Mode"
        eips 10 11 "Hello World Dashboard"
        eips 10 12 "Press 'R' to refresh"
    else
        log "ERROR" "Failed to launch browser"
        eips 10 10 "Homeboard: Launch Error"
        eips 10 11 "Failed to start browser"
        return 1
    fi
}

# Main execution
main() {
    log "INFO" "Starting Hello World Dashboard"
    launch_hello_world
    return $?
}

# Execute main function
main "$@"