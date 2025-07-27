package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ChatSessionService manages interactive chat sessions for widget building
type ChatSessionService struct {
	enhancedLLM *EnhancedLLMService
	rssService  *RSSService
	httpClient  *http.Client
	sessions    map[string]*ChatSession
}

// ChatSession represents an ongoing conversation with context
type ChatSession struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	Messages     []ChatMessage          `json:"messages"`
	State        map[string]interface{} `json:"state"`
	CreatedAt    time.Time              `json:"created_at"`
	LastActivity time.Time              `json:"last_activity"`
	Phase        string                 `json:"phase"` // discovery, configuration, validation, completion
}

// ChatMessage represents a single message in the conversation
type ChatMessage struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // user, agent, system
	Content   string                 `json:"content"`
	AgentName string                 `json:"agent_name,omitempty"`
	Actions   []ChatAction           `json:"actions,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// ChatAction represents an actionable item from the agent
type ChatAction struct {
	Type       string                 `json:"type"` // template_suggestion, auto_populate, validation, etc.
	Label      string                 `json:"label"`
	Data       map[string]interface{} `json:"data"`
	Confidence float64                `json:"confidence"`
}

// ChatRequest represents an incoming chat message
type ChatRequest struct {
	SessionID string                 `json:"session_id"`
	UserID    string                 `json:"user_id"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// ChatResponse represents the agent's response
type ChatResponse struct {
	SessionID       string                 `json:"session_id"`
	Message         ChatMessage            `json:"message"`
	SessionState    map[string]interface{} `json:"session_state"`
	Phase           string                 `json:"phase"`
	NextSuggestions []string               `json:"next_suggestions,omitempty"`
}

// NewChatSessionService creates a new chat session service
func NewChatSessionService(enhancedLLM *EnhancedLLMService, rssService *RSSService) *ChatSessionService {
	return &ChatSessionService{
		enhancedLLM: enhancedLLM,
		rssService:  rssService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		sessions: make(map[string]*ChatSession),
	}
}

// GetOrCreateSession gets existing session or creates new one
func (s *ChatSessionService) GetOrCreateSession(sessionID, userID string) *ChatSession {
	if session, exists := s.sessions[sessionID]; exists {
		session.LastActivity = time.Now()
		return session
	}

	session := &ChatSession{
		ID:           sessionID,
		UserID:       userID,
		Messages:     []ChatMessage{},
		State:        make(map[string]interface{}),
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		Phase:        "discovery",
	}

	// Add initial agent message
	initialMessage := ChatMessage{
		ID:        generateMessageID(),
		Type:      "agent",
		Content:   "🤖 Hello! I'm your AI Widget Builder assistant. I'll help you create the perfect widget for your dashboard. Tell me what kind of widget you'd like to create, or provide an API URL and I'll analyze it for you.",
		AgentName: "WidgetBuilderAgent",
		Timestamp: time.Now(),
		Actions: []ChatAction{
			{
				Type:       "suggestion",
				Label:      "Get started with examples",
				Data:       map[string]interface{}{"examples": []string{"Weather widget", "API status widget", "RSS feed widget"}},
				Confidence: 1.0,
			},
		},
	}
	session.Messages = append(session.Messages, initialMessage)

	s.sessions[sessionID] = session
	return session
}

// ProcessMessage handles incoming user message and generates agent response
func (s *ChatSessionService) ProcessMessage(request ChatRequest) (*ChatResponse, error) {
	session := s.GetOrCreateSession(request.SessionID, request.UserID)

	// Add user message to session
	userMessage := ChatMessage{
		ID:        generateMessageID(),
		Type:      "user",
		Content:   request.Message,
		Timestamp: time.Now(),
	}
	session.Messages = append(session.Messages, userMessage)

	// Determine agent workflow based on message content and session state
	agentWorkflow := s.determineWorkflow(request.Message, session)

	// Build context for enhanced LLM
	context := map[string]interface{}{
		"session_id":           session.ID,
		"user_id":              session.UserID,
		"conversation_history": s.getRecentMessages(session, 5),
		"session_state":        session.State,
		"current_phase":        session.Phase,
	}

	// Merge any additional context from request
	for k, v := range request.Context {
		context[k] = v
	}

	// Call enhanced LLM service
	enhancedRequest := EnhancedAnalyzeRequest{
		NaturalLanguage: request.Message,
		Context:         context,
		AgentWorkflow:   agentWorkflow,
		UserIntent:      "widget_building_conversation",
	}

	llmResponse, err := s.enhancedLLM.AnalyzeWithEnhancedAgents(enhancedRequest)
	if err != nil {
		return nil, fmt.Errorf("LLM analysis failed: %w", err)
	}

	// Process LLM response and update session state
	agentMessage, updatedPhase := s.processLLMResponse(llmResponse, session)

	// Update session
	session.Messages = append(session.Messages, agentMessage)
	session.Phase = updatedPhase
	session.LastActivity = time.Now()

	// Apply any state updates from LLM response
	if llmResponse.SessionStateUpdates != nil {
		for k, v := range llmResponse.SessionStateUpdates {
			session.State[k] = v
		}
	}

	// Generate next suggestions based on current phase
	nextSuggestions := s.generateNextSuggestions(session)

	return &ChatResponse{
		SessionID:       session.ID,
		Message:         agentMessage,
		SessionState:    session.State,
		Phase:           session.Phase,
		NextSuggestions: nextSuggestions,
	}, nil
}

// determineWorkflow decides which agent workflow to use based on message and context
func (s *ChatSessionService) determineWorkflow(message string, session *ChatSession) string {
	messageLower := strings.ToLower(message)

	// Check for API URLs
	if strings.Contains(messageLower, "http") && (strings.Contains(messageLower, "api") || strings.Contains(messageLower, ".com") || strings.Contains(messageLower, ".org")) {
		return "api_analysis_and_mapping"
	}

	// Check for OpenAPI/Swagger specs
	if strings.Contains(messageLower, "openapi") || strings.Contains(messageLower, "swagger") {
		return "openapi_specification_parsing"
	}

	// Check for specific widget types
	if strings.Contains(messageLower, "weather") || strings.Contains(messageLower, "temperature") {
		return "weather_widget_generation"
	}

	if strings.Contains(messageLower, "rss") || strings.Contains(messageLower, "feed") || strings.Contains(messageLower, "news") {
		return "rss_widget_generation"
	}

	// Based on session phase
	switch session.Phase {
	case "discovery":
		return "natural_language_widget_generation"
	case "configuration":
		return "widget_configuration_refinement"
	case "validation":
		return "widget_validation_and_testing"
	default:
		return "comprehensive"
	}
}

// processLLMResponse converts LLM response to chat message and determines phase
func (s *ChatSessionService) processLLMResponse(llmResponse *EnhancedAnalyzeResponse, session *ChatSession) (ChatMessage, string) {
	// Determine which agent provided the primary response
	agentName := "WidgetBuilderAgent"
	if len(llmResponse.AgentReasoning) > 0 {
		agentName = llmResponse.AgentReasoning[len(llmResponse.AgentReasoning)-1].AgentName
	}

	// Extract main message content
	content := llmResponse.Reasoning
	if content == "" && llmResponse.GeneratedWidget != nil {
		content = fmt.Sprintf("I've analyzed your request and configured a widget for you. The widget type is '%s' and I've set up the initial configuration.", llmResponse.GeneratedWidget.TemplateType)
	}
	if content == "" {
		content = "I understand your request. Let me help you build that widget."
	}

	// Build actions based on LLM response
	actions := []ChatAction{}

	// Template suggestions
	if len(llmResponse.SuggestedTemplates) > 0 {
		for _, template := range llmResponse.SuggestedTemplates {
			actions = append(actions, ChatAction{
				Type:       "template_suggestion",
				Label:      fmt.Sprintf("Use %s", template.Name),
				Data:       map[string]interface{}{"template": template},
				Confidence: template.Confidence,
			})
		}
	}

	// Auto-population action
	if llmResponse.GeneratedWidget != nil {
		actions = append(actions, ChatAction{
			Type:       "auto_populate",
			Label:      "Apply suggested configuration",
			Data:       map[string]interface{}{"widget": llmResponse.GeneratedWidget},
			Confidence: llmResponse.Confidence,
		})
	}

	// API validation actions
	if llmResponse.APIValidation != nil && llmResponse.APIValidation.IsValid {
		actions = append(actions, ChatAction{
			Type:       "api_validation",
			Label:      "API endpoint validated ✓",
			Data:       map[string]interface{}{"validation": llmResponse.APIValidation},
			Confidence: 0.9,
		})
	}

	// Determine next phase
	nextPhase := session.Phase
	if llmResponse.GeneratedWidget != nil && session.Phase == "discovery" {
		nextPhase = "configuration"
	} else if llmResponse.WorkflowResults != nil && session.Phase == "configuration" {
		nextPhase = "validation"
	}

	message := ChatMessage{
		ID:        generateMessageID(),
		Type:      "agent",
		Content:   content,
		AgentName: agentName,
		Actions:   actions,
		Metadata: map[string]interface{}{
			"confidence":      llmResponse.Confidence,
			"workflow":        llmResponse.AgentWorkflow,
			"reasoning_steps": len(llmResponse.AgentReasoning),
		},
		Timestamp: time.Now(),
	}

	return message, nextPhase
}

// getRecentMessages returns the last N messages for context
func (s *ChatSessionService) getRecentMessages(session *ChatSession, count int) []ChatMessage {
	if len(session.Messages) <= count {
		return session.Messages
	}
	return session.Messages[len(session.Messages)-count:]
}

// generateNextSuggestions provides contextual suggestions based on current phase
func (s *ChatSessionService) generateNextSuggestions(session *ChatSession) []string {
	switch session.Phase {
	case "discovery":
		return []string{
			"Describe the type of data you want to display",
			"Provide an API URL for automatic analysis",
			"Tell me about your dashboard requirements",
		}
	case "configuration":
		return []string{
			"Adjust the widget settings",
			"Test the API connection",
			"Preview the widget",
		}
	case "validation":
		return []string{
			"Save the widget",
			"Make final adjustments",
			"Create another widget",
		}
	case "completion":
		return []string{
			"Create another widget",
			"Go to dashboard builder",
			"View all widgets",
		}
	default:
		return []string{"How can I help you with your widget?"}
	}
}

// generateMessageID creates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// GetSession retrieves a session by ID
func (s *ChatSessionService) GetSession(sessionID string) (*ChatSession, bool) {
	session, exists := s.sessions[sessionID]
	return session, exists
}

// CleanupOldSessions removes inactive sessions (older than 24 hours)
func (s *ChatSessionService) CleanupOldSessions() {
	cutoff := time.Now().Add(-24 * time.Hour)
	for id, session := range s.sessions {
		if session.LastActivity.Before(cutoff) {
			delete(s.sessions, id)
		}
	}
}
