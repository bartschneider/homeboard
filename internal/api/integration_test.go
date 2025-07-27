package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bartosz/homeboard/internal/api/dto"
)

// IntegrationTestSuite provides setup for integration tests
type IntegrationTestSuite struct {
	server      *httptest.Server
	adkServer   *httptest.Server
	apiHandlers *APIHandlers
}

// SetupIntegrationTest initializes the test environment
func SetupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	// Create mock ADK server
	adkServer := createMockADKServer(t)

	// Create API handlers with mock ADK URL (simplified for testing)
	handlers := &APIHandlers{}

	// Create test server
	server := httptest.NewServer(createTestRouter(handlers))

	return &IntegrationTestSuite{
		server:      server,
		adkServer:   adkServer,
		apiHandlers: handlers,
	}
}

// Cleanup tears down the test environment
func (suite *IntegrationTestSuite) Cleanup() {
	if suite.server != nil {
		suite.server.Close()
	}
	if suite.adkServer != nil {
		suite.adkServer.Close()
	}
	// Database cleanup removed for simplified testing
}

// createMockADKServer creates a mock Java ADK service for testing
func createMockADKServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/adk/chat":
			handleMockADKChat(w, r, t)
		case "/api/adk/health":
			handleMockADKHealth(w, r, t)
		default:
			http.NotFound(w, r)
		}
	}))
}

func handleMockADKChat(w http.ResponseWriter, r *http.Request, t *testing.T) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request dto.ADKChatRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Create mock response based on request
	response := createMockADKResponse(request, t)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		t.Errorf("Failed to encode response: %v", err)
	}
}

func handleMockADKHealth(w http.ResponseWriter, r *http.Request, t *testing.T) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "Mock ADK Widget Builder",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(health); err != nil {
		t.Errorf("Failed to encode health response: %v", err)
	}
}

func createMockADKResponse(request dto.ADKChatRequest, t *testing.T) dto.ADKChatResponse {
	// Simulate different responses based on message content
	var agentName, phase, content string
	var actions []dto.ADKChatActionDTO
	var suggestions []string

	message := request.Message
	switch {
	case containsIgnoreCase(message, "weather"):
		agentName = "WidgetDesigner"
		phase = "configuration"
		content = "I'll help you create a weather widget. What's your preferred weather API?"
		actions = []dto.ADKChatActionDTO{
			{
				Type:       "template_suggestion",
				Label:      "Use Weather Template",
				Data:       map[string]interface{}{"template": map[string]interface{}{"type": "weather", "name": "Weather Widget"}},
				Confidence: 0.9,
			},
		}
		suggestions = []string{"Provide weather API URL", "Use OpenWeatherMap", "Configure location"}

	case containsIgnoreCase(message, "api"):
		agentName = "APIAnalyzer"
		phase = "discovery"
		content = "I can analyze this API endpoint for you. Please provide the full API URL."
		actions = []dto.ADKChatActionDTO{
			{
				Type:       "api_validation",
				Label:      "API endpoint detected",
				Data:       map[string]interface{}{"validation": map[string]interface{}{"status": "detected"}},
				Confidence: 0.8,
			},
		}
		suggestions = []string{"Provide API documentation", "Test API endpoint", "Configure authentication"}

	case containsIgnoreCase(message, "rss") || containsIgnoreCase(message, "news"):
		agentName = "WidgetDesigner"
		phase = "configuration"
		content = "I'll help you create an RSS feed widget. Please provide the RSS feed URL."
		actions = []dto.ADKChatActionDTO{
			{
				Type:       "template_suggestion",
				Label:      "Use RSS Template",
				Data:       map[string]interface{}{"template": map[string]interface{}{"type": "rss", "name": "RSS Feed Widget"}},
				Confidence: 0.9,
			},
		}
		suggestions = []string{"Provide RSS feed URL", "Configure item count", "Set update frequency"}

	default:
		agentName = "Coordinator"
		phase = "discovery"
		content = "Hello! I'm your AI Widget Builder assistant. What kind of widget would you like to create?"
		actions = []dto.ADKChatActionDTO{
			{
				Type:       "suggestion",
				Label:      "Get started with examples",
				Data:       map[string]interface{}{"examples": []string{"Weather widget", "API status widget", "RSS feed widget"}},
				Confidence: 1.0,
			},
		}
		suggestions = []string{"Tell me what data you want to display", "Provide an API URL", "Choose from widget templates"}
	}

	return dto.ADKChatResponse{
		SessionID: request.SessionID,
		Message: dto.ADKChatMessageDTO{
			ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
			Type:      "agent",
			Content:   content,
			AgentName: agentName,
			Actions:   actions,
			Metadata:  map[string]interface{}{"confidence": 0.9, "phase": phase},
			Timestamp: time.Now(),
		},
		SessionState:    map[string]interface{}{"agent_used": agentName, "phase": phase},
		Phase:           phase,
		NextSuggestions: suggestions,
	}
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					bytes.Contains([]byte(s), []byte(substr))))
}

// Integration Tests

func TestIntegration_ADKChatWorkflow_Weather(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup()

	// Test weather widget creation workflow
	chatRequest := dto.ADKChatRequest{
		SessionID: "test-session-weather",
		UserID:    "test-user-123",
		Message:   "I want to create a weather widget",
		Context:   map[string]interface{}{"intent": "widget_creation"},
	}

	// Make request to Go backend, which should forward to Java ADK service
	response := makeADKChatRequest(t, suite.server.URL, chatRequest)

	// Verify response
	if response.SessionID != chatRequest.SessionID {
		t.Errorf("Expected session ID %s, got %s", chatRequest.SessionID, response.SessionID)
	}

	if response.Phase != "configuration" {
		t.Errorf("Expected phase 'configuration', got %s", response.Phase)
	}

	if response.Message.AgentName != "WidgetDesigner" {
		t.Errorf("Expected agent 'WidgetDesigner', got %s", response.Message.AgentName)
	}

	if !bytes.Contains([]byte(response.Message.Content), []byte("weather")) {
		t.Error("Expected response to mention weather")
	}

	if len(response.Message.Actions) == 0 {
		t.Error("Expected actions in response")
	}

	// Verify template suggestion action
	hasTemplateSuggestion := false
	for _, action := range response.Message.Actions {
		if action.Type == "template_suggestion" && action.Label == "Use Weather Template" {
			hasTemplateSuggestion = true
			break
		}
	}

	if !hasTemplateSuggestion {
		t.Error("Expected template suggestion action")
	}
}

func TestIntegration_ADKChatWorkflow_API(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup()

	// Test API widget creation workflow
	chatRequest := dto.ADKChatRequest{
		SessionID: "test-session-api",
		UserID:    "test-user-456",
		Message:   "Analyze this API: https://api.github.com/users/octocat",
		Context:   map[string]interface{}{"intent": "api_analysis"},
	}

	// Make request
	response := makeADKChatRequest(t, suite.server.URL, chatRequest)

	// Verify response
	if response.Phase != "discovery" {
		t.Errorf("Expected phase 'discovery', got %s", response.Phase)
	}

	if response.Message.AgentName != "APIAnalyzer" {
		t.Errorf("Expected agent 'APIAnalyzer', got %s", response.Message.AgentName)
	}

	if !bytes.Contains([]byte(response.Message.Content), []byte("API")) {
		t.Error("Expected response to mention API")
	}

	// Verify API validation action
	hasAPIValidation := false
	for _, action := range response.Message.Actions {
		if action.Type == "api_validation" {
			hasAPIValidation = true
			break
		}
	}

	if !hasAPIValidation {
		t.Error("Expected API validation action")
	}
}

func TestIntegration_ADKService_ErrorHandling(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup()

	// Stop the mock ADK server to simulate service unavailability
	suite.adkServer.Close()

	chatRequest := dto.ADKChatRequest{
		SessionID: "test-session-error",
		UserID:    "test-user-error",
		Message:   "Hello",
		Context:   map[string]interface{}{},
	}

	// Make request - should handle ADK service error gracefully
	url := fmt.Sprintf("%s/api/chat/widget-builder", suite.server.URL)
	requestBody, _ := json.Marshal(chatRequest)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should return error response but not crash
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}
}

func TestIntegration_ADKHealthCheck(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup()

	// Test health check endpoint
	url := fmt.Sprintf("%s/api/health", suite.server.URL)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to make health check request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode health response: %v", err)
	}

	if status, ok := health["status"].(string); !ok || status != "healthy" {
		t.Errorf("Expected healthy status, got %v", health["status"])
	}
}

func TestIntegration_WidgetCRUD_WithADK(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup()

	// Step 1: Use ADK to design a widget
	chatRequest := dto.ADKChatRequest{
		SessionID: "test-session-crud",
		UserID:    "test-user-crud",
		Message:   "Create a weather widget for London",
		Context:   map[string]interface{}{},
	}

	_ = makeADKChatRequest(t, suite.server.URL, chatRequest)

	// Step 2: Create widget based on ADK suggestions
	createRequest := dto.CreateWidgetRequest{
		Name:         "London Weather Widget",
		TemplateType: "weather_current",
		DataSource:   "api",
		APIURL:       "https://api.openweathermap.org/data/2.5/weather?q=London",
		APIHeaders:   map[string]string{"Authorization": "Bearer test-key"},
		Description:  "Weather widget created via ADK conversation",
		Timeout:      30,
	}

	// Create widget via API
	widget := makeCreateWidgetRequest(t, suite.server.URL, createRequest)

	if widget.Name != createRequest.Name {
		t.Errorf("Expected widget name %s, got %s", createRequest.Name, widget.Name)
	}

	// Step 3: Verify widget was created and can be retrieved
	retrievedWidget := makeGetWidgetRequest(t, suite.server.URL, widget.ID)

	if retrievedWidget.ID != widget.ID {
		t.Errorf("Expected widget ID %d, got %d", widget.ID, retrievedWidget.ID)
	}

	// Step 4: Update widget
	newName := "Updated London Weather Widget"
	updateRequest := dto.UpdateWidgetRequest{
		Name: &newName,
	}

	updatedWidget := makeUpdateWidgetRequest(t, suite.server.URL, widget.ID, updateRequest)

	if updatedWidget.Name != newName {
		t.Errorf("Expected updated name %s, got %s", newName, updatedWidget.Name)
	}

	// Step 5: List widgets
	widgets := makeListWidgetsRequest(t, suite.server.URL)

	if len(widgets.Widgets) != 1 {
		t.Errorf("Expected 1 widget, got %d", len(widgets.Widgets))
	}

	// Step 6: Delete widget
	makeDeleteWidgetRequest(t, suite.server.URL, widget.ID)

	// Verify deletion
	widgets = makeListWidgetsRequest(t, suite.server.URL)

	if len(widgets.Widgets) != 0 {
		t.Errorf("Expected 0 widgets after deletion, got %d", len(widgets.Widgets))
	}
}

// Helper functions for making HTTP requests

func makeADKChatRequest(t *testing.T, baseURL string, request dto.ADKChatRequest) dto.ADKChatResponse {
	url := fmt.Sprintf("%s/api/chat/widget-builder", baseURL)
	requestBody, _ := json.Marshal(request)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Failed to make ADK chat request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var response dto.ADKChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode ADK response: %v", err)
	}

	return response
}

func makeCreateWidgetRequest(t *testing.T, baseURL string, request dto.CreateWidgetRequest) dto.WidgetResponse {
	url := fmt.Sprintf("%s/api/widgets", baseURL)
	requestBody, _ := json.Marshal(request)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Failed to make create widget request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", resp.StatusCode)
	}

	var response dto.WidgetResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode widget response: %v", err)
	}

	return response
}

func makeGetWidgetRequest(t *testing.T, baseURL string, widgetID int) dto.WidgetResponse {
	url := fmt.Sprintf("%s/api/widgets/%d", baseURL, widgetID)

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to make get widget request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var response dto.WidgetResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode widget response: %v", err)
	}

	return response
}

func makeUpdateWidgetRequest(t *testing.T, baseURL string, widgetID int, request dto.UpdateWidgetRequest) dto.WidgetResponse {
	url := fmt.Sprintf("%s/api/widgets/%d", baseURL, widgetID)
	requestBody, _ := json.Marshal(request)

	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make update widget request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var response dto.WidgetResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode widget response: %v", err)
	}

	return response
}

func makeListWidgetsRequest(t *testing.T, baseURL string) dto.WidgetListResponse {
	url := fmt.Sprintf("%s/api/widgets", baseURL)

	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to make list widgets request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var response dto.WidgetListResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode widget list response: %v", err)
	}

	return response
}

func makeDeleteWidgetRequest(t *testing.T, baseURL string, widgetID int) {
	url := fmt.Sprintf("%s/api/widgets/%d", baseURL, widgetID)

	req, _ := http.NewRequest(http.MethodDelete, url, nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make delete widget request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("Expected status 204, got %d", resp.StatusCode)
	}
}

// createTestRouter creates a test router (simplified for testing)
func createTestRouter(handlers *APIHandlers) http.Handler {
	mux := http.NewServeMux()

	// Mock chat endpoint that responds successfully
	mux.HandleFunc("/api/chat/widget-builder", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			response := dto.ADKChatResponse{
				SessionID: "test-session",
				Message: dto.ADKChatMessageDTO{
					ID:        "msg_test",
					Type:      "agent",
					Content:   "Mock response",
					AgentName: "MockAgent",
					Actions:   []dto.ADKChatActionDTO{},
					Metadata:  map[string]interface{}{},
					Timestamp: time.Now(),
				},
				SessionState:    map[string]interface{}{},
				Phase:           "discovery",
				NextSuggestions: []string{"Continue"},
			}
			json.NewEncoder(w).Encode(response)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Mock health endpoint
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			health := map[string]interface{}{
				"status":    "healthy",
				"service":   "Test Widget Builder",
				"timestamp": time.Now().Format(time.RFC3339),
			}
			json.NewEncoder(w).Encode(health)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Mock widget CRUD endpoints for integration testing
	mux.HandleFunc("/api/widgets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			// Mock create widget
			response := dto.WidgetResponse{
				ID:           1,
				Name:         "Test Widget",
				TemplateType: "weather_current",
				DataSource:   "api",
				Enabled:      true,
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(response)
		case http.MethodGet:
			// Mock list widgets
			response := dto.WidgetListResponse{
				Widgets: []dto.WidgetSummaryResponse{},
				Pagination: dto.PaginationResponse{
					Page:  1,
					Limit: 10,
					Total: 0,
				},
			}
			json.NewEncoder(w).Encode(response)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Mock individual widget operations
	mux.HandleFunc("/api/widgets/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			response := dto.WidgetResponse{
				ID:           1,
				Name:         "Test Widget",
				TemplateType: "weather_current",
				DataSource:   "api",
				Enabled:      true,
			}
			json.NewEncoder(w).Encode(response)
		case http.MethodPut:
			response := dto.WidgetResponse{
				ID:           1,
				Name:         "Updated Widget",
				TemplateType: "weather_current",
				DataSource:   "api",
				Enabled:      false,
			}
			json.NewEncoder(w).Encode(response)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}
