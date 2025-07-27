package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/bartosz/homeboard/internal/config"
	"github.com/bartosz/homeboard/internal/db"
	"github.com/bartosz/homeboard/internal/widgets"
)

// designSystemCSS contains the embedded CSS for the design system
const designSystemCSS = `/**
 * E-Paper Dashboard Design System
 * Inspired by TRMNL Weather Widget Excellence
 * Optimized for E-Paper displays with high contrast and clean typography
 */

/* ===== DESIGN TOKENS ===== */
:root {
  /* Spacing Scale */
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --spacing-xl: 32px;
  --spacing-xxl: 48px;
  
  /* Border & Radius */
  --border-radius: 6px;
  --border-radius-lg: 12px;
  --border-width: 1px;
  --border-dotted: 2px dotted;
  --border-solid: 1px solid;
  
  /* Typography Scale */
  --font-size-xs: 11px;
  --font-size-sm: 13px;
  --font-size-md: 16px;
  --font-size-lg: 20px;
  --font-size-xl: 24px;
  --font-size-xxl: 32px;
  --font-size-xxxl: 40px;
  
  /* Font Weights */
  --font-weight-normal: 400;
  --font-weight-medium: 500;
  --font-weight-semibold: 600;
  --font-weight-bold: 700;
  
  /* Line Heights */
  --line-height-tight: 1.2;
  --line-height-normal: 1.4;
  --line-height-relaxed: 1.6;
  
  /* Color Palette - E-Paper Optimized */
  --color-primary: #000000;
  --color-secondary: #333333;
  --color-tertiary: #666666;
  --color-subtle: #999999;
  --color-muted: #cccccc;
  --color-background: #ffffff;
  --color-surface: #fafafa;
  --color-border: #e5e5e5;
  --color-border-strong: #cccccc;
  
  /* Shadows - Minimal for E-Paper */
  --shadow-subtle: 0 1px 3px rgba(0, 0, 0, 0.1);
  --shadow-soft: 0 2px 8px rgba(0, 0, 0, 0.08);
  
  /* Z-Index Scale */
  --z-base: 1;
  --z-elevated: 10;
  --z-modal: 100;
  --z-overlay: 1000;
}

/* ===== RESET & BASE STYLES ===== */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html {
  font-size: 16px;
  -webkit-text-size-adjust: 100%;
}

body {
  font-family: "Times New Roman", Times, serif;
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-normal);
  line-height: var(--line-height-normal);
  color: var(--color-primary);
  background-color: var(--color-background);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* ===== LAYOUT FOUNDATION ===== */
.dashboard {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.dashboard-header {
  background: var(--color-background);
  border-bottom: var(--border-dotted) var(--color-border-strong);
  padding: var(--spacing-md) var(--spacing-lg);
  flex-shrink: 0;
}

.dashboard-content {
  flex: 1;
  padding: var(--spacing-lg);
  overflow-y: auto;
}

.dashboard-footer {
  background: var(--color-surface);
  border-top: var(--border-solid) var(--color-border);
  padding: var(--spacing-sm) var(--spacing-lg);
  flex-shrink: 0;
}

/* ===== GRID SYSTEM ===== */
.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: var(--spacing-lg);
  width: 100%;
}

.grid-1 { grid-template-columns: 1fr; }
.grid-2 { grid-template-columns: repeat(2, 1fr); }
.grid-3 { grid-template-columns: repeat(3, 1fr); }
.grid-4 { grid-template-columns: repeat(4, 1fr); }

.col-span-1 { grid-column: span 1; }
.col-span-2 { grid-column: span 2; }
.col-span-3 { grid-column: span 3; }
.col-span-4 { grid-column: span 4; }

/* ===== WIDGET CARD SYSTEM ===== */
.widget-card {
  background: var(--color-background);
  border: var(--border-dotted) var(--color-border-strong);
  border-radius: var(--border-radius);
  padding: var(--spacing-lg);
  display: flex;
  flex-direction: column;
  position: relative;
  min-height: 200px;
  transition: border-color 0.2s ease;
}

.widget-card:hover {
  border-color: var(--color-secondary);
}

.widget-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-md);
  padding-bottom: var(--spacing-sm);
  border-bottom: var(--border-solid) var(--color-border);
}

.widget-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.widget-footer {
  margin-top: var(--spacing-md);
  padding-top: var(--spacing-sm);
  border-top: var(--border-solid) var(--color-border);
}

/* ===== TYPOGRAPHY SYSTEM ===== */
.title {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-bold);
  color: var(--color-primary);
  line-height: var(--line-height-tight);
  margin: 0;
}

.title--large {
  font-size: var(--font-size-xl);
}

.title--huge {
  font-size: var(--font-size-xxl);
}

.subtitle {
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-semibold);
  color: var(--color-secondary);
  line-height: var(--line-height-tight);
  margin: 0;
}

.subtitle--small {
  font-size: var(--font-size-sm);
}

.value {
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-primary);
  line-height: var(--line-height-tight);
  margin: 0;
}

.value--small {
  font-size: var(--font-size-md);
}

.value--large {
  font-size: var(--font-size-xxl);
}

.value--huge {
  font-size: var(--font-size-xxxl);
}

.description {
  font-size: var(--font-size-sm);
  color: var(--color-secondary);
  line-height: var(--line-height-relaxed);
  margin: 0;
}

.meta {
  font-size: var(--font-size-xs);
  color: var(--color-subtle);
  font-style: italic;
  line-height: var(--line-height-normal);
  margin: 0;
}

.label {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  color: var(--color-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin: 0;
}

/* ===== COMPONENT LIBRARY ===== */

/* Metrics Grid */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: var(--spacing-md);
}

.metric-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-md);
  border: var(--border-solid) var(--color-border);
  border-radius: var(--border-radius);
  background: var(--color-surface);
}

.metric-icon {
  font-size: var(--font-size-lg);
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-background);
  border-radius: var(--border-radius);
  border: var(--border-solid) var(--color-border);
}

.metric-info {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
  flex: 1;
}

/* Weather Components */
.weather-current {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  margin-bottom: var(--spacing-md);
}

.weather-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 80px;
  height: 80px;
}

.weather-icon img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.weather-temp {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
  flex: 1;
}

.weather-divider {
  border-top: var(--border-solid) var(--color-border);
  margin: var(--spacing-md) 0;
}

.hourly-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(60px, 1fr));
  gap: var(--spacing-sm);
  margin-top: var(--spacing-md);
}

.hour-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-xs);
  text-align: center;
  padding: var(--spacing-sm);
  border-radius: var(--border-radius);
  border: var(--border-solid) var(--color-border);
}

.hour-item img {
  width: 24px;
  height: 24px;
  object-fit: contain;
}

/* Time Display */
.time-display {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: var(--spacing-lg);
}

.time-main {
  padding: var(--spacing-xl);
  border: var(--border-dotted) var(--color-border-strong);
  border-radius: var(--border-radius-lg);
  background: var(--color-surface);
}

.time-meta {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
  align-items: center;
}

/* Status Indicators */
.status-indicator {
  display: inline-flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--border-radius);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
}

.status-indicator--success {
  background: #f0f9f0;
  border: var(--border-solid) #d0f0d0;
  color: #2d5a2d;
}

.status-indicator--warning {
  background: #fff8e1;
  border: var(--border-solid) #ffe082;
  color: #b8860b;
}

.status-indicator--error {
  background: #fef2f2;
  border: var(--border-solid) #fecaca;
  color: #dc2626;
}

/* Error States */
.widget-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  gap: var(--spacing-md);
  padding: var(--spacing-xl);
  border: var(--border-dotted) var(--color-muted);
  border-radius: var(--border-radius);
  background: var(--color-surface);
}

.error-icon {
  font-size: var(--font-size-xxl);
  color: var(--color-subtle);
}

.error-message {
  font-size: var(--font-size-sm);
  color: var(--color-secondary);
  line-height: var(--line-height-relaxed);
}

.error-hint {
  font-size: var(--font-size-xs);
  color: var(--color-subtle);
  font-style: italic;
  line-height: var(--line-height-relaxed);
}

/* ===== RESPONSIVE DESIGN ===== */

/* Tablet Layout */
@media screen and (max-width: 1024px) {
  .dashboard-grid {
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: var(--spacing-md);
  }
  
  .widget-card {
    padding: var(--spacing-md);
  }
  
  .dashboard-content {
    padding: var(--spacing-md);
  }
}

/* Mobile/E-Reader Layout */
@media screen and (max-width: 768px) {
  :root {
    --font-size-xs: 10px;
    --font-size-sm: 12px;
    --font-size-md: 14px;
    --font-size-lg: 18px;
    --font-size-xl: 22px;
    --font-size-xxl: 28px;
    --font-size-xxxl: 34px;
  }
  
  .dashboard-grid {
    grid-template-columns: 1fr;
    gap: var(--spacing-sm);
  }
  
  .dashboard-content {
    padding: var(--spacing-sm);
  }
  
  .widget-card {
    padding: var(--spacing-md);
    min-height: 160px;
  }
  
  .metrics-grid {
    grid-template-columns: 1fr;
  }
  
  .hourly-grid {
    grid-template-columns: repeat(auto-fit, minmax(50px, 1fr));
  }
  
  .weather-current {
    flex-direction: column;
    text-align: center;
    gap: var(--spacing-md);
  }
}

/* Kindle/Small E-Paper */
@media screen and (max-width: 600px) {
  :root {
    --spacing-xs: 3px;
    --spacing-sm: 6px;
    --spacing-md: 12px;
    --spacing-lg: 18px;
    --spacing-xl: 24px;
    --spacing-xxl: 36px;
  }
  
  .widget-card {
    padding: var(--spacing-sm);
    min-height: 140px;
  }
  
  .dashboard-header,
  .dashboard-footer {
    padding: var(--spacing-sm) var(--spacing-md);
  }
  
  .time-main {
    padding: var(--spacing-lg);
  }
  
  .weather-icon {
    width: 60px;
    height: 60px;
  }
  
  .metric-item {
    padding: var(--spacing-sm);
  }
}

/* High DPI E-Paper Displays */
@media screen and (min-resolution: 2dppx) {
  body {
    -webkit-font-smoothing: subpixel-antialiased;
  }
  
  .widget-card {
    border-width: 1.5px;
  }
}

/* ===== UTILITY CLASSES ===== */
.text-center { text-align: center; }
.text-left { text-align: left; }
.text-right { text-align: right; }

.flex { display: flex; }
.flex-col { flex-direction: column; }
.flex-row { flex-direction: row; }
.items-center { align-items: center; }
.items-start { align-items: flex-start; }
.items-end { align-items: flex-end; }
.justify-center { justify-content: center; }
.justify-between { justify-content: space-between; }
.justify-start { justify-content: flex-start; }
.justify-end { justify-content: flex-end; }

.gap-xs { gap: var(--spacing-xs); }
.gap-sm { gap: var(--spacing-sm); }
.gap-md { gap: var(--spacing-md); }
.gap-lg { gap: var(--spacing-lg); }
.gap-xl { gap: var(--spacing-xl); }

.mb-xs { margin-bottom: var(--spacing-xs); }
.mb-sm { margin-bottom: var(--spacing-sm); }
.mb-md { margin-bottom: var(--spacing-md); }
.mb-lg { margin-bottom: var(--spacing-lg); }
.mb-xl { margin-bottom: var(--spacing-xl); }

.mt-xs { margin-top: var(--spacing-xs); }
.mt-sm { margin-top: var(--spacing-sm); }
.mt-md { margin-top: var(--spacing-md); }
.mt-lg { margin-top: var(--spacing-lg); }
.mt-xl { margin-top: var(--spacing-xl); }

.text-primary { color: var(--color-primary); }
.text-secondary { color: var(--color-secondary); }
.text-tertiary { color: var(--color-tertiary); }
.text-subtle { color: var(--color-subtle); }
.text-muted { color: var(--color-muted); }

.bg-surface { background-color: var(--color-surface); }
.bg-background { background-color: var(--color-background); }

.border { border: var(--border-solid) var(--color-border); }
.border-strong { border: var(--border-solid) var(--color-border-strong); }
.border-dotted { border: var(--border-dotted) var(--color-border-strong); }
.rounded { border-radius: var(--border-radius); }
.rounded-lg { border-radius: var(--border-radius-lg); }

.w-full { width: 100%; }
.h-full { height: 100%; }
.flex-1 { flex: 1; }
.flex-grow { flex-grow: 1; }
.flex-shrink-0 { flex-shrink: 0; }

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

/* ===== ANIMATION & INTERACTION ===== */
.transition {
  transition: all 0.2s ease;
}

.hover-lift:hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-soft);
}

.hover-border:hover {
  border-color: var(--color-secondary);
}

/* ===== PRINT OPTIMIZATION ===== */
@media print {
  .widget-card {
    border: var(--border-solid) var(--color-border-strong) !important;
    break-inside: avoid;
    page-break-inside: avoid;
  }
  
  .dashboard-grid {
    gap: var(--spacing-md);
  }
  
  .hover-lift,
  .hover-border,
  .transition {
    transform: none !important;
    transition: none !important;
    box-shadow: none !important;
  }
}`

// DashboardHandler handles dashboard requests
type DashboardHandler struct {
	configPath    string
	executor      *widgets.Executor
	template      *template.Template
	database      *db.Database
	clientRepo    *db.ClientRepository
	dashboardRepo *db.DashboardRepository
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(configPath string, executor *widgets.Executor, database *db.Database) (*DashboardHandler, error) {
	// Parse the dashboard template with safe HTML function
	tmpl, err := template.New("dashboard").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"div": func(a, b int64) float64 {
			return float64(a) / float64(b)
		},
	}).Parse(dashboardTemplate)
	if err != nil {
		return nil, err
	}

	return &DashboardHandler{
		configPath:    configPath,
		executor:      executor,
		template:      tmpl,
		database:      database,
		clientRepo:    db.NewClientRepository(database),
		dashboardRepo: db.NewDashboardRepository(database),
	}, nil
}

// ServeHTTP handles HTTP requests for the dashboard
func (h *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Get client IP address
	clientIP := h.getClientIP(r)

	// Update client last seen
	userAgent := r.Header.Get("User-Agent")
	if err := h.clientRepo.UpdateLastSeen(clientIP, userAgent); err != nil {
		log.Printf("Warning: Failed to update client last seen: %v", err)
	}

	// Get client and assigned dashboard
	client, err := h.clientRepo.GetByIP(clientIP)
	if err != nil {
		log.Printf("Error getting client: %v", err)
	}

	var dashboard *db.Dashboard
	if client != nil && client.AssignedDashboardID != nil {
		// Get assigned dashboard
		dashboard, err = h.dashboardRepo.GetByID(*client.AssignedDashboardID)
		if err != nil {
			log.Printf("Error loading assigned dashboard: %v", err)
		}
	}

	// Fall back to default dashboard if no specific assignment
	if dashboard == nil {
		dashboard, err = h.dashboardRepo.GetDefault()
		if err != nil {
			log.Printf("Error loading default dashboard: %v", err)
		}
	}

	// Fall back to legacy config system if no database dashboard found
	if dashboard == nil {
		h.serveLegacyDashboard(w, r, start)
		return
	}

	// Execute widgets from database configuration
	results := h.executeDBWidgets(dashboard.Widgets)

	// Log execution statistics
	stats := h.calculateStats(results)
	log.Printf("Dashboard rendered: %d widgets, %d successful, %d failed, total time: %v",
		stats["total_widgets"], stats["successful"], stats["failed"], stats["total_time"])

	// Load config for basic settings (theme, etc.)
	cfg, err := config.LoadConfig(h.configPath)
	if err != nil {
		log.Printf("Warning: Failed to load config: %v", err)
		cfg = config.DefaultConfig()
	}

	// Prepare template data
	data := DashboardData{
		Config:        cfg,
		WidgetResults: results,
		GeneratedAt:   time.Now(),
		RenderTime:    time.Since(start),
		Stats:         stats,
		Dashboard:     dashboard,
		Client:        client,
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

	log.Printf("Dashboard request completed in %v (client: %s, dashboard: %s)",
		time.Since(start), clientIP, dashboard.Name)
}

// DashboardData represents data passed to the dashboard template
type DashboardData struct {
	Config        *config.Config
	WidgetResults []widgets.ExecutorResult
	GeneratedAt   time.Time
	RenderTime    time.Duration
	Stats         map[string]interface{}
	Dashboard     *db.Dashboard
	Client        *db.Client
}

// dashboardTemplate is the HTML template for the dashboard
const dashboardTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Config.Title}}</title>
    <link rel="icon" href="/favicon.ico">
    <style>
` + designSystemCSS + `
    </style>
</head>
<body>
    <div class="dashboard">
        <!-- Dashboard Header -->
        <header class="dashboard-header">
            <div class="flex items-center justify-between">
                <h1 class="title title--large">{{.Config.Title}}</h1>
                <div class="flex items-center gap-md">
                    <span class="status-indicator status-indicator--success">
                        {{len .WidgetResults}} widgets
                    </span>
                    <span class="meta">{{.GeneratedAt.Format "15:04"}}</span>
                </div>
            </div>
        </header>

        <!-- Dashboard Content -->
        <main class="dashboard-content">
            <div class="dashboard-grid">
                {{range .WidgetResults}}
                <div class="widget-card transition hover-lift">
                    <div class="widget-header">
                        <h2 class="title">{{.Name}}</h2>
                        {{if .Error}}
                        <span class="status-indicator status-indicator--error">Error</span>
                        {{else}}
                        <span class="status-indicator status-indicator--success">
                            {{printf "%.0fms" (.Elapsed.Nanoseconds | div 1000000)}}
                        </span>
                        {{end}}
                    </div>
                    
                    <div class="widget-content">
                        {{if .Error}}
                        <div class="widget-error">
                            <div class="error-icon">⚠️</div>
                            <div class="error-message">{{.Error}}</div>
                            <div class="error-hint">Check widget configuration and dependencies</div>
                        </div>
                        {{else}}
                        {{.HTML | safeHTML}}
                        {{end}}
                    </div>
                </div>
                {{end}}
            </div>
        </main>

        <!-- Dashboard Footer -->
        <footer class="dashboard-footer">
            <div class="footer-info flex items-center justify-center gap-md">
                <span class="meta">Last updated: {{.GeneratedAt.Format "15:04:05"}}</span>
                <span class="meta">•</span>
                <span class="meta">Render time: {{.RenderTime.Milliseconds}}ms</span>
                <span class="meta">•</span>
                <span class="meta">{{len .WidgetResults}} widgets loaded</span>
                {{if .Stats.failed}}
                <span class="meta">•</span>
                <span class="status-indicator status-indicator--warning">
                    {{.Stats.failed}} failed
                </span>
                {{end}}
            </div>
        </footer>
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

        // Add visual feedback for refresh
        document.addEventListener('keydown', function(event) {
            if (event.key === 'r' || event.key === 'R') {
                document.body.style.opacity = '0.7';
                setTimeout(() => window.location.reload(), 100);
            }
        });
    </script>
</body>
</html>`

// getClientIP extracts client IP address from request
func (h *DashboardHandler) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for reverse proxies)
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// Get the first IP in the list
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// Fall back to RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}

// executeDBWidgets executes widgets from database configuration
func (h *DashboardHandler) executeDBWidgets(dashboardWidgets []db.DashboardWidget) []widgets.ExecutorResult {
	var results []widgets.ExecutorResult

	for _, dw := range dashboardWidgets {
		if dw.Widget == nil || !dw.Widget.Enabled {
			continue
		}

		widget := dw.Widget
		result := widgets.ExecutorResult{
			Name: widget.Name,
		}

		start := time.Now()

		// Prepare widget configuration for generic script
		config := map[string]interface{}{
			"api_url":       widget.APIURL,
			"api_headers":   widget.APIHeaders,
			"template_type": widget.TemplateType,
			"data_mapping":  widget.DataMapping,
			"timeout":       widget.Timeout,
		}

		configJSON, err := json.Marshal(config)
		if err != nil {
			result.Error = fmt.Errorf("failed to serialize widget config: %w", err)
			result.Elapsed = time.Since(start)
			results = append(results, result)
			continue
		}

		// Check if this is our enhanced weather widget
		var cmd *exec.Cmd
		if widget.Name == "🌤️ Weather" && widget.TemplateType == "weather_current" {
			// Use our enhanced weather widget
			cmd = exec.Command("python3", "widgets/weather_enhanced.py", string(configJSON))
		} else {
			// Execute generic widget script
			cmd = exec.Command("python3", "widgets/generic_api_widget.py", string(configJSON))
		}
		output, err := cmd.Output()
		result.Elapsed = time.Since(start)

		if err != nil {
			result.Error = fmt.Errorf("widget execution failed: %w", err)
		} else {
			result.HTML = string(output)
		}

		results = append(results, result)
	}

	return results
}

// serveLegacyDashboard serves dashboard using legacy config system
func (h *DashboardHandler) serveLegacyDashboard(w http.ResponseWriter, r *http.Request, start time.Time) {
	// Load configuration
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
	log.Printf("Legacy dashboard rendered: %d widgets, %d successful, %d failed, total time: %v",
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

	log.Printf("Legacy dashboard request completed in %v", time.Since(start))
}

// calculateStats calculates execution statistics
func (h *DashboardHandler) calculateStats(results []widgets.ExecutorResult) map[string]interface{} {
	stats := make(map[string]interface{})

	totalWidgets := len(results)
	successful := 0
	failed := 0
	var totalTime time.Duration

	for _, result := range results {
		totalTime += result.Elapsed
		if result.Error != nil {
			failed++
		} else {
			successful++
		}
	}

	stats["total_widgets"] = totalWidgets
	stats["successful"] = successful
	stats["failed"] = failed
	stats["total_time"] = totalTime

	return stats
}
