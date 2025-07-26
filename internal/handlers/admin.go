package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bartosz/homeboard/internal/config"
)

// AdminHandler handles admin panel requests
type AdminHandler struct {
	configPath string
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(configPath string) *AdminHandler {
	return &AdminHandler{
		configPath: configPath,
	}
}

// ServeHTTP handles HTTP requests for the admin panel
func (h *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetAdmin(w, r)
	case http.MethodPost:
		h.handlePostAdmin(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetAdmin serves the admin interface
func (h *AdminHandler) handleGetAdmin(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	// Simple admin interface template
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Admin Panel - %s</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            margin-bottom: 30px;
        }
        .section {
            margin-bottom: 30px;
            padding: 20px;
            border: 1px solid #ddd;
            border-radius: 4px;
        }
        .section h2 {
            margin-top: 0;
            color: #555;
        }
        pre {
            background: #f8f8f8;
            padding: 15px;
            border-radius: 4px;
            overflow-x: auto;
            white-space: pre-wrap;
        }
        .status {
            padding: 10px;
            border-radius: 4px;
            margin: 10px 0;
        }
        .status.info {
            background-color: #d1ecf1;
            border-color: #bee5eb;
            color: #0c5460;
        }
        button {
            background-color: #007bff;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 4px;
            cursor: pointer;
            margin-right: 10px;
        }
        button:hover {
            background-color: #0056b3;
        }
        .widget-list {
            list-style: none;
            padding: 0;
        }
        .widget-item {
            padding: 10px;
            margin: 5px 0;
            background: #f8f9fa;
            border-radius: 4px;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .widget-enabled {
            background: #d4edda;
        }
        .widget-disabled {
            background: #f8d7da;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>E-Paper Dashboard Admin Panel</h1>
        
        <div class="status info">
            <strong>Status:</strong> Dashboard is running<br>
            <strong>Server Port:</strong> %d<br>
            <strong>Refresh Interval:</strong> %d minutes<br>
            <strong>Last Loaded:</strong> %s
        </div>

        <div class="section">
            <h2>Current Configuration</h2>
            <pre>%s</pre>
        </div>

        <div class="section">
            <h2>Widgets (%d total)</h2>
            <ul class="widget-list">
                %s
            </ul>
        </div>

        <div class="section">
            <h2>Quick Actions</h2>
            <button onclick="window.open('/', '_blank')">View Dashboard</button>
            <button onclick="location.reload()">Refresh Admin</button>
        </div>

        <div class="section">
            <h2>Configuration Management</h2>
            <p><strong>Note:</strong> This is a placeholder admin interface. 
            Currently, please edit <code>config.json</code> directly to modify settings.</p>
            <p><strong>Future features:</strong></p>
            <ul>
                <li>Widget configuration editor</li>
                <li>Theme customization</li>
                <li>Real-time widget testing</li>
                <li>Performance monitoring</li>
            </ul>
        </div>
    </div>

    <script>
        console.log('Admin panel loaded');
        
        // Auto-refresh admin panel every 30 seconds
        setTimeout(function() {
            location.reload();
        }, 30000);
    </script>
</body>
</html>`,
		cfg.Title,
		cfg.ServerPort,
		cfg.RefreshInterval,
		time.Now().Format("2006-01-02 15:04:05"),
		h.formatConfigJSON(cfg),
		len(cfg.Widgets),
		h.formatWidgetList(cfg.Widgets),
	)

	w.Write([]byte(html))
}

// handlePostAdmin handles configuration updates (placeholder)
func (h *AdminHandler) handlePostAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]string{
		"status":  "not_implemented",
		"message": "Configuration updates via admin panel are not yet implemented. Please edit config.json directly.",
	}
	
	json.NewEncoder(w).Encode(response)
}

// formatConfigJSON formats the configuration as pretty JSON
func (h *AdminHandler) formatConfigJSON(cfg *config.Config) string {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting config: %v", err)
	}
	return string(data)
}

// formatWidgetList formats the widget list as HTML
func (h *AdminHandler) formatWidgetList(widgets []config.Widget) string {
	if len(widgets) == 0 {
		return "<li>No widgets configured</li>"
	}

	html := ""
	for _, widget := range widgets {
		status := "widget-disabled"
		statusText := "Disabled"
		if widget.Enabled {
			status = "widget-enabled"
			statusText = "Enabled"
		}

		html += fmt.Sprintf(`
			<li class="widget-item %s">
				<div>
					<strong>%s</strong><br>
					<small>Script: %s | Timeout: %ds</small>
				</div>
				<div>%s</div>
			</li>`,
			status, widget.Name, widget.Script, widget.Timeout, statusText)
	}

	return html
}