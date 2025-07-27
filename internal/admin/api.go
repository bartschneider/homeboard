package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/bartosz/homeboard/internal/config"
	"github.com/bartosz/homeboard/internal/widgets"
)

// AdminAPI handles all admin panel API requests
type AdminAPI struct {
	configPath string
	executor   *widgets.Executor
	validator  *ConfigValidator
	backup     *BackupManager
	websocket  *WebSocketManager
	metrics    *MetricsCollector
	testRunner *WidgetTestRunner
}

// NewAdminAPI creates a new admin API handler
func NewAdminAPI(configPath string, executor *widgets.Executor) *AdminAPI {
	return &AdminAPI{
		configPath: configPath,
		executor:   executor,
		validator:  NewConfigValidator(),
		backup:     NewBackupManager(configPath),
		websocket:  NewWebSocketManager(),
		metrics:    NewMetricsCollector(),
		testRunner: NewWidgetTestRunner(executor),
	}
}

// SetupRoutes configures all admin API routes
func (api *AdminAPI) SetupRoutes(router *mux.Router) {
	// Create admin API subrouter
	adminAPI := router.PathPrefix("/api/admin").Subrouter()

	// Configuration management
	adminAPI.HandleFunc("/config", api.handleConfig).Methods("GET", "PUT")
	adminAPI.HandleFunc("/config/validate", api.handleConfigValidate).Methods("POST")
	adminAPI.HandleFunc("/config/backup", api.handleConfigBackup).Methods("POST")
	adminAPI.HandleFunc("/config/restore", api.handleConfigRestore).Methods("POST")
	adminAPI.HandleFunc("/config/backups", api.handleConfigBackups).Methods("GET")

	// Widget management
	adminAPI.HandleFunc("/widgets", api.handleWidgets).Methods("GET", "POST")
	adminAPI.HandleFunc("/widgets/{id}", api.handleWidget).Methods("GET", "PUT", "DELETE")
	adminAPI.HandleFunc("/widgets/{id}/test", api.handleWidgetTest).Methods("POST")
	adminAPI.HandleFunc("/widgets/{id}/toggle", api.handleWidgetToggle).Methods("POST")

	// System monitoring
	adminAPI.HandleFunc("/status", api.handleSystemStatus).Methods("GET")
	adminAPI.HandleFunc("/metrics", api.handleMetrics).Methods("GET")
	adminAPI.HandleFunc("/logs", api.handleLogs).Methods("GET")

	// WebSocket endpoint
	adminAPI.HandleFunc("/ws", api.handleWebSocket)

	// Add CORS middleware
	adminAPI.Use(api.corsMiddleware)
}

// Configuration Management Handlers

func (api *AdminAPI) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		api.getConfig(w, r)
	case "PUT":
		api.updateConfig(w, r)
	}
}

func (api *AdminAPI) getConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig(api.configPath)
	if err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to load configuration", err)
		return
	}

	response := ConfigResponse{
		Config:    cfg,
		Timestamp: time.Now(),
	}

	api.writeJSONResponse(w, http.StatusOK, response)
}

func (api *AdminAPI) updateConfig(w http.ResponseWriter, r *http.Request) {
	var configRequest ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&configRequest); err != nil {
		api.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err)
		return
	}

	// Validate configuration
	validation := api.validator.ValidateConfig(&configRequest)
	if !validation.Valid {
		response := ConfigResponse{
			Validation: validation,
			Timestamp:  time.Now(),
		}
		api.writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	// Create backup before updating
	if _, err := api.backup.CreateBackup(); err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create backup", err)
		return
	}

	// Convert request to config and save
	newConfig := api.requestToConfig(&configRequest)
	if err := config.SaveConfig(newConfig, api.configPath); err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to save configuration", err)
		return
	}

	// Broadcast configuration change
	api.websocket.BroadcastConfigUpdate(newConfig)

	response := ConfigResponse{
		Config:     newConfig,
		Validation: validation,
		Timestamp:  time.Now(),
	}

	api.writeJSONResponse(w, http.StatusOK, response)
}

func (api *AdminAPI) handleConfigValidate(w http.ResponseWriter, r *http.Request) {
	var configRequest ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&configRequest); err != nil {
		api.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err)
		return
	}

	validation := api.validator.ValidateConfig(&configRequest)
	api.writeJSONResponse(w, http.StatusOK, validation)
}

func (api *AdminAPI) handleConfigBackup(w http.ResponseWriter, r *http.Request) {
	backupID, err := api.backup.CreateBackup()
	if err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create backup", err)
		return
	}

	response := map[string]interface{}{
		"backup_id": backupID,
		"timestamp": time.Now(),
		"message":   "Backup created successfully",
	}

	api.writeJSONResponse(w, http.StatusOK, response)
}

func (api *AdminAPI) handleConfigRestore(w http.ResponseWriter, r *http.Request) {
	var request struct {
		BackupID string `json:"backup_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		api.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err)
		return
	}

	if err := api.backup.RestoreBackup(request.BackupID); err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to restore backup", err)
		return
	}

	// Reload configuration and broadcast update
	cfg, err := config.LoadConfig(api.configPath)
	if err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to reload configuration", err)
		return
	}

	api.websocket.BroadcastConfigUpdate(cfg)

	response := map[string]interface{}{
		"message":   "Configuration restored successfully",
		"timestamp": time.Now(),
	}

	api.writeJSONResponse(w, http.StatusOK, response)
}

func (api *AdminAPI) handleConfigBackups(w http.ResponseWriter, r *http.Request) {
	backups, err := api.backup.ListBackups()
	if err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list backups", err)
		return
	}

	api.writeJSONResponse(w, http.StatusOK, backups)
}

// Widget Management Handlers

func (api *AdminAPI) handleWidgets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		api.getWidgets(w, r)
	case "POST":
		api.createWidget(w, r)
	}
}

func (api *AdminAPI) getWidgets(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig(api.configPath)
	if err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to load configuration", err)
		return
	}

	// Enhance widgets with runtime status
	widgetStatuses := make([]WidgetStatus, len(cfg.Widgets))
	for i, widget := range cfg.Widgets {
		status := api.metrics.GetWidgetStatus(widget.Name)
		widgetStatuses[i] = WidgetStatus{
			Widget:        widget,
			Status:        status.Status,
			LastExecution: status.LastExecution,
			ExecutionTime: status.ExecutionTime,
			ErrorMessage:  status.Error,
		}
	}

	api.writeJSONResponse(w, http.StatusOK, widgetStatuses)
}

func (api *AdminAPI) createWidget(w http.ResponseWriter, r *http.Request) {
	var widgetRequest WidgetRequest
	if err := json.NewDecoder(r.Body).Decode(&widgetRequest); err != nil {
		api.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err)
		return
	}

	// Validate widget
	if err := api.validator.ValidateWidget(&widgetRequest); err != nil {
		api.writeErrorResponse(w, http.StatusBadRequest, "Widget validation failed", err)
		return
	}

	// Load current configuration
	cfg, err := config.LoadConfig(api.configPath)
	if err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to load configuration", err)
		return
	}

	// Add new widget
	newWidget := config.Widget{
		Name:       widgetRequest.Name,
		Script:     widgetRequest.Script,
		Parameters: widgetRequest.Parameters,
		Timeout:    widgetRequest.Timeout,
		Enabled:    widgetRequest.Enabled,
	}

	cfg.Widgets = append(cfg.Widgets, newWidget)

	// Save configuration
	if err := config.SaveConfig(cfg, api.configPath); err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to save configuration", err)
		return
	}

	// Broadcast update
	api.websocket.BroadcastConfigUpdate(cfg)

	api.writeJSONResponse(w, http.StatusCreated, newWidget)
}

func (api *AdminAPI) handleWidget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	widgetID := vars["id"]

	switch r.Method {
	case "GET":
		api.getWidget(w, r, widgetID)
	case "PUT":
		api.updateWidget(w, r, widgetID)
	case "DELETE":
		api.deleteWidget(w, r, widgetID)
	}
}

func (api *AdminAPI) getWidget(w http.ResponseWriter, r *http.Request, widgetID string) {
	cfg, err := config.LoadConfig(api.configPath)
	if err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to load configuration", err)
		return
	}

	widget, index := api.findWidget(cfg.Widgets, widgetID)
	if index == -1 {
		api.writeErrorResponse(w, http.StatusNotFound, "Widget not found", nil)
		return
	}

	// Get widget status
	status := api.metrics.GetWidgetStatus(widget.Name)
	widgetStatus := WidgetStatus{
		Widget:        *widget,
		Status:        status.Status,
		LastExecution: status.LastExecution,
		ExecutionTime: status.ExecutionTime,
		ErrorMessage:  status.Error,
	}

	api.writeJSONResponse(w, http.StatusOK, widgetStatus)
}

func (api *AdminAPI) updateWidget(w http.ResponseWriter, r *http.Request, widgetID string) {
	var widgetRequest WidgetRequest
	if err := json.NewDecoder(r.Body).Decode(&widgetRequest); err != nil {
		api.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err)
		return
	}

	// Validate widget
	if err := api.validator.ValidateWidget(&widgetRequest); err != nil {
		api.writeErrorResponse(w, http.StatusBadRequest, "Widget validation failed", err)
		return
	}

	cfg, err := config.LoadConfig(api.configPath)
	if err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to load configuration", err)
		return
	}

	_, index := api.findWidget(cfg.Widgets, widgetID)
	if index == -1 {
		api.writeErrorResponse(w, http.StatusNotFound, "Widget not found", nil)
		return
	}

	// Update widget
	cfg.Widgets[index] = config.Widget{
		Name:       widgetRequest.Name,
		Script:     widgetRequest.Script,
		Parameters: widgetRequest.Parameters,
		Timeout:    widgetRequest.Timeout,
		Enabled:    widgetRequest.Enabled,
	}

	// Save configuration
	if err := config.SaveConfig(cfg, api.configPath); err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to save configuration", err)
		return
	}

	// Broadcast update
	api.websocket.BroadcastConfigUpdate(cfg)

	api.writeJSONResponse(w, http.StatusOK, cfg.Widgets[index])
}

func (api *AdminAPI) deleteWidget(w http.ResponseWriter, r *http.Request, widgetID string) {
	cfg, err := config.LoadConfig(api.configPath)
	if err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to load configuration", err)
		return
	}

	_, index := api.findWidget(cfg.Widgets, widgetID)
	if index == -1 {
		api.writeErrorResponse(w, http.StatusNotFound, "Widget not found", nil)
		return
	}

	// Remove widget
	cfg.Widgets = append(cfg.Widgets[:index], cfg.Widgets[index+1:]...)

	// Save configuration
	if err := config.SaveConfig(cfg, api.configPath); err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to save configuration", err)
		return
	}

	// Broadcast update
	api.websocket.BroadcastConfigUpdate(cfg)

	response := map[string]interface{}{
		"message":   "Widget deleted successfully",
		"timestamp": time.Now(),
	}

	api.writeJSONResponse(w, http.StatusOK, response)
}

func (api *AdminAPI) handleWidgetTest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	widgetID := vars["id"]

	var testRequest WidgetTestRequest
	if err := json.NewDecoder(r.Body).Decode(&testRequest); err != nil {
		api.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err)
		return
	}

	cfg, err := config.LoadConfig(api.configPath)
	if err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to load configuration", err)
		return
	}

	widget, _ := api.findWidget(cfg.Widgets, widgetID)
	if widget == nil {
		api.writeErrorResponse(w, http.StatusNotFound, "Widget not found", nil)
		return
	}

	// Test widget with provided parameters
	testResult := api.testRunner.TestWidget(*widget, testRequest.Parameters, testRequest.Timeout)

	api.writeJSONResponse(w, http.StatusOK, testResult)
}

func (api *AdminAPI) handleWidgetToggle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	widgetID := vars["id"]

	cfg, err := config.LoadConfig(api.configPath)
	if err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to load configuration", err)
		return
	}

	_, index := api.findWidget(cfg.Widgets, widgetID)
	if index == -1 {
		api.writeErrorResponse(w, http.StatusNotFound, "Widget not found", nil)
		return
	}

	// Toggle enabled state
	cfg.Widgets[index].Enabled = !cfg.Widgets[index].Enabled

	// Save configuration
	if err := config.SaveConfig(cfg, api.configPath); err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to save configuration", err)
		return
	}

	// Broadcast update
	api.websocket.BroadcastConfigUpdate(cfg)

	response := map[string]interface{}{
		"enabled":   cfg.Widgets[index].Enabled,
		"message":   fmt.Sprintf("Widget %s", map[bool]string{true: "enabled", false: "disabled"}[cfg.Widgets[index].Enabled]),
		"timestamp": time.Now(),
	}

	api.writeJSONResponse(w, http.StatusOK, response)
}

// System Monitoring Handlers

func (api *AdminAPI) handleSystemStatus(w http.ResponseWriter, r *http.Request) {
	status := api.metrics.GetSystemStatus()
	api.writeJSONResponse(w, http.StatusOK, status)
}

func (api *AdminAPI) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := api.metrics.GetAllMetrics()
	api.writeJSONResponse(w, http.StatusOK, metrics)
}

func (api *AdminAPI) handleLogs(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	level := r.URL.Query().Get("level")
	since := r.URL.Query().Get("since")

	logs, err := api.metrics.GetLogs(limit, level, since)
	if err != nil {
		api.writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve logs", err)
		return
	}

	api.writeJSONResponse(w, http.StatusOK, logs)
}

// WebSocket Handler
func (api *AdminAPI) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	api.websocket.HandleConnection(w, r)
}

// Helper Methods

func (api *AdminAPI) findWidget(widgets []config.Widget, identifier string) (*config.Widget, int) {
	for i, widget := range widgets {
		if widget.Name == identifier || fmt.Sprintf("%d", i) == identifier {
			return &widget, i
		}
	}
	return nil, -1
}

func (api *AdminAPI) requestToConfig(req *ConfigRequest) *config.Config {
	return &config.Config{
		RefreshInterval: req.RefreshInterval,
		ServerPort:      req.ServerPort,
		Title:           req.Title,
		Theme:           req.Theme,
		Widgets:         req.Widgets,
	}
}

func (api *AdminAPI) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (api *AdminAPI) writeErrorResponse(w http.ResponseWriter, status int, message string, err error) {
	response := ErrorResponse{
		Error:     message,
		Timestamp: time.Now(),
	}
	if err != nil {
		response.Details = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func (api *AdminAPI) corsMiddleware(next http.Handler) http.Handler {
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
