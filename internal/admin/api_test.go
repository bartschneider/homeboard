package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/bartosz/homeboard/internal/config"
	"github.com/bartosz/homeboard/internal/widgets"
)

// Test setup helpers

func setupTestAPI(t *testing.T) (*AdminAPI, string, func()) {
	// Create temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	// Create test configuration
	testConfig := &config.Config{
		RefreshInterval: 15,
		ServerPort:      8081,
		Title:           "Test Dashboard",
		Theme: config.Theme{
			FontFamily: "serif",
			FontSize:   "16px",
			Background: "#ffffff",
			Foreground: "#000000",
		},
		Widgets: []config.Widget{
			{
				Name:    "Test Widget",
				Script:  "test.py",
				Enabled: true,
				Timeout: 10,
				Parameters: map[string]interface{}{
					"test_param": "value",
				},
			},
		},
	}

	// Save test configuration
	if err := config.SaveConfig(testConfig, configPath); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Create executor
	executor := widgets.NewExecutor("python3", 30*time.Second)

	// Create API
	api := NewAdminAPI(configPath, executor)

	// Cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return api, configPath, cleanup
}

func createTestRequest(method, url string, body interface{}) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, url, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

// Configuration API Tests

func TestGetConfig(t *testing.T) {
	api, _, cleanup := setupTestAPI(t)
	defer cleanup()

	req := createTestRequest("GET", "/api/admin/config", nil)
	w := httptest.NewRecorder()

	api.getConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response ConfigResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Config.Title != "Test Dashboard" {
		t.Errorf("Expected title 'Test Dashboard', got '%s'", response.Config.Title)
	}
}

func TestUpdateConfig(t *testing.T) {
	api, _, cleanup := setupTestAPI(t)
	defer cleanup()

	updateRequest := ConfigRequest{
		RefreshInterval: 20,
		ServerPort:      8082,
		Title:           "Updated Dashboard",
		Theme: config.Theme{
			FontFamily: "sans-serif",
			FontSize:   "18px",
			Background: "#ffffff",
			Foreground: "#000000",
		},
		Widgets: []config.Widget{
			{
				Name:    "Updated Widget",
				Script:  "updated.py",
				Enabled: true,
				Timeout: 15,
				Parameters: map[string]interface{}{
					"updated_param": "new_value",
				},
			},
		},
	}

	req := createTestRequest("PUT", "/api/admin/config", updateRequest)
	w := httptest.NewRecorder()

	api.updateConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response ConfigResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Config.Title != "Updated Dashboard" {
		t.Errorf("Expected title 'Updated Dashboard', got '%s'", response.Config.Title)
	}

	if response.Config.RefreshInterval != 20 {
		t.Errorf("Expected refresh interval 20, got %d", response.Config.RefreshInterval)
	}
}

func TestUpdateConfigValidation(t *testing.T) {
	api, _, cleanup := setupTestAPI(t)
	defer cleanup()

	// Test invalid configuration
	invalidRequest := ConfigRequest{
		RefreshInterval: -1,  // Invalid
		ServerPort:      100, // Invalid (too low)
		Title:           "",  // Invalid (empty)
		Theme: config.Theme{
			FontFamily: "",
			FontSize:   "",
			Background: "invalid-color",
			Foreground: "invalid-color",
		},
		Widgets: []config.Widget{},
	}

	req := createTestRequest("PUT", "/api/admin/config", invalidRequest)
	w := httptest.NewRecorder()

	api.updateConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response ConfigResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Validation.Valid {
		t.Error("Expected validation to fail, but it passed")
	}

	if len(response.Validation.Errors) == 0 {
		t.Error("Expected validation errors, but got none")
	}
}

func TestConfigValidateEndpoint(t *testing.T) {
	api, _, cleanup := setupTestAPI(t)
	defer cleanup()

	validRequest := ConfigRequest{
		RefreshInterval: 15,
		ServerPort:      8081,
		Title:           "Valid Dashboard",
		Theme: config.Theme{
			FontFamily: "serif",
			FontSize:   "16px",
			Background: "#ffffff",
			Foreground: "#000000",
		},
		Widgets: []config.Widget{
			{
				Name:    "Valid Widget",
				Script:  "valid.py",
				Enabled: true,
				Timeout: 10,
				Parameters: map[string]interface{}{
					"param": "value",
				},
			},
		},
	}

	req := createTestRequest("POST", "/api/admin/config/validate", validRequest)
	w := httptest.NewRecorder()

	api.handleConfigValidate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response ValidationResult
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Valid {
		t.Errorf("Expected validation to pass, but it failed: %v", response.Errors)
	}
}

// Widget Management Tests

func TestGetWidgets(t *testing.T) {
	api, _, cleanup := setupTestAPI(t)
	defer cleanup()

	req := createTestRequest("GET", "/api/admin/widgets", nil)
	w := httptest.NewRecorder()

	api.getWidgets(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var widgets []WidgetStatus
	if err := json.NewDecoder(w.Body).Decode(&widgets); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(widgets) != 1 {
		t.Errorf("Expected 1 widget, got %d", len(widgets))
	}

	if widgets[0].Widget.Name != "Test Widget" {
		t.Errorf("Expected widget name 'Test Widget', got '%s'", widgets[0].Widget.Name)
	}
}

func TestCreateWidget(t *testing.T) {
	api, _, cleanup := setupTestAPI(t)
	defer cleanup()

	newWidget := WidgetRequest{
		Name:    "New Widget",
		Script:  "new_widget.py",
		Enabled: true,
		Timeout: 20,
		Parameters: map[string]interface{}{
			"new_param": "new_value",
		},
	}

	req := createTestRequest("POST", "/api/admin/widgets", newWidget)
	w := httptest.NewRecorder()

	api.createWidget(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var widget config.Widget
	if err := json.NewDecoder(w.Body).Decode(&widget); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if widget.Name != "New Widget" {
		t.Errorf("Expected widget name 'New Widget', got '%s'", widget.Name)
	}

	// Verify widget was added to configuration
	cfg, err := config.LoadConfig(api.configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg.Widgets) != 2 {
		t.Errorf("Expected 2 widgets, got %d", len(cfg.Widgets))
	}
}

func TestUpdateWidget(t *testing.T) {
	api, _, cleanup := setupTestAPI(t)
	defer cleanup()

	// Setup router for URL parameters
	router := mux.NewRouter()
	router.HandleFunc("/api/admin/widgets/{id}", api.handleWidget).Methods("PUT")

	updatedWidget := WidgetRequest{
		Name:    "Updated Test Widget",
		Script:  "updated_test.py",
		Enabled: false,
		Timeout: 25,
		Parameters: map[string]interface{}{
			"updated_param": "updated_value",
		},
	}

	req := createTestRequest("PUT", "/api/admin/widgets/Test%20Widget", updatedWidget)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var widget config.Widget
	if err := json.NewDecoder(w.Body).Decode(&widget); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if widget.Name != "Updated Test Widget" {
		t.Errorf("Expected widget name 'Updated Test Widget', got '%s'", widget.Name)
	}

	if widget.Enabled {
		t.Error("Expected widget to be disabled")
	}
}

func TestDeleteWidget(t *testing.T) {
	api, _, cleanup := setupTestAPI(t)
	defer cleanup()

	// Setup router for URL parameters
	router := mux.NewRouter()
	router.HandleFunc("/api/admin/widgets/{id}", api.handleWidget).Methods("DELETE")

	req := createTestRequest("DELETE", "/api/admin/widgets/Test%20Widget", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify widget was removed from configuration
	cfg, err := config.LoadConfig(api.configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(cfg.Widgets) != 0 {
		t.Errorf("Expected 0 widgets, got %d", len(cfg.Widgets))
	}
}

func TestToggleWidget(t *testing.T) {
	api, _, cleanup := setupTestAPI(t)
	defer cleanup()

	// Setup router for URL parameters
	router := mux.NewRouter()
	router.HandleFunc("/api/admin/widgets/{id}/toggle", api.handleWidgetToggle).Methods("POST")

	req := createTestRequest("POST", "/api/admin/widgets/Test%20Widget/toggle", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	enabled, ok := response["enabled"].(bool)
	if !ok {
		t.Fatal("Expected 'enabled' field in response")
	}

	// Widget was initially enabled, should now be disabled
	if enabled {
		t.Error("Expected widget to be disabled after toggle")
	}

	// Verify in configuration
	cfg, err := config.LoadConfig(api.configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Widgets[0].Enabled {
		t.Error("Expected widget to be disabled in configuration")
	}
}

// System Status Tests

func TestGetSystemStatus(t *testing.T) {
	api, _, cleanup := setupTestAPI(t)
	defer cleanup()

	req := createTestRequest("GET", "/api/admin/status", nil)
	w := httptest.NewRecorder()

	api.handleSystemStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var status SystemStatus
	if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if status.Status == "" {
		t.Error("Expected status to be set")
	}

	if status.Uptime == 0 {
		t.Error("Expected uptime to be greater than 0")
	}
}

func TestGetMetrics(t *testing.T) {
	api, _, cleanup := setupTestAPI(t)
	defer cleanup()

	req := createTestRequest("GET", "/api/admin/metrics", nil)
	w := httptest.NewRecorder()

	api.handleMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var metrics map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&metrics); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if _, exists := metrics["system_metrics"]; !exists {
		t.Error("Expected 'system_metrics' in response")
	}
}

func TestGetLogs(t *testing.T) {
	api, _, cleanup := setupTestAPI(t)
	defer cleanup()

	// Add some test log entries
	api.metrics.AddLogEntry("info", "Test log entry", "test", nil)
	api.metrics.AddLogEntry("error", "Test error entry", "test", map[string]interface{}{
		"error_code": 500,
	})

	req := createTestRequest("GET", "/api/admin/logs?limit=10&level=info", nil)
	w := httptest.NewRecorder()

	api.handleLogs(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var logs []LogEntry
	if err := json.NewDecoder(w.Body).Decode(&logs); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should only get info logs due to level filter
	infoCount := 0
	for _, log := range logs {
		if log.Level == "info" {
			infoCount++
		}
	}

	if infoCount == 0 {
		t.Error("Expected at least one info log entry")
	}
}

// Backup Tests

func TestCreateBackup(t *testing.T) {
	api, _, cleanup := setupTestAPI(t)
	defer cleanup()

	req := createTestRequest("POST", "/api/admin/config/backup", nil)
	w := httptest.NewRecorder()

	api.handleConfigBackup(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	backupID, exists := response["backup_id"].(string)
	if !exists || backupID == "" {
		t.Error("Expected backup_id in response")
	}
}

func TestListBackups(t *testing.T) {
	api, _, cleanup := setupTestAPI(t)
	defer cleanup()

	// Create a backup first
	_, err := api.backup.CreateBackup()
	if err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	req := createTestRequest("GET", "/api/admin/config/backups", nil)
	w := httptest.NewRecorder()

	api.handleConfigBackups(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var backups []BackupInfo
	if err := json.NewDecoder(w.Body).Decode(&backups); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(backups) == 0 {
		t.Error("Expected at least one backup")
	}
}

// Error Handling Tests

func TestErrorHandling(t *testing.T) {
	api, configPath, cleanup := setupTestAPI(t)
	defer cleanup()

	// Test with invalid config path
	api.configPath = "/nonexistent/path/config.json"

	req := createTestRequest("GET", "/api/admin/config", nil)
	w := httptest.NewRecorder()

	api.getConfig(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	// Restore valid config path
	api.configPath = configPath
}

func TestInvalidJSON(t *testing.T) {
	api, _, cleanup := setupTestAPI(t)
	defer cleanup()

	req := httptest.NewRequest("PUT", "/api/admin/config", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	api.updateConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// Integration Tests

func TestCompleteWorkflow(t *testing.T) {
	api, _, cleanup := setupTestAPI(t)
	defer cleanup()

	// 1. Get initial config
	req := createTestRequest("GET", "/api/admin/config", nil)
	w := httptest.NewRecorder()
	api.getConfig(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get initial config: %d", w.Code)
	}

	// 2. Create a backup
	req = createTestRequest("POST", "/api/admin/config/backup", nil)
	w = httptest.NewRecorder()
	api.handleConfigBackup(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to create backup: %d", w.Code)
	}

	// 3. Add a new widget
	newWidget := WidgetRequest{
		Name:    "Integration Test Widget",
		Script:  "integration_test.py",
		Enabled: true,
		Timeout: 15,
		Parameters: map[string]interface{}{
			"test": true,
		},
	}

	req = createTestRequest("POST", "/api/admin/widgets", newWidget)
	w = httptest.NewRecorder()
	api.createWidget(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create widget: %d", w.Code)
	}

	// 4. Verify widget was added
	req = createTestRequest("GET", "/api/admin/widgets", nil)
	w = httptest.NewRecorder()
	api.getWidgets(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get widgets: %d", w.Code)
	}

	var widgets []WidgetStatus
	if err := json.NewDecoder(w.Body).Decode(&widgets); err != nil {
		t.Fatalf("Failed to decode widgets: %v", err)
	}

	if len(widgets) != 2 {
		t.Fatalf("Expected 2 widgets, got %d", len(widgets))
	}

	// 5. Check system status
	req = createTestRequest("GET", "/api/admin/status", nil)
	w = httptest.NewRecorder()
	api.handleSystemStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get system status: %d", w.Code)
	}

	var status SystemStatus
	if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to decode status: %v", err)
	}

	if status.ActiveWidgets < 1 {
		t.Error("Expected at least 1 active widget")
	}
}
