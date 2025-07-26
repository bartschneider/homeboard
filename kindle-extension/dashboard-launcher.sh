#!/bin/sh
#
# E-Paper Dashboard Launcher for Kindle
# This script launches the dashboard in the Kindle's browser
#

# Configuration - EDIT THESE VARIABLES
SERVER_IP="192.168.1.100"  # Change to your homelab server IP
SERVER_PORT="8081"         # Change to match your server port

# Construct the dashboard URL
DASHBOARD_URL="http://${SERVER_IP}:${SERVER_PORT}/"

# Log the launch attempt
echo "$(date): Launching E-Paper Dashboard at ${DASHBOARD_URL}" >> /mnt/us/dashboard.log

# Disable screensaver to prevent the Kindle from going to sleep
# This keeps the dashboard always visible
lipc-set-prop com.lab126.powerd preventScreenSaver 1

# Optional: Set a longer timeout before screensaver (in milliseconds)
# lipc-set-prop com.lab126.powerd screenSaverTimeout 3600000

# Clear any existing browser cache/cookies that might interfere
rm -f /var/local/browser/cache/* 2>/dev/null
rm -f /var/local/browser/cookies.txt 2>/dev/null

# Launch the Kindle's built-in browser in frameless mode
# The browser will open in fullscreen/kiosk mode
if [ -x "/usr/bin/webreader" ]; then
    echo "$(date): Starting webreader..." >> /mnt/us/dashboard.log
    /usr/bin/webreader "${DASHBOARD_URL}" &
elif [ -x "/usr/bin/browser" ]; then
    echo "$(date): Starting browser..." >> /mnt/us/dashboard.log
    /usr/bin/browser "${DASHBOARD_URL}" &
else
    # Fallback message if browser not found
    echo "$(date): ERROR - Browser not found!" >> /mnt/us/dashboard.log
    eips -c
    eips -g "Error: Browser not found"
    eips -g "Check Kindle jailbreak status"
    sleep 3
    exit 1
fi

# Optional: Wait a moment and then check if the browser started
sleep 2

# Log successful launch
echo "$(date): Dashboard launch completed" >> /mnt/us/dashboard.log

# Optional: Add manual refresh functionality
# You can create a script that users can run to force refresh the page
cat > /mnt/us/refresh-dashboard.sh << 'EOF'
#!/bin/sh
# Manual dashboard refresh script
killall webreader 2>/dev/null
killall browser 2>/dev/null
sleep 1
/mnt/us/extensions/dashboard-launcher/dashboard-launcher.sh
EOF

chmod +x /mnt/us/refresh-dashboard.sh

exit 0