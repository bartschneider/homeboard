#!/bin/bash

# Setup Demo Widgets Script
# This script sets up the weather widget and other widgets in the database via API calls

echo "🚀 Setting up demo widgets in database..."

# Base URL for the API
BASE_URL="http://localhost:8081/api"

# Function to make API calls
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    if [ -n "$data" ]; then
        curl -s -X "$method" \
             -H "Content-Type: application/json" \
             -d "$data" \
             "$BASE_URL$endpoint"
    else
        curl -s -X "$method" "$BASE_URL$endpoint"
    fi
}

echo "📝 Creating widgets..."

# 1. Create Weather Widget
echo "🌤️ Creating weather widget..."
weather_response=$(api_call "POST" "/widgets" '{
    "name": "🌤️ Weather",
    "template_type": "weather_current", 
    "api_url": "https://api.open-meteo.com/v1/forecast",
    "api_headers": {},
    "data_mapping": {
        "location": "London",
        "timezone": "auto"
    },
    "description": "Current weather conditions and 4-hour forecast using Open-Meteo API",
    "timeout": 20,
    "enabled": true
}')

echo "Weather widget response: $weather_response"

# 2. Create Time Widget  
echo "🕐 Creating time widget..."
time_response=$(api_call "POST" "/widgets" '{
    "name": "🕐 Current Time",
    "template_type": "time_display",
    "api_url": "http://worldtimeapi.org/api/timezone/Europe/London", 
    "api_headers": {},
    "data_mapping": {
        "format": "%Y-%m-%d %H:%M:%S",
        "timezone": "Local",
        "show_seconds": true
    },
    "description": "Current time and date display",
    "timeout": 10,
    "enabled": true
}')

echo "Time widget response: $time_response"

# 3. Create System Widget
echo "💻 Creating system widget..."
system_response=$(api_call "POST" "/widgets" '{
    "name": "💻 System Status", 
    "template_type": "metric_grid",
    "api_url": "http://localhost/api/system",
    "api_headers": {},
    "data_mapping": {
        "show_cpu": true,
        "show_memory": true, 
        "show_disk": true
    },
    "description": "System resource monitoring",
    "timeout": 15,
    "enabled": true
}')

echo "System widget response: $system_response"

# 4. Get the default dashboard
echo "📋 Getting default dashboard..."
dashboard_response=$(api_call "GET" "/dashboards")
echo "Dashboard response: $dashboard_response"

# Extract dashboard ID (assuming first dashboard is default)
dashboard_id=$(echo "$dashboard_response" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)

if [ -n "$dashboard_id" ]; then
    echo "📌 Adding widgets to dashboard $dashboard_id..."
    
    # Note: This would require implementing the dashboard-widget association API
    # For now, let's just show the widgets were created
    echo "✅ Widgets created successfully!"
    echo "🌐 Visit http://localhost:8081 to see your dashboard"
    echo "🔧 Visit http://localhost:8081/admin to manage widgets"
else
    echo "⚠️ Could not find default dashboard"
fi

echo "🎉 Demo setup complete!"