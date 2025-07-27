# Homeboard KUAL Extension

A KUAL (Kindle Unified Application Launcher) extension that connects Kindle devices to the Homeboard dashboard server, enabling e-ink dashboard displays with offline fallback capability.

## Features

- **Server Connection**: Connect to Homeboard dashboard server via WiFi
- **Device Assignment**: Automatic device registration and dashboard assignment
- **Offline Mode**: Hello World dashboard that works without network connection
- **E-ink Optimized**: Designed specifically for Kindle e-ink displays
- **Configuration UI**: Easy setup through KUAL menu interface
- **Connection Testing**: Built-in network and server connectivity tests
- **Auto-refresh**: Configurable refresh intervals optimized for e-ink

## Installation

### Prerequisites

1. Kindle device with KUAL installed
2. Homebrew/jailbreak access to Kindle filesystem
3. WiFi network access (for server connection)

### Installation Steps

1. **Copy Extension to Kindle**:
   ```bash
   # Copy the entire homeboard-kual-extension directory to:
   /mnt/us/extensions/homeboard/
   ```

2. **Set Permissions**:
   ```bash
   chmod +x /mnt/us/extensions/homeboard/bin/*.sh
   ```

3. **Launch KUAL**: The Homeboard extension will appear in the KUAL menu

## Quick Start

### Option 1: Server Connection (Recommended)

1. Open KUAL on your Kindle
2. Select "Homeboard Dashboard" → "Configure Server"
3. Enter your Homeboard server IP address and port
4. The extension will auto-generate a device ID
5. Test connection with "Connection Test"
6. Launch dashboard with "Launch Dashboard"

### Option 2: Offline Mode (No Network Required)

1. Open KUAL on your Kindle
2. Select "Homeboard Dashboard" → "Hello World (Offline)"
3. View the default dashboard with current time and device status

## Configuration

### Server Settings

- **Server URL**: IP address or hostname of Homeboard server
- **Server Port**: Port number (default: 8080)
- **Device ID**: Auto-generated unique identifier
- **Device Name**: Human-readable name for the device

### Display Settings

- **Refresh Interval**: How often to refresh the dashboard (default: 15 minutes)
- **Fullscreen Mode**: Launch browser in fullscreen (default: enabled)
- **Offline Mode**: Enable fallback to offline dashboard (default: enabled)

### Debug Settings

- **Debug Mode**: Enable detailed logging (default: disabled)
- **Log File**: Location of debug logs (/tmp/homeboard_kual.log)

## Menu Options

| Menu Item | Description |
|-----------|-------------|
| Launch Dashboard | Connect to server and display assigned dashboard |
| Configure Server | Set up server connection and device settings |
| Hello World (Offline) | Launch offline demo dashboard |
| Device Settings | Manage device configuration and preferences |
| Connection Test | Test network connectivity and server connection |

## File Structure

```
/mnt/us/extensions/homeboard/
├── menu.json                 # KUAL menu configuration
├── README.md                 # This documentation
├── bin/                      # Executable scripts
│   ├── launch_dashboard.sh   # Main dashboard launcher
│   ├── configure_server.sh   # Server configuration wizard
│   ├── hello_world.sh        # Offline dashboard launcher
│   ├── device_settings.sh    # Settings management
│   └── test_connection.sh    # Connection testing
├── config/                   # Configuration files
│   └── device.conf           # Device-specific settings
├── html/                     # HTML dashboards
│   └── hello_world.html      # Offline dashboard
└── css/                      # Stylesheets
    └── (future CSS files)
```

## Server Integration

### Device Registration

The extension automatically registers the device with the Homeboard server:

```json
{
  "device_id": "kindle_B01234567890",
  "device_name": "Kindle Dashboard",
  "device_type": "kindle",
  "capabilities": ["dashboard_display", "e_ink"]
}
```

### Dashboard Assignment

Devices can be assigned specific dashboards through the server API:

```bash
# Get device assignment
GET /api/device/{device_id}/dashboard

# Response
{
  "device_id": "kindle_B01234567890",
  "dashboard_url": "/dashboard/weather",
  "refresh_interval": 900
}
```

## Offline Dashboard

The Hello World dashboard provides:

- Current time and date display
- Device status information
- Connection status indicator
- Basic device information
- Keyboard shortcuts (R to refresh, H for KUAL home)

## Troubleshooting

### Common Issues

1. **Extension not appearing in KUAL**
   - Check file permissions: `chmod +x /mnt/us/extensions/homeboard/bin/*.sh`
   - Verify directory structure is correct
   - Restart KUAL

2. **Cannot connect to server**
   - Verify WiFi connection
   - Check server IP address and port
   - Use "Connection Test" to diagnose issues
   - Ensure server is running and accessible

3. **Dashboard not loading**
   - Check server logs for errors
   - Verify device is registered with server
   - Try offline mode to test extension functionality

4. **Browser crashes or freezes**
   - Reduce refresh interval
   - Check available memory on Kindle
   - Restart Kindle device

### Debug Mode

Enable debug mode for detailed logging:

1. Go to "Device Settings" → "Toggle Debug Mode"
2. View logs at `/tmp/homeboard_kual.log`
3. Logs include connection attempts, errors, and status updates

### Log Analysis

```bash
# View recent logs
tail -f /tmp/homeboard_kual.log

# Search for errors
grep ERROR /tmp/homeboard_kual.log

# Check connection attempts
grep "Testing connectivity" /tmp/homeboard_kual.log
```

## Keyboard Shortcuts

When dashboard is displayed:

- **R**: Refresh page
- **H**: Return to KUAL home
- **Ctrl+Alt+H**: Force close browser (emergency)

## Compatibility

### Tested Devices

- Kindle Paperwhite (multiple generations)
- Kindle Oasis
- Kindle Scribe
- Basic Kindle (newer models)

### Requirements

- KUAL 2.7 or later
- Kindle firmware with browser support
- At least 50MB free storage space
- WiFi capability (for server connection)

## Performance Optimization

### E-ink Considerations

- Refresh intervals optimized for e-ink displays
- Minimal animations and transitions
- High contrast design
- Large, readable fonts

### Memory Management

- Automatic browser restart on memory issues
- Configurable refresh intervals
- Cleanup of temporary files

## Security

- No sensitive data stored in configuration
- Device ID generation uses Kindle serial number
- Local storage only for device settings
- No remote code execution

## Support

For issues and questions:

1. Check the troubleshooting section above
2. Enable debug mode and check logs
3. Test with offline mode to isolate network issues
4. Verify Homeboard server functionality independently

## Version History

- **v1.0.0**: Initial release with server connection and offline mode
  - Basic dashboard display
  - Device registration
  - Offline Hello World dashboard
  - Configuration wizard
  - Connection testing

## License

This extension is provided as-is for use with the Homeboard dashboard system. Use at your own risk on jailbroken Kindle devices.