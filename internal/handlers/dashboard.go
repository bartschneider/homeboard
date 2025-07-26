package handlers

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/bartosz/homeboard/internal/config"
	"github.com/bartosz/homeboard/internal/widgets"
)

// DashboardHandler handles dashboard requests
type DashboardHandler struct {
	configPath string
	executor   *widgets.Executor
	template   *template.Template
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(configPath string, executor *widgets.Executor) (*DashboardHandler, error) {
	// Parse the dashboard template
	tmpl, err := template.New("dashboard").Parse(dashboardTemplate)
	if err != nil {
		return nil, err
	}

	return &DashboardHandler{
		configPath: configPath,
		executor:   executor,
		template:   tmpl,
	}, nil
}

// ServeHTTP handles HTTP requests for the dashboard
func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Load configuration on each request for hot-reloading
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		http.Error(w, "Configuration error", http.StatusInternalServerError)
		return
	}

	// Get enabled widgets
	enabledWidgets := cfg.GetEnabledWidgets()

	// Execute all widgets concurrently
	results := h.executor.ExecuteAll(enabledWidgets)

	// Log execution statistics
	stats := h.executor.GetExecutionStats(results)
	log.Printf("Dashboard rendered: %d widgets, %d successful, %d failed, total time: %v",
		stats["total_widgets"], stats["successful"], stats["failed"], stats["total_time"])

	// Prepare template data
	data := DashboardData{
		Config:        cfg,
		WidgetResults: results,
		GeneratedAt:   time.Now(),
		RenderTime:    time.Since(start),
		Stats:         stats,
	}

	// Set content type and render template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	if err := h.template.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	log.Printf("Dashboard request completed in %v", time.Since(start))
}

// DashboardData represents data passed to the dashboard template
type DashboardData struct {
	Config        *config.Config
	WidgetResults []widgets.ExecutorResult
	GeneratedAt   time.Time
	RenderTime    time.Duration
	Stats         map[string]interface{}
}

// dashboardTemplate is the HTML template for the dashboard
const dashboardTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Config.Title}}</title>
    <style>
        /* E-Paper optimized styles */
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: {{.Config.Theme.FontFamily}}, serif;
            font-size: {{.Config.Theme.FontSize}};
            background-color: {{.Config.Theme.Background}};
            color: {{.Config.Theme.Foreground}};
            line-height: 1.4;
            height: 100vh;
            overflow: hidden;
        }

        .dashboard {
            display: flex;
            flex-direction: column;
            height: 100vh;
            width: 100vw;
        }

        .header {
            text-align: center;
            padding: 10px 0;
            border-bottom: 2px solid {{.Config.Theme.Foreground}};
            flex-shrink: 0;
        }

        .header h1 {
            font-size: 1.2em;
            font-weight: bold;
        }

        .widgets-container {
            flex: 1;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }

        .widget {
            flex: 1;
            padding: 15px;
            border-bottom: 1px solid {{.Config.Theme.Foreground}};
            display: flex;
            flex-direction: column;
            justify-content: center;
        }

        .widget:last-child {
            border-bottom: none;
        }

        .widget h2 {
            font-size: 1.1em;
            margin-bottom: 8px;
            font-weight: bold;
        }

        .widget-content {
            flex: 1;
            display: flex;
            flex-direction: column;
            justify-content: center;
        }

        .widget-error {
            text-align: center;
        }

        .widget-error h3 {
            font-size: 1.1em;
            margin-bottom: 8px;
        }

        .error-message {
            font-size: 0.9em;
            margin-bottom: 5px;
        }

        .error-hint {
            font-size: 0.8em;
            font-style: italic;
        }

        .footer {
            text-align: center;
            padding: 5px 0;
            border-top: 1px solid {{.Config.Theme.Foreground}};
            font-size: 0.8em;
            flex-shrink: 0;
        }

        /* Responsive adjustments for different Kindle models */
        @media screen and (max-width: 800px) {
            body {
                font-size: 14px;
            }
            .widget {
                padding: 10px;
            }
        }

        @media screen and (max-height: 600px) {
            .header {
                padding: 5px 0;
            }
            .widget {
                padding: 8px;
            }
            .footer {
                padding: 3px 0;
            }
        }
    </style>
</head>
<body>
    <div class="dashboard">
        <div class="header">
            <h1>{{.Config.Title}}</h1>
        </div>

        <div class="widgets-container">
            {{range .WidgetResults}}
            <div class="widget">
                <div class="widget-content">
                    {{.HTML}}
                </div>
            </div>
            {{end}}
        </div>

        <div class="footer">
            Last updated: {{.GeneratedAt.Format "15:04:05"}} | Render time: {{printf "%.0fms" .RenderTime.Seconds | printf "%.0f" (.RenderTime.Seconds | printf "%.3f" | printf "%.0f")}}ms
        </div>
    </div>

    <script>
        // Auto-refresh functionality
        function autoRefresh() {
            const refreshInterval = {{.Config.RefreshInterval}} * 60 * 1000; // Convert minutes to milliseconds
            
            setTimeout(function() {
                console.log('Auto-refreshing dashboard...');
                window.location.reload();
            }, refreshInterval);
        }

        // Start auto-refresh when page loads
        document.addEventListener('DOMContentLoaded', function() {
            autoRefresh();
            console.log('Dashboard loaded. Auto-refresh set to {{.Config.RefreshInterval}} minutes.');
        });

        // Manual refresh on key press (for debugging)
        document.addEventListener('keydown', function(event) {
            if (event.key === 'r' || event.key === 'R') {
                window.location.reload();
            }
        });
    </script>
</body>
</html>`