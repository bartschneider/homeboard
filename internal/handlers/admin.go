package handlers

import (
	"encoding/json"
	"net/http"
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Serve the React admin panel
	http.ServeFile(w, r, "static/admin.html")
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
