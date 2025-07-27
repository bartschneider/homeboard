#!/bin/bash
# Homeboard KUAL Extension Installation Script
# This script helps install the extension on a Kindle device

EXTENSION_NAME="homeboard"
KUAL_EXTENSIONS_DIR="/mnt/us/extensions"
EXTENSION_DIR="$KUAL_EXTENSIONS_DIR/$EXTENSION_NAME"
SOURCE_DIR="$(dirname "$0")"

echo "Homeboard KUAL Extension Installer"
echo "=================================="
echo

# Check if we're running on a Kindle
if [ ! -d "/mnt/us" ]; then
    echo "Warning: This doesn't appear to be a Kindle device."
    echo "The /mnt/us directory was not found."
    echo
    read -p "Continue anyway? (y/N): " continue_install
    if [ "$continue_install" != "y" ] && [ "$continue_install" != "Y" ]; then
        echo "Installation cancelled."
        exit 1
    fi
fi

# Check if KUAL is installed
if [ ! -d "$KUAL_EXTENSIONS_DIR" ]; then
    echo "Error: KUAL extensions directory not found: $KUAL_EXTENSIONS_DIR"
    echo "Please install KUAL first before installing this extension."
    exit 1
fi

echo "Installing Homeboard KUAL Extension..."
echo "Source: $SOURCE_DIR"
echo "Target: $EXTENSION_DIR"
echo

# Create extension directory
if [ -d "$EXTENSION_DIR" ]; then
    echo "Extension directory already exists."
    read -p "Remove existing installation? (y/N): " remove_existing
    if [ "$remove_existing" = "y" ] || [ "$remove_existing" = "Y" ]; then
        echo "Removing existing installation..."
        rm -rf "$EXTENSION_DIR"
    else
        echo "Installation cancelled."
        exit 1
    fi
fi

echo "Creating extension directory..."
mkdir -p "$EXTENSION_DIR"

# Copy files
echo "Copying extension files..."
cp -r "$SOURCE_DIR"/* "$EXTENSION_DIR"/

# Set permissions
echo "Setting file permissions..."
chmod +x "$EXTENSION_DIR"/bin/*.sh
chmod 644 "$EXTENSION_DIR"/menu.json
chmod 644 "$EXTENSION_DIR"/html/*.html
chmod 644 "$EXTENSION_DIR"/config/device.conf

# Create log directory if it doesn't exist
mkdir -p /tmp

# Verify installation
if [ -f "$EXTENSION_DIR/menu.json" ] && [ -x "$EXTENSION_DIR/bin/launch_dashboard.sh" ]; then
    echo
    echo "✓ Installation completed successfully!"
    echo
    echo "The Homeboard extension has been installed to:"
    echo "  $EXTENSION_DIR"
    echo
    echo "Next steps:"
    echo "1. Launch KUAL on your Kindle"
    echo "2. Look for 'Homeboard Dashboard' in the menu"
    echo "3. Start with 'Configure Server' to set up your connection"
    echo "4. Or try 'Hello World (Offline)' to test the extension"
    echo
    echo "For troubleshooting:"
    echo "- Check logs at: /tmp/homeboard_kual.log"
    echo "- Run 'Connection Test' to diagnose network issues"
    echo "- Use 'Device Settings' to manage configuration"
    echo
else
    echo
    echo "✗ Installation failed!"
    echo "Some files may not have been copied correctly."
    echo "Please check the source directory and try again."
    exit 1
fi