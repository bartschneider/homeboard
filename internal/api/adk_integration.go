package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ADKIntegrationService handles communication with the Java ADK microservice
type ADKIntegrationService struct {
	BaseURL    string
	HTTPClient *http.Client
}

// ADKChatRequest represents the request to the ADK service
type ADKChatRequest struct {
	SessionID string                 `json:"session_id"`
	UserID    string                 `json:"user_id"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// ADKChatResponse represents the response from the ADK service
type ADKChatResponse struct {
	SessionID       string                 `json:"session_id"`
	Message         ADKChatMessage         `json:"message"`
	SessionState    map[string]interface{} `json:"session_state"`
	Phase           string                 `json:"phase"`
	NextSuggestions []string               `json:"next_suggestions,omitempty"`
}

// ADKChatMessage represents a single message in the conversation
type ADKChatMessage struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Content   string                 `json:"content"`
	AgentName string                 `json:"agent_name,omitempty"`
	Actions   []ADKChatAction        `json:"actions,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// ADKChatAction represents an actionable item from the agent
type ADKChatAction struct {
	Type       string                 `json:"type"`
	Label      string                 `json:"label"`
	Data       map[string]interface{} `json:"data"`
	Confidence float64                `json:"confidence"`
}

// ADKSessionInfo represents session information from ADK service
type ADKSessionInfo struct {
	SessionID   string                 `json:"session_id"`
	UserID      string                 `json:"user_id"`
	State       map[string]interface{} `json:"state"`
	EventsCount int                    `json:"events_count"`
	LastUpdate  string                 `json:"last_update"`
}

// NewADKIntegrationService creates a new ADK integration service
func NewADKIntegrationService(baseURL string) *ADKIntegrationService {
	return &ADKIntegrationService{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProcessChatMessage sends a chat message to the ADK service
func (s *ADKIntegrationService) ProcessChatMessage(sessionID, userID, message string, context map[string]interface{}) (*ADKChatResponse, error) {
	request := ADKChatRequest{
		SessionID: sessionID,
		UserID:    userID,
		Message:   message,
		Context:   context,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.HTTPClient.Post(
		s.BaseURL+"/api/adk/chat",
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to ADK service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ADK service returned status %d: %s", resp.StatusCode, string(body))
	}

	var response ADKChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode ADK response: %w", err)
	}

	return &response, nil
}

// GetSession retrieves session information from the ADK service
func (s *ADKIntegrationService) GetSession(sessionID string) (*ADKSessionInfo, error) {
	resp, err := s.HTTPClient.Get(s.BaseURL + "/api/adk/sessions/" + sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session from ADK service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Session not found
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ADK service returned status %d: %s", resp.StatusCode, string(body))
	}

	var sessionInfo ADKSessionInfo
	if err := json.NewDecoder(resp.Body).Decode(&sessionInfo); err != nil {
		return nil, fmt.Errorf("failed to decode session info: %w", err)
	}

	return &sessionInfo, nil
}

// HealthCheck checks if the ADK service is healthy
func (s *ADKIntegrationService) HealthCheck() error {
	resp, err := s.HTTPClient.Get(s.BaseURL + "/api/adk/health")
	if err != nil {
		return fmt.Errorf("failed to check ADK service health: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ADK service health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// CleanupOldSessions triggers cleanup of old sessions in the ADK service
func (s *ADKIntegrationService) CleanupOldSessions() error {
	resp, err := s.HTTPClient.Post(s.BaseURL+"/api/adk/maintenance/cleanup", "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to trigger cleanup: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cleanup failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
