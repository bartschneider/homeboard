#!/bin/bash
# Homeboard KUAL Extension - Server Configuration
# This script helps configure the connection to the homeboard server

EXTENSION_DIR="/mnt/us/extensions/homeboard"
CONFIG_FILE="$EXTENSION_DIR/config/device.conf"
TEMP_CONFIG="/tmp/homeboard_config.tmp"

# Kindle UI helper functions
show_message() {
    local title="$1"
    local message="$2"
    eips 0 0 "$(printf '%-40s' ' ')"  # Clear line
    eips 0 0 "$title"
    eips 0 1 "$message"
}

get_user_input() {
    local prompt="$1"
    local default="$2"
    local input=""
    
    show_message "Homeboard Config" "$prompt"
    if [ -n "$default" ]; then
        eips 0 3 "Default: $default"
        eips 0 4 "Press Enter for default, or type new value:"
    else
        eips 0 3 "Enter value:"
    fi
    
    # For simplicity, we'll use a basic input method
    # In a real implementation, you might use a more sophisticated input method
    read -r input
    
    if [ -z "$input" ] && [ -n "$default" ]; then
        echo "$default"
    else
        echo "$input"
    fi
}

# Load current configuration
load_current_config() {
    if [ -f "$CONFIG_FILE" ]; then
        source "$CONFIG_FILE"
    fi
}

# Save configuration
save_config() {
    cat > "$CONFIG_FILE" << EOF
# Homeboard KUAL Extension Configuration
# Generated on $(date)

# Dashboard Server Configuration
SERVER_URL="$NEW_SERVER_URL"
SERVER_PORT="$NEW_SERVER_PORT"
DEVICE_ID="$NEW_DEVICE_ID"
DEVICE_NAME="$NEW_DEVICE_NAME"
ASSIGNED_DASHBOARD="$NEW_ASSIGNED_DASHBOARD"

# Network Configuration
WIFI_SSID="$NEW_WIFI_SSID"
CONNECTION_TIMEOUT="$NEW_CONNECTION_TIMEOUT"
RETRY_ATTEMPTS="$NEW_RETRY_ATTEMPTS"

# Display Configuration
REFRESH_INTERVAL="$NEW_REFRESH_INTERVAL"
FULLSCREEN="$NEW_FULLSCREEN"
ORIENTATION="$NEW_ORIENTATION"

# Fallback Configuration
ENABLE_OFFLINE_MODE="$NEW_ENABLE_OFFLINE_MODE"
OFFLINE_DASHBOARD="$NEW_OFFLINE_DASHBOARD"

# Debug Configuration
DEBUG_MODE="$NEW_DEBUG_MODE"
LOG_LEVEL="$NEW_LOG_LEVEL"
LOG_FILE="$NEW_LOG_FILE"
EOF
}

# Generate device ID based on Kindle serial
generate_device_id() {
    local kindle_serial=""
    
    # Try to get Kindle serial number
    if [ -f /proc/usid ]; then
        kindle_serial=$(cat /proc/usid 2>/dev/null)
    elif [ -f /var/local/system/serial ]; then
        kindle_serial=$(cat /var/local/system/serial 2>/dev/null)
    fi
    
    if [ -n "$kindle_serial" ]; then
        echo "kindle_${kindle_serial}"
    else
        # Fallback to MAC address or random ID
        local mac_addr=$(ifconfig | grep -o -E '([[:xdigit:]]{1,2}:){5}[[:xdigit:]]{1,2}' | head -1 | tr -d ':')
        if [ -n "$mac_addr" ]; then
            echo "kindle_${mac_addr}"
        else
            echo "kindle_$(date +%s)_$(shuf -i 1000-9999 -n 1)"
        fi
    fi
}

# Test server connection
test_server_connection() {
    local server_url="$1"
    local server_port="$2"
    
    show_message "Testing Connection" "Connecting to $server_url:$server_port..."
    
    if ping -c 1 -W 5 "$server_url" > /dev/null 2>&1; then
        if nc -z -w5 "$server_url" "$server_port" 2>/dev/null; then
            show_message "Success" "Connection successful!"
            sleep 2
            return 0
        else
            show_message "Error" "Port $server_port not reachable"
            sleep 3
            return 1
        fi
    else
        show_message "Error" "Cannot reach $server_url"
        sleep 3
        return 1
    fi
}

# Register device with server
register_device() {
    local server_url="$1"
    local server_port="$2"
    local device_id="$3"
    local device_name="$4"
    
    show_message "Registering Device" "Registering with server..."
    
    # Create registration payload
    local payload="{\"device_id\":\"$device_id\",\"device_name\":\"$device_name\",\"device_type\":\"kindle\",\"capabilities\":[\"dashboard_display\",\"e_ink\"]}"
    
    # Register device with server
    local response=$(wget -qO- --timeout=30 \
        --header="Content-Type: application/json" \
        --post-data="$payload" \
        "http://$server_url:$server_port/api/devices/register" 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        show_message "Success" "Device registered successfully"
        sleep 2
        return 0
    else
        show_message "Warning" "Registration failed, but continuing..."
        sleep 2
        return 0  # Don't fail the setup for registration issues
    fi
}

# Main configuration wizard
main() {
    clear
    show_message "Homeboard Setup" "Welcome to Homeboard configuration"
    sleep 2
    
    # Load current configuration
    load_current_config
    
    # Configure server URL
    NEW_SERVER_URL=$(get_user_input "Enter server IP/hostname:" "$SERVER_URL")
    
    # Configure server port
    NEW_SERVER_PORT=$(get_user_input "Enter server port:" "${SERVER_PORT:-8080}")
    
    # Generate or configure device ID
    if [ -z "$DEVICE_ID" ]; then
        NEW_DEVICE_ID=$(generate_device_id)
        show_message "Device ID" "Generated: $NEW_DEVICE_ID"
        sleep 2
    else
        NEW_DEVICE_ID=$(get_user_input "Device ID:" "$DEVICE_ID")
    fi
    
    # Configure device name
    NEW_DEVICE_NAME=$(get_user_input "Device name:" "${DEVICE_NAME:-Kindle Dashboard}")
    
    # Set other defaults
    NEW_ASSIGNED_DASHBOARD="${ASSIGNED_DASHBOARD:-}"
    NEW_WIFI_SSID="${WIFI_SSID:-}"
    NEW_CONNECTION_TIMEOUT="${CONNECTION_TIMEOUT:-30}"
    NEW_RETRY_ATTEMPTS="${RETRY_ATTEMPTS:-3}"
    NEW_REFRESH_INTERVAL="${REFRESH_INTERVAL:-900}"
    NEW_FULLSCREEN="${FULLSCREEN:-true}"
    NEW_ORIENTATION="${ORIENTATION:-landscape}"
    NEW_ENABLE_OFFLINE_MODE="${ENABLE_OFFLINE_MODE:-true}"
    NEW_OFFLINE_DASHBOARD="${OFFLINE_DASHBOARD:-hello_world}"
    NEW_DEBUG_MODE="${DEBUG_MODE:-false}"
    NEW_LOG_LEVEL="${LOG_LEVEL:-info}"
    NEW_LOG_FILE="${LOG_FILE:-/tmp/homeboard_kual.log}"
    
    # Test connection if server details provided
    if [ -n "$NEW_SERVER_URL" ] && [ -n "$NEW_SERVER_PORT" ]; then
        if test_server_connection "$NEW_SERVER_URL" "$NEW_SERVER_PORT"; then
            register_device "$NEW_SERVER_URL" "$NEW_SERVER_PORT" "$NEW_DEVICE_ID" "$NEW_DEVICE_NAME"
        fi
    fi
    
    # Save configuration
    save_config
    
    show_message "Setup Complete" "Configuration saved successfully"
    sleep 2
    
    # Ask if user wants to launch dashboard now
    show_message "Launch Dashboard?" "Press Enter to launch now, or any key to exit"
    read -r -n 1 response
    
    if [ -z "$response" ]; then
        "$EXTENSION_DIR/bin/launch_dashboard.sh"
    fi
}

# Execute main function
main "$@"