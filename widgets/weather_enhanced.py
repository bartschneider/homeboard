#!/usr/bin/env python3
"""
Enhanced Weather Widget for E-Paper Dashboard
Uses Open-Meteo API (no API key required) with TRMNL-inspired design
Displays current weather and 4-hour forecast with professional layout
"""

import json
import sys
from typing import Dict, Any
from datetime import datetime

try:
    import requests

    HAS_REQUESTS = True
except ImportError:
    HAS_REQUESTS = False


def load_parameters() -> Dict[str, Any]:
    """Load and parse widget parameters from command line argument"""
    if len(sys.argv) < 2:
        return {}

    try:
        return json.loads(sys.argv[1])
    except (json.JSONDecodeError, IndexError):
        return {}


def get_coordinates(location: str) -> Dict[str, Any]:
    """Get coordinates for a location using Open-Meteo Geocoding API"""
    if not HAS_REQUESTS:
        return {"error": "requests library not available"}

    try:
        # Use Open-Meteo Geocoding API (free, no key required)
        url = "https://geocoding-api.open-meteo.com/v1/search"
        params = {"name": location, "count": 1, "language": "en", "format": "json"}

        response = requests.get(url, params=params, timeout=10)
        response.raise_for_status()

        data = response.json()

        if not data.get("results"):
            return {"error": f"Location '{location}' not found"}

        result = data["results"][0]
        return {
            "latitude": result["latitude"],
            "longitude": result["longitude"],
            "name": result["name"],
            "country": result.get("country", ""),
            "admin1": result.get("admin1", ""),
        }

    except requests.exceptions.RequestException as e:
        return {"error": f"Geocoding error: {str(e)}"}
    except Exception as e:
        return {"error": f"Location lookup error: {str(e)}"}


def get_weather_data(
    latitude: float, longitude: float, timezone: str = "auto"
) -> Dict[str, Any]:
    """Fetch weather data from Open-Meteo API"""
    if not HAS_REQUESTS:
        return {"error": "requests library not available"}

    try:
        # Open-Meteo API - free, no API key required
        url = "https://api.open-meteo.com/v1/forecast"
        params = {
            "latitude": latitude,
            "longitude": longitude,
            "current": [
                "temperature_2m",
                "relative_humidity_2m",
                "apparent_temperature",
                "weather_code",
                "wind_speed_10m",
            ],
            "hourly": ["temperature_2m", "weather_code"],
            "timezone": timezone,
            "forecast_days": 1,
        }

        response = requests.get(url, params=params, timeout=15)
        response.raise_for_status()

        data = response.json()

        # Process current weather
        current = data["current"]
        current_weather = {
            "temperature": round(current["temperature_2m"]),
            "feels_like": round(current["apparent_temperature"]),
            "humidity": current["relative_humidity_2m"],
            "wind_speed": current["wind_speed_10m"],
            "weather_code": current["weather_code"],
            "time": current["time"],
        }

        # Process hourly forecast (next 4 hours)
        hourly = data["hourly"]
        current_time = datetime.fromisoformat(current["time"].replace("Z", "+00:00"))

        hourly_forecast = []
        for i in range(len(hourly["time"])):
            hour_time = datetime.fromisoformat(hourly["time"][i].replace("Z", "+00:00"))
            if hour_time > current_time and len(hourly_forecast) < 4:
                hourly_forecast.append(
                    {
                        "time": hour_time.strftime("%H:%M"),
                        "temperature": round(hourly["temperature_2m"][i]),
                        "weather_code": hourly["weather_code"][i],
                    }
                )

        # If no future hours available, get next 4 starting from current hour
        if not hourly_forecast:
            for i in range(min(4, len(hourly["time"]))):
                hour_time = datetime.fromisoformat(
                    hourly["time"][i].replace("Z", "+00:00")
                )
                hourly_forecast.append(
                    {
                        "time": hour_time.strftime("%H:%M"),
                        "temperature": round(hourly["temperature_2m"][i]),
                        "weather_code": hourly["weather_code"][i],
                    }
                )

        return {"current": current_weather, "hourly": hourly_forecast}

    except requests.exceptions.RequestException as e:
        return {"error": f"Weather API error: {str(e)}"}
    except KeyError as e:
        return {"error": f"Invalid weather response: {str(e)}"}
    except Exception as e:
        return {"error": f"Weather data error: {str(e)}"}


def get_weather_info(weather_code: int) -> Dict[str, str]:
    """Convert WMO weather code to description and emoji"""
    # WMO Weather interpretation codes
    weather_map = {
        0: {"description": "Clear Sky", "emoji": "☀️"},
        1: {"description": "Mainly Clear", "emoji": "🌤️"},
        2: {"description": "Partly Cloudy", "emoji": "⛅"},
        3: {"description": "Overcast", "emoji": "☁️"},
        45: {"description": "Fog", "emoji": "🌫️"},
        48: {"description": "Freezing Fog", "emoji": "🌫️"},
        51: {"description": "Light Drizzle", "emoji": "🌦️"},
        53: {"description": "Moderate Drizzle", "emoji": "🌦️"},
        55: {"description": "Dense Drizzle", "emoji": "🌧️"},
        56: {"description": "Light Freezing Drizzle", "emoji": "🌨️"},
        57: {"description": "Dense Freezing Drizzle", "emoji": "🌨️"},
        61: {"description": "Light Rain", "emoji": "🌦️"},
        63: {"description": "Moderate Rain", "emoji": "🌧️"},
        65: {"description": "Heavy Rain", "emoji": "🌧️"},
        66: {"description": "Light Freezing Rain", "emoji": "🌨️"},
        67: {"description": "Heavy Freezing Rain", "emoji": "🌨️"},
        71: {"description": "Light Snow", "emoji": "🌨️"},
        73: {"description": "Moderate Snow", "emoji": "❄️"},
        75: {"description": "Heavy Snow", "emoji": "❄️"},
        77: {"description": "Snow Grains", "emoji": "🌨️"},
        80: {"description": "Light Showers", "emoji": "🌦️"},
        81: {"description": "Moderate Showers", "emoji": "🌧️"},
        82: {"description": "Heavy Showers", "emoji": "🌧️"},
        85: {"description": "Light Snow Showers", "emoji": "🌨️"},
        86: {"description": "Heavy Snow Showers", "emoji": "❄️"},
        95: {"description": "Thunderstorm", "emoji": "⛈️"},
        96: {"description": "Thunderstorm with Hail", "emoji": "⛈️"},
        99: {"description": "Heavy Thunderstorm", "emoji": "⛈️"},
    }

    return weather_map.get(weather_code, {"description": "Unknown", "emoji": "🌤️"})


def generate_html(location_data: Dict[str, Any], weather_data: Dict[str, Any]) -> str:
    """Generate HTML output using TRMNL-inspired design system"""
    if "error" in location_data:
        return f"""
        <div class="widget-error">
            <div class="error-icon">🌤️</div>
            <div class="error-message">Location Error: {location_data['error']}</div>
            <div class="error-hint">Check location name and network connection</div>
        </div>
        """

    if "error" in weather_data:
        return f"""
        <div class="widget-error">
            <div class="error-icon">🌤️</div>
            <div class="error-message">Weather Error: {weather_data['error']}</div>
            <div class="error-hint">Check network connection and try again</div>
        </div>
        """

    current = weather_data["current"]
    hourly = weather_data["hourly"]

    # Get weather info for current conditions
    weather_info = get_weather_info(current["weather_code"])

    # Location display with admin1 (state/region) if available
    location_display = location_data["name"]
    if location_data.get("admin1") and location_data["admin1"] != location_data["name"]:
        location_display += f", {location_data['admin1']}"

    html = f"""
    <div class="weather-current">
        <div class="weather-icon">
            <span style="font-size: 48px;">{weather_info['emoji']}</span>
        </div>
        <div class="weather-temp">
            <div class="value value--huge">{current['temperature']}°</div>
            <div class="description">{weather_info['description']}</div>
            <div class="meta">Feels like {current['feels_like']}°C</div>
        </div>
    </div>
    <div class="weather-divider"></div>

    <div class="metrics-grid">
        <div class="metric-item">
            <div class="metric-icon">💧</div>
            <div class="metric-info">
                <div class="label">Humidity</div>
                <div class="value value--small">{current['humidity']}%</div>
            </div>
        </div>
        <div class="metric-item">
            <div class="metric-icon">💨</div>
            <div class="metric-info">
                <div class="label">Wind Speed</div>
                <div class="value value--small">{current['wind_speed']:.1f} km/h</div>
            </div>
        </div>
    </div>

    <div class="weather-divider"></div>

    <div class="subtitle subtitle--small text-center mb-sm">Next 4 Hours</div>
    <div class="hourly-grid">
    """

    # Add hourly forecast
    for hour in hourly:
        hour_weather = get_weather_info(hour["weather_code"])
        html += f"""
        <div class="hour-item">
            <div class="meta">{hour['time']}</div>
            <div style="font-size: 20px; margin: 4px 0;">{hour_weather['emoji']}</div>
            <div class="value value--small">{hour['temperature']}°</div>
        </div>
        """

    html += (
        """
    </div>

    <div class="mt-md text-center">
        <div class="meta">"""
        + location_display
        + """</div>
        <div class="meta">Updated: """
        + datetime.now().strftime("%H:%M")
        + """</div>
    </div>
    """
    )

    return html


def main():
    """Main widget execution function"""
    try:
        # Load parameters
        params = load_parameters()

        # Get configuration with defaults
        location = params.get("location", "London")
        timezone = params.get("timezone", "auto")

        # Get coordinates for the location
        location_data = get_coordinates(location)
        if "error" in location_data:
            print(generate_html(location_data, {}))
            return

        # Get weather data
        weather_data = get_weather_data(
            location_data["latitude"], location_data["longitude"], timezone
        )

        # Generate and output HTML
        html_output = generate_html(location_data, weather_data)
        print(html_output)

    except Exception as e:
        # Error handling - output user-friendly error message
        error_html = f"""
        <div class="widget-error">
            <div class="error-icon">🌤️</div>
            <div class="error-message">Widget Error: {str(e)}</div>
            <div class="error-hint">Check widget configuration and try again</div>
        </div>
        """
        print(error_html)


if __name__ == "__main__":
    main()
