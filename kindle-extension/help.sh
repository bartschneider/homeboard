#!/bin/sh
#
# E-Paper Dashboard Help Information
# Displays help text on the Kindle screen
#

# Clear the screen
eips -c

# Display help information
eips -g "E-Paper Dashboard Help"
eips -g ""
eips -g "Setup Instructions:"
eips -g "1. Edit dashboard-launcher.sh"
eips -g "2. Set SERVER_IP to your homelab IP"
eips -g "3. Ensure server is running"
eips -g "4. Launch Dashboard from KUAL menu"
eips -g ""
eips -g "Troubleshooting:"
eips -g "- Check network connection"
eips -g "- Verify server IP and port"
eips -g "- Check dashboard.log in /mnt/us/"
eips -g ""
eips -g "Manual refresh:"
eips -g "Run /mnt/us/refresh-dashboard.sh"
eips -g ""
eips -g "Press any key to continue..."

# Wait for user input
read -n 1

exit 0