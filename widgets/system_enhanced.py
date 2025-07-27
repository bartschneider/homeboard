#!/usr/bin/env python3
"""
Enhanced System Status Widget for E-Paper Dashboard
Features:
- CPU, Memory, Disk monitoring
- Enhanced metrics visualization with design system
- Clean card-based layout with icons
- Graceful error handling
"""

import json
import sys
import shutil
from typing import Dict, Any

# Try to import psutil, handle gracefully if not available
try:
    import psutil

    PSUTIL_AVAILABLE = True
except ImportError:
    PSUTIL_AVAILABLE = False


def execute_widget(parameters: Dict[str, Any]) -> str:
    """Execute the enhanced system status widget"""
    try:
        # Get parameters with defaults
        show_cpu = parameters.get("show_cpu", True)
        show_memory = parameters.get("show_memory", True)
        show_disk = parameters.get("show_disk", True)
        show_network = parameters.get("show_network", False)
        show_uptime = parameters.get("show_uptime", False)

        if not PSUTIL_AVAILABLE:
            return generate_fallback_display()

        # Collect system metrics
        metrics = []

        if show_cpu:
            cpu_percent = psutil.cpu_percent(interval=1)
            cpu_count = psutil.cpu_count()
            metrics.append(
                {
                    "icon": "📊",
                    "label": "CPU Usage",
                    "value": f"{cpu_percent:.1f}%",
                    "detail": f"{cpu_count} cores",
                    "status": get_status_level(cpu_percent, 70, 90),
                }
            )

        if show_memory:
            memory = psutil.virtual_memory()
            memory_percent = memory.percent
            memory_gb = memory.total / (1024**3)
            metrics.append(
                {
                    "icon": "🧠",
                    "label": "Memory",
                    "value": f"{memory_percent:.1f}%",
                    "detail": f"{memory_gb:.1f}GB total",
                    "status": get_status_level(memory_percent, 75, 90),
                }
            )

        if show_disk:
            disk = psutil.disk_usage("/")
            disk_percent = (disk.used / disk.total) * 100
            disk_gb = disk.total / (1024**3)
            metrics.append(
                {
                    "icon": "💾",
                    "label": "Disk Space",
                    "value": f"{disk_percent:.1f}%",
                    "detail": f"{disk_gb:.1f}GB total",
                    "status": get_status_level(disk_percent, 80, 95),
                }
            )

        if show_network:
            network = psutil.net_io_counters()
            bytes_sent_mb = network.bytes_sent / (1024**2)
            bytes_recv_mb = network.bytes_recv / (1024**2)
            metrics.append(
                {
                    "icon": "🌐",
                    "label": "Network",
                    "value": f"↑{bytes_sent_mb:.1f}MB",
                    "detail": f"↓{bytes_recv_mb:.1f}MB",
                    "status": "success",
                }
            )

        if show_uptime:
            boot_time = psutil.boot_time()
            import datetime

            uptime = datetime.datetime.now() - datetime.datetime.fromtimestamp(
                boot_time
            )
            uptime_str = str(uptime).split(".")[0]  # Remove microseconds
            metrics.append(
                {
                    "icon": "⏱️",
                    "label": "Uptime",
                    "value": uptime_str,
                    "detail": "Running time",
                    "status": "success",
                }
            )

        return generate_metrics_display(metrics)

    except Exception as e:
        return generate_error_display(str(e))


def get_status_level(
    value: float, warning_threshold: float, critical_threshold: float
) -> str:
    """Determine status level based on value and thresholds"""
    if value >= critical_threshold:
        return "error"
    elif value >= warning_threshold:
        return "warning"
    else:
        return "success"


def generate_metrics_display(metrics: list) -> str:
    """Generate HTML for metrics display using design system"""
    html_parts = []

    html_parts.append('<div class="metrics-grid">')

    for metric in metrics:
        status_class = f"status-indicator--{metric['status']}"
        html_parts.append('  <div class="metric-item">')
        html_parts.append(f'    <div class="metric-icon">{metric["icon"]}</div>')
        html_parts.append('    <div class="metric-info">')
        html_parts.append(f'      <span class="label">{metric["label"]}</span>')
        html_parts.append(
            f'      <span class="value value--small">{metric["value"]}</span>'
        )
        if "detail" in metric:
            html_parts.append(f'      <span class="meta">{metric["detail"]}</span>')
        html_parts.append("    </div>")
        html_parts.append(f'    <span class="status-indicator {status_class}">●</span>')
        html_parts.append("  </div>")

    html_parts.append("</div>")

    # Add system info summary
    try:
        import platform

        system_info = platform.system()
        release_info = platform.release()
        html_parts.append('<div class="mt-md">')
        html_parts.append(
            '  <span class="description text-center w-full flex justify-center">'
        )
        html_parts.append(f"    {system_info} {release_info}")
        html_parts.append("  </span>")
        html_parts.append("</div>")
    except Exception:
        pass

    return "\n".join(html_parts)


def generate_fallback_display() -> str:
    """Generate fallback display when psutil is not available"""
    html_parts = []

    html_parts.append('<div class="widget-error">')
    html_parts.append('  <div class="error-icon">📊</div>')
    html_parts.append(
        '  <div class="error-message">System monitoring requires psutil library</div>'
    )
    html_parts.append(
        '  <div class="error-hint">Install with: pip3 install psutil</div>'
    )
    html_parts.append("</div>")

    # Show basic disk space using shutil as fallback
    try:
        total, used, free = shutil.disk_usage("/")
        total_gb = total / (1024**3)
        used_gb = used / (1024**3)
        used_percent = (used / total) * 100

        html_parts.append('<div class="mt-md">')
        html_parts.append('  <div class="metric-item">')
        html_parts.append('    <div class="metric-icon">💾</div>')
        html_parts.append('    <div class="metric-info">')
        html_parts.append('      <span class="label">Disk Space (Basic)</span>')
        html_parts.append(
            f'      <span class="value value--small">{used_percent:.1f}%</span>'
        )
        html_parts.append(
            f'      <span class="meta">{used_gb:.1f}GB / {total_gb:.1f}GB</span>'
        )
        html_parts.append("    </div>")
        html_parts.append("  </div>")
        html_parts.append("</div>")
    except Exception:
        pass

    return "\n".join(html_parts)


def generate_error_display(error_message: str) -> str:
    """Generate error display"""
    return f"""
    <div class="widget-error">
        <div class="error-icon">⚠️</div>
        <div class="error-message">System widget error: {error_message}</div>
        <div class="error-hint">Check system permissions and dependencies</div>
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
        print(generate_error_display(str(e)))
        sys.exit(1)


if __name__ == "__main__":
    main()
