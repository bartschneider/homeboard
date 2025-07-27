#!/bin/bash
# Homeboard KUAL Extension - Device Settings Manager
# This script manages device-specific settings and preferences

EXTENSION_DIR="/mnt/us/extensions/homeboard"
CONFIG_FILE="$EXTENSION_DIR/config/device.conf"
LOG_FILE="/tmp/homeboard_kual.log"

# Load configuration
load_config() {
    if [ -f "$CONFIG_FILE" ]; then
        source "$CONFIG_FILE"
    else
        echo "Configuration file not found: $CONFIG_FILE"
        exit 1
    fi
}

# Display current settings
show_current_settings() {
    clear
    eips 0 0 "Current Device Settings"
    eips 0 1 "========================"
    eips 0 3 "Server: $SERVER_URL:$SERVER_PORT"
    eips 0 4 "Device ID: $DEVICE_ID"
    eips 0 5 "Device Name: $DEVICE_NAME"
    eips 0 6 "Dashboard: $ASSIGNED_DASHBOARD"
    eips 0 7 "Offline Mode: $ENABLE_OFFLINE_MODE"
    eips 0 8 "Debug Mode: $DEBUG_MODE"
    eips 0 9 "Refresh Interval: ${REFRESH_INTERVAL}s"
    eips 0 11 "Options:"
    eips 0 12 "1) Reconfigure Server"
    eips 0 13 "2) Change Device Name"
    eips 0 14 "3) Toggle Offline Mode"
    eips 0 15 "4) Toggle Debug Mode"
    eips 0 16 "5) Set Refresh Interval"
    eips 0 17 "6) View Device Info"
    eips 0 18 "7) Reset to Defaults"
    eips 0 19 "0) Exit"
    eips 0 21 "Enter choice:"
}

# Get device information
get_device_info() {
    clear
    eips 0 0 "Device Information"
    eips 0 1 "=================="
    
    # Kindle model
    local model="Unknown"
    if [ -f /proc/version ]; then
        model=$(cat /proc/version | head -1)
    fi
    
    # Serial number
    local serial="Unknown"
    if [ -f /proc/usid ]; then
        serial=$(cat /proc/usid 2>/dev/null)
    elif [ -f /var/local/system/serial ]; then
        serial=$(cat /var/local/system/serial 2>/dev/null)
    fi
    
    # Memory info
    local memory="Unknown"
    if [ -f /proc/meminfo ]; then
        memory=$(grep MemTotal /proc/meminfo | awk '{print $2 " " $3}')
    fi
    
    # WiFi status
    local wifi_status="Unknown"
    if command -v iwconfig >/dev/null 2>&1; then
        if iwconfig 2>/dev/null | grep -q "ESSID:"; then
            wifi_status="Connected"
            local essid=$(iwconfig 2>/dev/null | grep "ESSID:" | cut -d'"' -f2)
            wifi_status="Connected to $essid"
        else
            wifi_status="Disconnected"
        fi
    fi
    
    # Display info
    eips 0 3 "Model: ${model:0:40}"
    eips 0 4 "Serial: $serial"
    eips 0 5 "Memory: $memory"
    eips 0 6 "WiFi: $wifi_status"
    eips 0 7 "Extension: v1.0.0"
    eips 0 8 "Config: $CONFIG_FILE"
    eips 0 9 "Log: $LOG_FILE"
    
    eips 0 15 "Press any key to continue..."
    read -r -n 1
}

# Change device name
change_device_name() {
    clear
    eips 0 0 "Change Device Name"
    eips 0 1 "=================="
    eips 0 3 "Current name: $DEVICE_NAME"
    eips 0 5 "Enter new device name:"
    
    read -r new_name
    
    if [ -n "$new_name" ]; then
        # Update config file
        sed -i "s/DEVICE_NAME=\".*\"/DEVICE_NAME=\"$new_name\"/" "$CONFIG_FILE"
        eips 0 7 "Device name updated to: $new_name"
    else
        eips 0 7 "No change made"
    fi
    
    sleep 2
}

# Toggle offline mode
toggle_offline_mode() {
    clear
    eips 0 0 "Toggle Offline Mode"
    eips 0 1 "==================="
    eips 0 3 "Current setting: $ENABLE_OFFLINE_MODE"
    
    if [ "$ENABLE_OFFLINE_MODE" = "true" ]; then
        sed -i 's/ENABLE_OFFLINE_MODE="true"/ENABLE_OFFLINE_MODE="false"/' "$CONFIG_FILE"
        eips 0 5 "Offline mode DISABLED"
    else
        sed -i 's/ENABLE_OFFLINE_MODE="false"/ENABLE_OFFLINE_MODE="true"/' "$CONFIG_FILE"
        eips 0 5 "Offline mode ENABLED"
    fi
    
    sleep 2
}

# Toggle debug mode
toggle_debug_mode() {
    clear
    eips 0 0 "Toggle Debug Mode"
    eips 0 1 "================="
    eips 0 3 "Current setting: $DEBUG_MODE"
    
    if [ "$DEBUG_MODE" = "true" ]; then
        sed -i 's/DEBUG_MODE="true"/DEBUG_MODE="false"/' "$CONFIG_FILE"
        eips 0 5 "Debug mode DISABLED"
    else
        sed -i 's/DEBUG_MODE="false"/DEBUG_MODE="true"/' "$CONFIG_FILE"
        eips 0 5 "Debug mode ENABLED"
    fi
    
    sleep 2
}

# Set refresh interval
set_refresh_interval() {
    clear
    eips 0 0 "Set Refresh Interval"
    eips 0 1 "===================="
    eips 0 3 "Current interval: ${REFRESH_INTERVAL}s"
    eips 0 5 "Common intervals:"
    eips 0 6 "  300 = 5 minutes"
    eips 0 7 "  900 = 15 minutes"
    eips 0 8 " 1800 = 30 minutes"
    eips 0 9 " 3600 = 1 hour"
    eips 0 11 "Enter new interval (seconds):"
    
    read -r new_interval
    
    if [[ "$new_interval" =~ ^[0-9]+$ ]] && [ "$new_interval" -gt 60 ]; then
        sed -i "s/REFRESH_INTERVAL=\".*\"/REFRESH_INTERVAL=\"$new_interval\"/" "$CONFIG_FILE"
        eips 0 13 "Refresh interval updated to: ${new_interval}s"
    else
        eips 0 13 "Invalid interval (must be > 60 seconds)"
    fi
    
    sleep 2
}

# Reset to defaults
reset_to_defaults() {
    clear
    eips 0 0 "Reset to Defaults"
    eips 0 1 "================="
    eips 0 3 "This will reset all settings to defaults"
    eips 0 4 "Server configuration will be preserved"
    eips 0 6 "Are you sure? (y/N):"
    
    read -r -n 1 confirm
    
    if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
        # Preserve server settings
        local server_url="$SERVER_URL"
        local server_port="$SERVER_PORT"
        local device_id="$DEVICE_ID"
        
        # Generate new config with defaults
        cat > "$CONFIG_FILE" << EOF
# Homeboard KUAL Extension Configuration
# Reset to defaults on $(date)

# Dashboard Server Configuration
SERVER_URL="$server_url"
SERVER_PORT="$server_port"
DEVICE_ID="$device_id"
DEVICE_NAME="Kindle Dashboard"
ASSIGNED_DASHBOARD=""

# Network Configuration
WIFI_SSID=""
CONNECTION_TIMEOUT="30"
RETRY_ATTEMPTS="3"

# Display Configuration
REFRESH_INTERVAL="900"
FULLSCREEN="true"
ORIENTATION="landscape"

# Fallback Configuration
ENABLE_OFFLINE_MODE="true"
OFFLINE_DASHBOARD="hello_world"

# Debug Configuration
DEBUG_MODE="false"
LOG_LEVEL="info"
LOG_FILE="/tmp/homeboard_kual.log"
EOF
        
        eips 0 8 "Settings reset to defaults"
    else
        eips 0 8 "Reset cancelled"
    fi
    
    sleep 2
}

# Main menu loop
main() {
    load_config
    
    while true; do
        show_current_settings
        read -r -n 1 choice
        
        case "$choice" in
            1)
                "$EXTENSION_DIR/bin/configure_server.sh"
                load_config
                ;;
            2)
                change_device_name
                load_config
                ;;
            3)
                toggle_offline_mode
                load_config
                ;;
            4)
                toggle_debug_mode
                load_config
                ;;
            5)
                set_refresh_interval
                load_config
                ;;
            6)
                get_device_info
                ;;
            7)
                reset_to_defaults
                load_config
                ;;
            0)
                clear
                exit 0
                ;;
            *)
                continue
                ;;
        esac
    done
}

# Execute main function
main "$@"