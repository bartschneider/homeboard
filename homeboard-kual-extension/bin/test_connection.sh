#!/bin/bash
# Homeboard KUAL Extension - Connection Test
# This script tests the connection to the homeboard server

EXTENSION_DIR="/mnt/us/extensions/homeboard"
CONFIG_FILE="$EXTENSION_DIR/config/device.conf"
LOG_FILE="/tmp/homeboard_kual.log"

# Load configuration
if [ -f "$CONFIG_FILE" ]; then
    source "$CONFIG_FILE"
else
    clear
    eips 0 0 "Configuration Error"
    eips 0 1 "No configuration found"
    eips 0 2 "Run 'Configure Server' first"
    eips 0 4 "Press any key to exit..."
    read -r -n 1
    exit 1
fi

# Test functions
test_wifi_connection() {
    eips 0 3 "Testing WiFi connection..."
    
    if command -v iwconfig >/dev/null 2>&1; then
        if iwconfig 2>/dev/null | grep -q "ESSID:"; then
            local essid=$(iwconfig 2>/dev/null | grep "ESSID:" | cut -d'"' -f2)
            eips 0 4 "✓ WiFi connected to: $essid"
            return 0
        else
            eips 0 4 "✗ WiFi not connected"
            return 1
        fi
    else
        eips 0 4 "? WiFi status unknown"
        return 1
    fi
}

test_dns_resolution() {
    eips 0 6 "Testing DNS resolution..."
    
    if [ -z "$SERVER_URL" ]; then
        eips 0 7 "✗ No server URL configured"
        return 1
    fi
    
    # Try to resolve the hostname
    if nslookup "$SERVER_URL" >/dev/null 2>&1; then
        eips 0 7 "✓ DNS resolution successful"
        return 0
    else
        eips 0 7 "✗ Cannot resolve: $SERVER_URL"
        return 1
    fi
}

test_ping_connectivity() {
    eips 0 9 "Testing ping connectivity..."
    
    if ping -c 1 -W 5 "$SERVER_URL" >/dev/null 2>&1; then
        eips 0 10 "✓ Server is reachable"
        return 0
    else
        eips 0 10 "✗ Server not reachable"
        return 1
    fi
}

test_port_connection() {
    eips 0 12 "Testing port connection..."
    
    if nc -z -w5 "$SERVER_URL" "$SERVER_PORT" 2>/dev/null; then
        eips 0 13 "✓ Port $SERVER_PORT is open"
        return 0
    else
        eips 0 13 "✗ Port $SERVER_PORT is closed/filtered"
        return 1
    fi
}

test_http_response() {
    eips 0 15 "Testing HTTP response..."
    
    local test_url="http://$SERVER_URL:$SERVER_PORT/"
    local response_code=$(wget -qO- --spider --server-response "$test_url" 2>&1 | grep "HTTP/" | tail -1 | awk '{print $2}')
    
    if [ "$response_code" = "200" ]; then
        eips 0 16 "✓ HTTP server responding (200 OK)"
        return 0
    elif [ -n "$response_code" ]; then
        eips 0 16 "! HTTP server responding ($response_code)"
        return 0
    else
        eips 0 16 "✗ No HTTP response"
        return 1
    fi
}

test_api_endpoint() {
    eips 0 18 "Testing API endpoint..."
    
    local api_url="http://$SERVER_URL:$SERVER_PORT/api/health"
    local response=$(wget -qO- --timeout=10 "$api_url" 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        eips 0 19 "✓ API endpoint accessible"
        return 0
    else
        eips 0 19 "✗ API endpoint not accessible"
        return 1
    fi
}

test_device_registration() {
    eips 0 21 "Testing device registration..."
    
    local register_url="http://$SERVER_URL:$SERVER_PORT/api/devices/$DEVICE_ID"
    local device_info=$(wget -qO- --timeout=10 "$register_url" 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$device_info" ]; then
        eips 0 22 "✓ Device is registered"
        return 0
    else
        eips 0 22 "? Device registration unclear"
        return 1
    fi
}

# Main test sequence
main() {
    clear
    eips 0 0 "Homeboard Connection Test"
    eips 0 1 "========================="
    eips 0 2 "Server: $SERVER_URL:$SERVER_PORT"
    
    local tests_passed=0
    local total_tests=7
    
    # Run all tests
    test_wifi_connection && tests_passed=$((tests_passed + 1))
    test_dns_resolution && tests_passed=$((tests_passed + 1))
    test_ping_connectivity && tests_passed=$((tests_passed + 1))
    test_port_connection && tests_passed=$((tests_passed + 1))
    test_http_response && tests_passed=$((tests_passed + 1))
    test_api_endpoint && tests_passed=$((tests_passed + 1))
    test_device_registration && tests_passed=$((tests_passed + 1))
    
    # Show results
    eips 0 24 "========================="
    eips 0 25 "Tests passed: $tests_passed/$total_tests"
    
    if [ $tests_passed -eq $total_tests ]; then
        eips 0 26 "Status: ALL TESTS PASSED"
        eips 0 27 "Your connection is working!"
    elif [ $tests_passed -ge 4 ]; then
        eips 0 26 "Status: MOSTLY WORKING"
        eips 0 27 "Some issues detected"
    else
        eips 0 26 "Status: CONNECTION PROBLEMS"
        eips 0 27 "Check network and server"
    fi
    
    eips 0 29 "Press 'L' to launch dashboard"
    eips 0 30 "Press 'S' to configure server"
    eips 0 31 "Press any other key to exit"
    
    read -r -n 1 choice
    
    case "$choice" in
        l|L)
            "$EXTENSION_DIR/bin/launch_dashboard.sh"
            ;;
        s|S)
            "$EXTENSION_DIR/bin/configure_server.sh"
            ;;
        *)
            clear
            ;;
    esac
}

# Execute main function
main "$@"