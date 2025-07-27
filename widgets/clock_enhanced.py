#!/usr/bin/env python3
"""
Enhanced Clock Widget for E-Paper Dashboard
Features:
- Multiple time zones support
- Enhanced typography with design system
- Date formatting options
- Clean card-based layout
"""

import json
import sys
import datetime
from typing import Dict, Any
import pytz


def execute_widget(parameters: Dict[str, Any]) -> str:
    """Execute the enhanced clock widget"""
    try:
        # Get parameters with defaults
        # time_format = parameters.get("format", "%Y-%m-%d %H:%M:%S")  # Unused for now
        timezone_str = parameters.get("timezone", "Local")
        show_seconds = parameters.get("show_seconds", True)
        show_date = parameters.get("show_date", True)
        show_timezone = parameters.get("show_timezone", True)
        date_format = parameters.get("date_format", "%A, %B %d, %Y")

        # Handle timezone
        if timezone_str == "Local":
            now = datetime.datetime.now()
            tz_info = "Local Time"
        else:
            try:
                tz = pytz.timezone(timezone_str)
                now = datetime.datetime.now(tz)
                tz_info = tz.zone
            except Exception:
                # Fallback to local time if timezone is invalid
                now = datetime.datetime.now()
                tz_info = "Local Time (invalid timezone specified)"

        # Format time components
        if show_seconds:
            time_display = now.strftime("%H:%M:%S")
        else:
            time_display = now.strftime("%H:%M")

        # Build HTML using design system classes
        html_parts = []

        # Time display section
        html_parts.append('<div class="time-display">')
        html_parts.append('  <div class="time-main">')
        html_parts.append(f'    <span class="value value--huge">{time_display}</span>')
        html_parts.append("  </div>")

        # Time metadata
        html_parts.append('  <div class="time-meta">')

        if show_date:
            date_display = now.strftime(date_format)
            html_parts.append(f'    <span class="subtitle">{date_display}</span>')

        if show_timezone:
            html_parts.append(f'    <span class="description">{tz_info}</span>')

        # Additional time info
        day_of_year = now.timetuple().tm_yday
        week_number = now.isocalendar()[1]
        html_parts.append(
            '    <div class="flex items-center justify-center gap-md mt-sm">'
        )
        html_parts.append(f'      <span class="meta">Day {day_of_year}</span>')
        html_parts.append('      <span class="meta">•</span>')
        html_parts.append(f'      <span class="meta">Week {week_number}</span>')
        html_parts.append("    </div>")

        html_parts.append("  </div>")
        html_parts.append("</div>")

        return "\n".join(html_parts)

    except Exception as e:
        # Error state using design system
        return f"""
        <div class="widget-error">
            <div class="error-icon">🕐</div>
            <div class="error-message">Clock widget error: {str(e)}</div>
            <div class="error-hint">Check timezone configuration</div>
        </div>
        """


def main():
    """Main function for widget execution"""
    try:
        # Read parameters from stdin
        params_json = sys.stdin.read().strip()
        if params_json:
            parameters = json.loads(params_json)
        else:
            parameters = {}

        # Execute widget and return HTML
        html_output = execute_widget(parameters)
        print(html_output)

    except Exception as e:
        # Fallback error display
        print(
            f"""
        <div class="widget-error">
            <div class="error-icon">⚠️</div>
            <div class="error-message">Clock widget execution failed: {str(e)}</div>
            <div class="error-hint">Check widget configuration and environment</div>
        </div>
        """
        )
        sys.exit(1)


if __name__ == "__main__":
    main()
