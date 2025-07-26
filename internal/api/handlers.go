package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/bartosz/homeboard/internal/db"
)

// APIHandlers contains all API handlers
type APIHandlers struct {
	clientRepo    *db.ClientRepository
	widgetRepo    *db.WidgetRepository
	dashboardRepo *db.DashboardRepository
	llmService    *LLMService
	rssService    *RSSService
}

// NewAPIHandlers creates new API handlers
func NewAPIHandlers(database *db.Database, geminiAPIKey string) *APIHandlers {
	return &APIHandlers{
		clientRepo:    db.NewClientRepository(database),
		widgetRepo:    db.NewWidgetRepository(database),
		dashboardRepo: db.NewDashboardRepository(database),
		llmService:    NewLLMService(geminiAPIKey),
		rssService:    NewRSSService(),
	}
}

// RegisterRoutes registers all API routes
func (h *APIHandlers) RegisterRoutes(router *mux.Router) {
	// API prefix
	api := router.PathPrefix("/api").Subrouter()

	// Client endpoints
	api.HandleFunc("/clients", h.GetClients).Methods("GET")
	api.HandleFunc("/clients/{id:[0-9]+}", h.AssignDashboardToClient).Methods("PUT")

	// Widget endpoints
	api.HandleFunc("/widgets", h.GetWidgets).Methods("GET")
	api.HandleFunc("/widgets", h.CreateWidget).Methods("POST")
	api.HandleFunc("/widgets/{id:[0-9]+}", h.GetWidget).Methods("GET")
	api.HandleFunc("/widgets/{id:[0-9]+}", h.UpdateWidget).Methods("PUT")
	api.HandleFunc("/widgets/{id:[0-9]+}", h.DeleteWidget).Methods("DELETE")

	// Dashboard endpoints
	api.HandleFunc("/dashboards", h.GetDashboards).Methods("GET")
	api.HandleFunc("/dashboards", h.CreateDashboard).Methods("POST")
	api.HandleFunc("/dashboards/{id:[0-9]+}", h.GetDashboard).Methods("GET")
	api.HandleFunc("/dashboards/{id:[0-9]+}", h.UpdateDashboard).Methods("PUT")
	api.HandleFunc("/dashboards/{id:[0-9]+}", h.DeleteDashboard).Methods("DELETE")
	api.HandleFunc("/dashboards/{id:[0-9]+}/widgets", h.AddWidgetToDashboard).Methods("POST")
	api.HandleFunc("/dashboards/{id:[0-9]+}/widgets/{widgetId:[0-9]+}", h.RemoveWidgetFromDashboard).Methods("DELETE")
	api.HandleFunc("/dashboards/{id:[0-9]+}/widgets/reorder", h.ReorderDashboardWidgets).Methods("PUT")

	// LLM proxy endpoint
	api.HandleFunc("/llm/analyze", h.AnalyzeWithLLM).Methods("POST")

	// Widget templates endpoint
	api.HandleFunc("/widgets/templates", h.GetWidgetTemplates).Methods("GET")

	// RSS endpoints
	api.HandleFunc("/rss/validate", h.ValidateRSSFeed).Methods("POST")
	api.HandleFunc("/rss/preview", h.PreviewRSSFeed).Methods("POST")

	// Health endpoint
	api.HandleFunc("/health", h.GetHealth).Methods("GET")

	// Add CORS middleware
	api.Use(corsMiddleware)
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Helper functions

func (h *APIHandlers) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *APIHandlers) writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *APIHandlers) getIDFromPath(r *http.Request, key string) (int, error) {
	vars := mux.Vars(r)
	idStr, ok := vars[key]
	if !ok {
		return 0, fmt.Errorf("missing %s parameter", key)
	}
	return strconv.Atoi(idStr)
}

// Client handlers

// GetClients returns all registered clients
func (h *APIHandlers) GetClients(w http.ResponseWriter, r *http.Request) {
	clients, err := h.clientRepo.GetAll()
	if err != nil {
		h.writeError(w, "Failed to fetch clients", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"clients": clients,
		"total":   len(clients),
	})
}

// AssignDashboardToClient assigns a dashboard to a client
func (h *APIHandlers) AssignDashboardToClient(w http.ResponseWriter, r *http.Request) {
	clientID, err := h.getIDFromPath(r, "id")
	if err != nil {
		h.writeError(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	var request struct {
		DashboardID int `json:"dashboard_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.clientRepo.AssignDashboard(clientID, request.DashboardID); err != nil {
		h.writeError(w, "Failed to assign dashboard", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]string{"status": "success"})
}

// Widget handlers

// GetWidgets returns all widgets
func (h *APIHandlers) GetWidgets(w http.ResponseWriter, r *http.Request) {
	widgets, err := h.widgetRepo.GetAll()
	if err != nil {
		h.writeError(w, "Failed to fetch widgets", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"widgets": widgets,
		"total":   len(widgets),
	})
}

// GetWidget returns a specific widget
func (h *APIHandlers) GetWidget(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIDFromPath(r, "id")
	if err != nil {
		h.writeError(w, "Invalid widget ID", http.StatusBadRequest)
		return
	}

	widget, err := h.widgetRepo.GetByID(id)
	if err != nil {
		h.writeError(w, "Failed to fetch widget", http.StatusInternalServerError)
		return
	}

	if widget == nil {
		h.writeError(w, "Widget not found", http.StatusNotFound)
		return
	}

	h.writeJSON(w, widget)
}

// CreateWidget creates a new widget
func (h *APIHandlers) CreateWidget(w http.ResponseWriter, r *http.Request) {
	var widget db.Widget
	if err := json.NewDecoder(r.Body).Decode(&widget); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.widgetRepo.Create(&widget); err != nil {
		if ve, ok := err.(*db.ValidationError); ok {
			h.writeError(w, ve.Message, http.StatusBadRequest)
			return
		}
		h.writeError(w, "Failed to create widget", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, widget)
}

// UpdateWidget updates an existing widget
func (h *APIHandlers) UpdateWidget(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIDFromPath(r, "id")
	if err != nil {
		h.writeError(w, "Invalid widget ID", http.StatusBadRequest)
		return
	}

	var widget db.Widget
	if err := json.NewDecoder(r.Body).Decode(&widget); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	widget.ID = id

	if err := h.widgetRepo.Update(&widget); err != nil {
		if ve, ok := err.(*db.ValidationError); ok {
			h.writeError(w, ve.Message, http.StatusBadRequest)
			return
		}
		h.writeError(w, "Failed to update widget", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, widget)
}

// DeleteWidget deletes a widget
func (h *APIHandlers) DeleteWidget(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIDFromPath(r, "id")
	if err != nil {
		h.writeError(w, "Invalid widget ID", http.StatusBadRequest)
		return
	}

	if err := h.widgetRepo.Delete(id); err != nil {
		h.writeError(w, "Failed to delete widget", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]string{"status": "deleted"})
}

// Dashboard handlers

// GetDashboards returns all dashboards
func (h *APIHandlers) GetDashboards(w http.ResponseWriter, r *http.Request) {
	dashboards, err := h.dashboardRepo.GetAll()
	if err != nil {
		h.writeError(w, "Failed to fetch dashboards", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"dashboards": dashboards,
		"total":      len(dashboards),
	})
}

// GetDashboard returns a specific dashboard with widgets
func (h *APIHandlers) GetDashboard(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIDFromPath(r, "id")
	if err != nil {
		h.writeError(w, "Invalid dashboard ID", http.StatusBadRequest)
		return
	}

	dashboard, err := h.dashboardRepo.GetByID(id)
	if err != nil {
		h.writeError(w, "Failed to fetch dashboard", http.StatusInternalServerError)
		return
	}

	if dashboard == nil {
		h.writeError(w, "Dashboard not found", http.StatusNotFound)
		return
	}

	h.writeJSON(w, dashboard)
}

// CreateDashboard creates a new dashboard
func (h *APIHandlers) CreateDashboard(w http.ResponseWriter, r *http.Request) {
	var dashboard db.Dashboard
	if err := json.NewDecoder(r.Body).Decode(&dashboard); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.dashboardRepo.Create(&dashboard); err != nil {
		if ve, ok := err.(*db.ValidationError); ok {
			h.writeError(w, ve.Message, http.StatusBadRequest)
			return
		}
		h.writeError(w, "Failed to create dashboard", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, dashboard)
}

// UpdateDashboard updates an existing dashboard
func (h *APIHandlers) UpdateDashboard(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIDFromPath(r, "id")
	if err != nil {
		h.writeError(w, "Invalid dashboard ID", http.StatusBadRequest)
		return
	}

	var dashboard db.Dashboard
	if err := json.NewDecoder(r.Body).Decode(&dashboard); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	dashboard.ID = id

	if err := h.dashboardRepo.Update(&dashboard); err != nil {
		if ve, ok := err.(*db.ValidationError); ok {
			h.writeError(w, ve.Message, http.StatusBadRequest)
			return
		}
		h.writeError(w, "Failed to update dashboard", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, dashboard)
}

// DeleteDashboard deletes a dashboard
func (h *APIHandlers) DeleteDashboard(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIDFromPath(r, "id")
	if err != nil {
		h.writeError(w, "Invalid dashboard ID", http.StatusBadRequest)
		return
	}

	if err := h.dashboardRepo.Delete(id); err != nil {
		h.writeError(w, "Failed to delete dashboard", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]string{"status": "deleted"})
}

// AddWidgetToDashboard adds a widget to a dashboard
func (h *APIHandlers) AddWidgetToDashboard(w http.ResponseWriter, r *http.Request) {
	dashboardID, err := h.getIDFromPath(r, "id")
	if err != nil {
		h.writeError(w, "Invalid dashboard ID", http.StatusBadRequest)
		return
	}

	var request struct {
		WidgetID     int `json:"widget_id"`
		DisplayOrder int `json:"display_order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.dashboardRepo.AddWidget(dashboardID, request.WidgetID, request.DisplayOrder); err != nil {
		h.writeError(w, "Failed to add widget to dashboard", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]string{"status": "added"})
}

// RemoveWidgetFromDashboard removes a widget from a dashboard
func (h *APIHandlers) RemoveWidgetFromDashboard(w http.ResponseWriter, r *http.Request) {
	dashboardID, err := h.getIDFromPath(r, "id")
	if err != nil {
		h.writeError(w, "Invalid dashboard ID", http.StatusBadRequest)
		return
	}

	widgetID, err := h.getIDFromPath(r, "widgetId")
	if err != nil {
		h.writeError(w, "Invalid widget ID", http.StatusBadRequest)
		return
	}

	if err := h.dashboardRepo.RemoveWidget(dashboardID, widgetID); err != nil {
		h.writeError(w, "Failed to remove widget from dashboard", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]string{"status": "removed"})
}

// ReorderDashboardWidgets reorders widgets in a dashboard
func (h *APIHandlers) ReorderDashboardWidgets(w http.ResponseWriter, r *http.Request) {
	dashboardID, err := h.getIDFromPath(r, "id")
	if err != nil {
		h.writeError(w, "Invalid dashboard ID", http.StatusBadRequest)
		return
	}

	var request struct {
		WidgetOrders []struct {
			WidgetID     int `json:"widget_id"`
			DisplayOrder int `json:"display_order"`
		} `json:"widget_orders"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.dashboardRepo.UpdateWidgetOrder(dashboardID, request.WidgetOrders); err != nil {
		h.writeError(w, "Failed to reorder widgets", http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, map[string]string{"status": "reordered"})
}

// AnalyzeWithLLM analyzes API data using LLM
func (h *APIHandlers) AnalyzeWithLLM(w http.ResponseWriter, r *http.Request) {
	var request db.LLMAnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.APIURL == "" || request.WidgetTemplate == "" {
		h.writeError(w, "API URL and widget template are required", http.StatusBadRequest)
		return
	}

	response, err := h.llmService.AnalyzeAPIData(request)
	if err != nil {
		h.writeError(w, fmt.Sprintf("LLM analysis failed: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, response)
}

// GetWidgetTemplates returns available widget templates
func (h *APIHandlers) GetWidgetTemplates(w http.ResponseWriter, r *http.Request) {
	templates := GetWidgetTemplates()
	h.writeJSON(w, map[string]interface{}{
		"templates": templates,
		"total":     len(templates),
	})
}

// ValidateRSSFeed validates an RSS feed URL
func (h *APIHandlers) ValidateRSSFeed(w http.ResponseWriter, r *http.Request) {
	var request struct {
		FeedURL string `json:"feed_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.FeedURL == "" {
		h.writeError(w, "Feed URL is required", http.StatusBadRequest)
		return
	}

	err := h.rssService.ValidateFeedURL(request.FeedURL)
	if err != nil {
		h.writeJSON(w, map[string]interface{}{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"valid":   true,
		"message": "RSS feed is valid",
	})
}

// PreviewRSSFeed fetches and returns a preview of an RSS feed
func (h *APIHandlers) PreviewRSSFeed(w http.ResponseWriter, r *http.Request) {
	var request struct {
		FeedURL   string        `json:"feed_url"`
		RSSConfig *db.RSSConfig `json:"rss_config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.FeedURL == "" {
		h.writeError(w, "Feed URL is required", http.StatusBadRequest)
		return
	}

	// Use default config if not provided
	if request.RSSConfig == nil {
		request.RSSConfig = &db.RSSConfig{
			MaxItems:     10,
			CacheMinutes: 5, // Short cache for preview
		}
	}

	feed, err := h.rssService.FetchFeed(request.FeedURL, request.RSSConfig)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to fetch RSS feed: %v", err), http.StatusBadRequest)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"feed":          feed,
		"preview_count": len(feed.Items),
	})
}

// GetHealth returns API health status
func (h *APIHandlers) GetHealth(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, map[string]interface{}{
		"status":    "healthy",
		"timestamp": "now",
		"version":   "1.0.0",
	})
}
