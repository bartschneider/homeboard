# E-Paper Homelab Dashboard

A lightweight, extensible dashboard system designed for e-paper displays, specifically optimized for jailbroken Amazon Kindle devices.

## Features

- **Go Backend**: Fast, concurrent widget execution with hot-reloading configuration
- **Python Widget System**: Extensible widget framework for custom data sources
- **E-Paper Optimized**: High-contrast, static design perfect for e-ink displays
- **Auto-Refresh**: Configurable refresh intervals with JavaScript-based reloading
- **Admin Panel**: Web-based configuration interface (placeholder)
- **Concurrent Processing**: All widgets execute in parallel for minimal load times

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Python 3.7+ with `psutil` and `pytz` packages
- Jailbroken Kindle with KUAL (for device deployment)

### Installation

1. **Clone and setup the project:**
```bash
git clone <repository-url>
cd homeboard
go mod tidy
```

2. **Install Python dependencies:**
```bash
pip3 install psutil pytz
```

3. **Configure the dashboard:**
Edit `config.json` to customize widgets and settings.

4. **Run the server:**
```bash
go run cmd/server/main.go
```

5. **Access the dashboard:**
- Dashboard: http://localhost:8080
- Admin Panel: http://localhost:8080/admin
- Health Check: http://localhost:8080/health

### Command Line Options

```bash
go run cmd/server/main.go [options]

Options:
  -config string    Path to configuration file (default "config.json")
  -python string    Path to Python interpreter (default "python3")
  -verbose          Enable verbose logging
```

## Project Structure

```
homeboard/
├── cmd/server/           # Main application entry point
│   └── main.go
├── internal/             # Internal packages
│   ├── config/          # Configuration management
│   ├── widgets/         # Widget execution engine
│   └── handlers/        # HTTP handlers
├── widgets/             # Python widget scripts
│   ├── clock.py         # Date/time display
│   ├── system.py        # System metrics
│   └── examples/        # Additional widget examples
├── config.json          # Main configuration file
├── go.mod               # Go module definition
└── README.md
```

## Configuration

The `config.json` file controls all dashboard settings:

```json
{
  "refresh_interval": 15,     // Auto-refresh interval in minutes
  "server_port": 8080,        // HTTP server port
  "title": "E-Paper Dashboard",
  "theme": {
    "font_family": "serif",
    "font_size": "16px",
    "background": "#ffffff",
    "foreground": "#000000"
  },
  "widgets": [
    {
      "name": "Clock",
      "script": "widgets/clock.py",
      "enabled": true,
      "timeout": 10,
      "parameters": {
        "format": "%Y-%m-%d %H:%M:%S",
        "timezone": "Local"
      }
    }
  ]
}
```

### Widget Configuration

Each widget supports:
- `name`: Display name for the widget
- `script`: Path to Python script
- `enabled`: Whether to execute this widget
- `timeout`: Maximum execution time in seconds
- `parameters`: Custom parameters passed to the script

## Creating Custom Widgets

Widgets are standalone Python scripts that:
1. Accept parameters as JSON via command line argument
2. Output HTML to stdout
3. Handle their own errors gracefully

### Widget Template

```python
#!/usr/bin/env python3
import json
import sys

def load_parameters():
    if len(sys.argv) < 2:
        return {}
    try:
        return json.loads(sys.argv[1])
    except (json.JSONDecodeError, IndexError):
        return {}

def main():
    try:
        params = load_parameters()
        
        # Your widget logic here
        data = fetch_data(params)
        
        # Generate HTML output
        html = f"""
        <div class="my-widget">
            <h2>My Widget</h2>
            <div>{data}</div>
        </div>
        """
        print(html)
        
    except Exception as e:
        # Error handling
        error_html = f"""
        <div class="my-widget">
            <h2>My Widget</h2>
            <div>⚠️ Error: {str(e)}</div>
        </div>
        """
        print(error_html)

if __name__ == "__main__":
    main()
```

## Kindle Deployment

### KUAL Extension Setup

1. **Create extension directory:**
```bash
mkdir -p /mnt/us/extensions/dashboard-launcher
```

2. **Create menu.json:**
```json
{
  "name": "E-Paper Dashboard",
  "priority": 1,
  "items": [
    {
      "name": "Launch Dashboard",
      "priority": 1,
      "action": "dashboard-launcher.sh"
    }
  ]
}
```

3. **Create launcher script:**
```bash
#!/bin/sh
# Disable screensaver
lipc-set-prop com.lab126.powerd preventScreenSaver 1
# Launch browser in kiosk mode
/usr/bin/webreader http://YOUR_SERVER_IP:8080
```

## Performance Targets

- **Dashboard Load Time**: < 500ms for 3-5 widgets
- **Widget Execution**: Concurrent processing with 30s timeout
- **Memory Usage**: < 50MB total server footprint
- **Auto-Refresh**: Configurable 15-minute default interval

## Built-in Widgets

### Clock Widget (`widgets/clock.py`)
Displays current date and time with timezone support.

**Parameters:**
- `format`: Python strftime format string
- `timezone`: Timezone name (pytz) or "Local"

### System Widget (`widgets/system.py`)
Shows system metrics including CPU, memory, and disk usage.

**Parameters:**
- `show_cpu`: Display CPU usage (default: true)
- `show_memory`: Display memory usage (default: true)  
- `show_disk`: Display disk usage (default: true)

## Development

### Running in Development Mode

```bash
# Terminal 1: Run server with verbose logging
go run cmd/server/main.go -verbose

# Terminal 2: Test widget directly
python3 widgets/clock.py '{"format": "%H:%M", "timezone": "UTC"}'

# Terminal 3: Monitor logs
tail -f /var/log/dashboard.log
```

### Testing Widgets

Each widget can be tested independently:
```bash
python3 widgets/system.py '{"show_cpu": true, "show_memory": false}'
```

### Hot-Reloading

The configuration file is reloaded on every dashboard request, enabling:
- Widget parameter changes without restart
- Theme updates in real-time
- Adding/removing widgets dynamically

## Troubleshooting

### Common Issues

1. **Widgets not appearing:**
   - Check widget script exists and is executable
   - Verify Python dependencies are installed
   - Check server logs for execution errors

2. **Kindle browser issues:**
   - Ensure KUAL is properly installed
   - Verify network connectivity to server
   - Check browser cache and cookies

3. **Performance issues:**
   - Reduce widget timeout values
   - Disable slow or problematic widgets
   - Monitor server resource usage

### Debug Mode

Enable verbose logging to troubleshoot issues:
```bash
go run cmd/server/main.go -verbose -config config.json
```

### Health Checks

Monitor server status:
```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/config
```

## Future Enhancements

- [ ] Full admin panel implementation
- [ ] Widget library with pre-built components
- [ ] Image/chart support with e-ink optimization
- [ ] Grid layout system for custom arrangements
- [ ] WebSocket support for real-time updates
- [ ] Docker containerization
- [ ] Prometheus metrics integration

## License

[License information here]

## Contributing

[Contributing guidelines here]