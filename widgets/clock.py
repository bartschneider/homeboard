#!/usr/bin/env python3
"""
Clock widget for E-Paper Dashboard
Displays current date and time with configurable formatting
"""

import json
import sys
import datetime
from typing import Dict, Any

try:
    import pytz

    HAS_PYTZ = True
except ImportError:
    HAS_PYTZ = False


def load_parameters() -> Dict[str, Any]:
    """Load and parse widget parameters from command line argument"""
    if len(sys.argv) < 2:
        return {}

    try:
        return json.loads(sys.argv[1])
    except (json.JSONDecodeError, IndexError):
        return {}


def get_current_time(timezone_name: str, time_format: str) -> str:
    """Get current time formatted according to parameters"""
    try:
        if timezone_name.lower() == "local" or not HAS_PYTZ:
            now = datetime.datetime.now()
        else:
            if HAS_PYTZ:
                tz = pytz.timezone(timezone_name)
                now = datetime.datetime.now(tz)
            else:
                # Fallback to local time if pytz not available
                now = datetime.datetime.now()

        return now.strftime(time_format)
    except Exception:
        # Fallback to simple local time
        return datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")


def generate_html(current_time: str) -> str:
    """Generate HTML output for the clock widget"""
    return f"""
    <div class="clock-widget">
        <h2>📅 Current Time</h2>
        <div style="font-size: 1.2em; font-weight: bold; \
text-align: center; margin-top: 10px;">
            {current_time}
        </div>
    </div>
    """


def main():
    """Main widget execution function"""
    try:
        # Load parameters
        params = load_parameters()

        # Get configuration with defaults
        timezone = params.get("timezone", "Local")
        time_format = params.get("format", "%Y-%m-%d %H:%M:%S")

        # Get current time
        current_time = get_current_time(timezone, time_format)

        # Generate and output HTML
        html_output = generate_html(current_time)
        print(html_output)

    except Exception as e:
        # Error handling - output user-friendly error message
        error_html = f"""
        <div class="clock-widget">
            <h2>📅 Clock</h2>
            <div style="text-align: center; color: #666;">
                ⚠️ Error: {str(e)}<br>
                <small>Check widget configuration</small>
            </div>
        </div>
        """
        print(error_html)


if __name__ == "__main__":
    main()
