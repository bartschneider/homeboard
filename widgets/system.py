#!/usr/bin/env python3
"""
System status widget for E-Paper Dashboard
Displays CPU, memory, and disk usage information
"""

import json
import sys
import platform
from typing import Dict, Any

try:
    import psutil

    HAS_PSUTIL = True
except ImportError:
    HAS_PSUTIL = False


def load_parameters() -> Dict[str, Any]:
    """Load and parse widget parameters from command line argument"""
    if len(sys.argv) < 2:
        return {}

    try:
        return json.loads(sys.argv[1])
    except (json.JSONDecodeError, IndexError):
        return {}


def get_system_info(params: Dict[str, Any]) -> Dict[str, Any]:
    """Collect system information based on parameters"""
    info = {}

    if not HAS_PSUTIL:
        info["error"] = (
            "psutil library not available - install with: pip3 install psutil"
        )
        return info

    try:
        # CPU information
        if params.get("show_cpu", True):
            cpu_percent = psutil.cpu_percent(interval=1)
            cpu_count = psutil.cpu_count()
            info["cpu"] = {"usage": cpu_percent, "cores": cpu_count}

        # Memory information
        if params.get("show_memory", True):
            memory = psutil.virtual_memory()
            info["memory"] = {
                "used_gb": round(memory.used / (1024**3), 1),
                "total_gb": round(memory.total / (1024**3), 1),
                "percent": memory.percent,
            }

        # Disk information
        if params.get("show_disk", True):
            disk = psutil.disk_usage("/")
            info["disk"] = {
                "used_gb": round(disk.used / (1024**3), 1),
                "total_gb": round(disk.total / (1024**3), 1),
                "percent": round((disk.used / disk.total) * 100, 1),
            }

        # System uptime
        boot_time = psutil.boot_time()
        uptime_seconds = psutil.time.time() - boot_time
        uptime_hours = round(uptime_seconds / 3600, 1)
        info["uptime"] = uptime_hours

        # System name
        info["hostname"] = platform.node()

    except Exception as e:
        info["error"] = str(e)

    return info


def generate_html(system_info: Dict[str, Any]) -> str:
    """Generate HTML output for the system widget"""
    if "error" in system_info:
        return f"""
        <div class="system-widget">
            <h2>💻 System Status</h2>
            <div style="text-align: center; color: #666;">
                ⚠️ Error: {system_info['error']}<br>
                <small>Unable to retrieve system information</small>
            </div>
        </div>
        """

    html_parts = []
    html_parts.append('<div class="system-widget">')
    html_parts.append("<h2>💻 System Status</h2>")

    # System name and uptime
    hostname = system_info.get("hostname", "Unknown")
    uptime = system_info.get("uptime", 0)
    html_parts.append(
        f'<div style="margin-bottom: 10px;"><strong>{hostname}</strong> \
• Uptime: {uptime}h</div>'
    )

    # Create metrics grid
    html_parts.append(
        '<div style="display: flex; justify-content: space-between; \
text-align: center;">'
    )

    # CPU info
    if "cpu" in system_info:
        cpu = system_info["cpu"]
        html_parts.append(
            f"""
        <div style="flex: 1;">
            <div style="font-weight: bold;">CPU</div>
            <div style="font-size: 1.1em;">{cpu['usage']:.1f}%</div>
            <div style="font-size: 0.8em;">{cpu['cores']} cores</div>
        </div>
        """
        )

    # Memory info
    if "memory" in system_info:
        memory = system_info["memory"]
        html_parts.append(
            f"""
        <div style="flex: 1; border-left: 1px solid #ccc; \
border-right: 1px solid #ccc;">
            <div style="font-weight: bold;">Memory</div>
            <div style="font-size: 1.1em;">{memory['percent']:.1f}%</div>
            <div style="font-size: 0.8em;">\
{memory['used_gb']}/{memory['total_gb']} GB</div>
        </div>
        """
        )

    # Disk info
    if "disk" in system_info:
        disk = system_info["disk"]
        html_parts.append(
            f"""
        <div style="flex: 1;">
            <div style="font-weight: bold;">Disk</div>
            <div style="font-size: 1.1em;">{disk['percent']:.1f}%</div>
            <div style="font-size: 0.8em;">{disk['used_gb']}/{disk['total_gb']} GB</div>
        </div>
        """
        )

    html_parts.append("</div>")  # Close metrics grid
    html_parts.append("</div>")  # Close widget

    return "".join(html_parts)


def main():
    """Main widget execution function"""
    try:
        # Load parameters
        params = load_parameters()

        # Get system information
        system_info = get_system_info(params)

        # Generate and output HTML
        html_output = generate_html(system_info)
        print(html_output)

    except Exception as e:
        # Error handling - output user-friendly error message
        error_html = f"""
        <div class="system-widget">
            <h2>💻 System Status</h2>
            <div style="text-align: center; color: #666;">
                ⚠️ Error: {str(e)}<br>
                <small>Check system permissions and psutil installation</small>
            </div>
        </div>
        """
        print(error_html)


if __name__ == "__main__":
    main()
