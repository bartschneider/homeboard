#!/usr/bin/env python3
"""
Weather widget for E-Paper Dashboard
Displays current weather conditions using OpenWeatherMap API
"""

import json
import sys
from typing import Dict, Any

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


def get_weather_data(
    api_key: str, location: str, units: str = "metric"
) -> Dict[str, Any]:
    """Fetch weather data from OpenWeatherMap API"""
    if not HAS_REQUESTS:
        return {
            "error": "requests library not available - install with: "
            "pip3 install requests"
        }

    if not api_key:
        return {"error": "API key required - get one from openweathermap.org"}

    try:
        url = "http://api.openweathermap.org/data/2.5/weather"
        params = {"q": location, "appid": api_key, "units": units}

        response = requests.get(url, params=params, timeout=10)
        response.raise_for_status()

        data = response.json()

        # Extract relevant information
        weather_info = {
            "location": data["name"],
            "country": data["sys"]["country"],
            "temperature": round(data["main"]["temp"]),
            "feels_like": round(data["main"]["feels_like"]),
            "humidity": data["main"]["humidity"],
            "description": data["weather"][0]["description"].title(),
            "icon": data["weather"][0]["icon"],
            "wind_speed": data.get("wind", {}).get("speed", 0),
        }

        return weather_info

    except requests.exceptions.RequestException as e:
        return {"error": f"Network error: {str(e)}"}
    except KeyError as e:
        return {"error": f"Invalid API response: {str(e)}"}
    except Exception as e:
        return {"error": f"Weather API error: {str(e)}"}


def get_weather_emoji(icon_code: str) -> str:
    """Convert OpenWeatherMap icon code to emoji"""
    icon_map = {
        "01d": "☀️",  # clear sky day
        "01n": "🌙",  # clear sky night
        "02d": "⛅",  # few clouds day
        "02n": "☁️",  # few clouds night
        "03d": "☁️",  # scattered clouds
        "03n": "☁️",  # scattered clouds
        "04d": "☁️",  # broken clouds
        "04n": "☁️",  # broken clouds
        "09d": "🌧️",  # shower rain
        "09n": "🌧️",  # shower rain
        "10d": "🌦️",  # rain day
        "10n": "🌧️",  # rain night
        "11d": "⛈️",  # thunderstorm
        "11n": "⛈️",  # thunderstorm
        "13d": "🌨️",  # snow
        "13n": "🌨️",  # snow
        "50d": "🌫️",  # mist
        "50n": "🌫️",  # mist
    }
    return icon_map.get(icon_code, "🌤️")


def generate_html(weather_data: Dict[str, Any], units: str) -> str:
    """Generate HTML output for the weather widget"""
    if "error" in weather_data:
        return f"""
        <div class="weather-widget">
            <h2>🌤️ Weather</h2>
            <div style="text-align: center; color: #666;">
                ⚠️ Error: {weather_data['error']}<br>
                <small>Check API key and network connection</small>
            </div>
        </div>
        """

    # Determine temperature unit symbol
    temp_unit = "°C" if units == "metric" else "°F" if units == "imperial" else "K"
    wind_unit = "m/s" if units == "metric" else "mph" if units == "imperial" else "m/s"

    emoji = get_weather_emoji(weather_data.get("icon", ""))

    html = f"""
    <div class="weather-widget">
        <h2>🌤️ Weather</h2>
        <div style="text-align: center;">
            <div style="font-size: 1.1em; margin-bottom: 8px;">
                <strong>{weather_data['location']}, {weather_data['country']}</strong>
            </div>
            <div style="display: flex; justify-content: space-between; \
align-items: center; margin: 10px 0;">
                <div style="font-size: 2em;">{emoji}</div>
                <div style="text-align: right;">
                    <div style="font-size: 1.4em; font-weight: bold;">
                        {weather_data['temperature']}{temp_unit}
                    </div>
                    <div style="font-size: 0.9em; color: #666;">
                        feels like {weather_data['feels_like']}{temp_unit}
                    </div>
                </div>
            </div>
            <div style="margin: 8px 0;">
                <strong>{weather_data['description']}</strong>
            </div>
            <div style="display: flex; justify-content: space-between; \
font-size: 0.9em;">
                <span>💧 {weather_data['humidity']}%</span>
                <span>💨 {weather_data['wind_speed']:.1f} {wind_unit}</span>
            </div>
        </div>
    </div>
    """

    return html


def main():
    """Main widget execution function"""
    try:
        # Load parameters
        params = load_parameters()

        # Get configuration with defaults
        api_key = params.get("api_key", "")
        location = params.get("location", "London")
        units = params.get("units", "metric")  # metric, imperial, or kelvin

        # Get weather data
        weather_data = get_weather_data(api_key, location, units)

        # Generate and output HTML
        html_output = generate_html(weather_data, units)
        print(html_output)

    except Exception as e:
        # Error handling - output user-friendly error message
        error_html = f"""
        <div class="weather-widget">
            <h2>🌤️ Weather</h2>
            <div style="text-align: center; color: #666;">
                ⚠️ Error: {str(e)}<br>
                <small>Check widget configuration</small>
            </div>
        </div>
        """
        print(error_html)


if __name__ == "__main__":
    main()
