package dto

import (
	"time"
)

// ADKChatRequest represents a request to the ADK chat service
type ADKChatRequest struct {
	SessionID string                 `json:"session_id" validate:"required,min=1,max=255"`
	UserID    string                 `json:"user_id" validate:"required,min=1,max=255"`
	Message   string                 `json:"message" validate:"required,min=1,max=2000"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// ADKChatResponse represents a response from the ADK chat service
type ADKChatResponse struct {
	SessionID       string                 `json:"session_id"`
	Message         ADKChatMessageDTO      `json:"message"`
	SessionState    map[string]interface{} `json:"session_state"`
	Phase           string                 `json:"phase"`
	NextSuggestions []string               `json:"next_suggestions,omitempty"`
}

// ADKChatMessageDTO represents a chat message in the ADK system
type ADKChatMessageDTO struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // user, agent, system
	Content   string                 `json:"content"`
	AgentName string                 `json:"agent_name,omitempty"`
	Actions   []ADKChatActionDTO     `json:"actions,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// ADKChatActionDTO represents an actionable item from the ADK agent
type ADKChatActionDTO struct {
	Type       string                 `json:"type"` // template_suggestion, auto_populate, validation, etc.
	Label      string                 `json:"label"`
	Data       map[string]interface{} `json:"data"`
	Confidence float64                `json:"confidence"`
}

// ADKSessionResponse represents an ADK session
type ADKSessionResponse struct {
	SessionID    string                 `json:"session_id"`
	UserID       string                 `json:"user_id"`
	Messages     []ADKChatMessageDTO    `json:"messages"`
	State        map[string]interface{} `json:"state"`
	Phase        string                 `json:"phase"`
	CreatedAt    time.Time              `json:"created_at"`
	LastActivity time.Time              `json:"last_activity"`
}

// LLMAnalyzeRequest represents a request for LLM analysis
type LLMAnalyzeRequest struct {
	APIURL         string            `json:"api_url" validate:"required,url"`
	WidgetTemplate string            `json:"widget_template" validate:"required"`
	APIHeaders     map[string]string `json:"api_headers,omitempty"`
	SampleData     interface{}       `json:"sample_data,omitempty"`
}

// LLMAnalyzeResponse represents the response from LLM analysis
type LLMAnalyzeResponse struct {
	APIData     interface{}            `json:"api_data,omitempty"`
	DataMapping map[string]interface{} `json:"data_mapping"`
	Suggestions []MappingSuggestionDTO `json:"suggestions,omitempty"`
	Reasoning   string                 `json:"reasoning,omitempty"`
	Confidence  float64                `json:"confidence,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// EnhancedLLMRequest represents an enhanced LLM analysis request
type EnhancedLLMRequest struct {
	NaturalLanguage string                 `json:"natural_language,omitempty"`
	APIURL          string                 `json:"api_url,omitempty" validate:"omitempty,url"`
	WidgetTemplate  string                 `json:"widget_template,omitempty"`
	APIHeaders      map[string]string      `json:"api_headers,omitempty"`
	OpenAPISpec     interface{}            `json:"openapi_spec,omitempty"`
	UserIntent      string                 `json:"user_intent,omitempty"`
	Context         map[string]interface{} `json:"context,omitempty"`
}

// EnhancedLLMResponse represents an enhanced LLM analysis response
type EnhancedLLMResponse struct {
	LLMAnalyzeResponse
	AgentWorkflow       string                  `json:"agent_workflow,omitempty"`
	WorkflowResults     map[string]interface{}  `json:"workflow_results,omitempty"`
	AgentReasoning      []AgentReasoningStepDTO `json:"agent_reasoning,omitempty"`
	GeneratedWidget     *GeneratedWidgetDTO     `json:"generated_widget,omitempty"`
	ValidationResult    *ValidationResultDTO    `json:"validation_result,omitempty"`
	Documentation       *WidgetDocumentationDTO `json:"documentation,omitempty"`
	SessionStateUpdates map[string]interface{}  `json:"session_state_updates,omitempty"`
	SuggestedTemplates  []TemplateConfigDTO     `json:"suggested_templates,omitempty"`
	APIValidation       *APIValidationResultDTO `json:"api_validation,omitempty"`
}

// AgentReasoningStepDTO tracks agent decision-making process
type AgentReasoningStepDTO struct {
	AgentName   string      `json:"agent_name"`
	Step        string      `json:"step"`
	Input       interface{} `json:"input,omitempty"`
	Output      interface{} `json:"output,omitempty"`
	Reasoning   string      `json:"reasoning"`
	Confidence  float64     `json:"confidence"`
	Duration    int64       `json:"duration_ms"` // Duration in milliseconds
	Tools       []string    `json:"tools,omitempty"`
	NextActions []string    `json:"next_actions,omitempty"`
}

// GeneratedWidgetDTO represents a widget generated by agents
type GeneratedWidgetDTO struct {
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	TemplateType     string                 `json:"template_type"`
	DataMapping      map[string]interface{} `json:"data_mapping"`
	APIConfiguration *APIConfigurationDTO   `json:"api_configuration,omitempty"`
	RSSConfiguration *RSSConfigDTO          `json:"rss_configuration,omitempty"`
	Metadata         *WidgetMetadataDTO     `json:"metadata,omitempty"`
}

// APIConfigurationDTO holds enhanced API configuration
type APIConfigurationDTO struct {
	URL            string              `json:"url"`
	Headers        map[string]string   `json:"headers"`
	Method         string              `json:"method"`
	Timeout        int                 `json:"timeout"`
	Authentication *AuthConfigDTO      `json:"authentication,omitempty"`
	RateLimiting   *RateLimitConfigDTO `json:"rate_limiting,omitempty"`
	Caching        *CacheConfigDTO     `json:"caching,omitempty"`
}

// AuthConfigDTO for API authentication
type AuthConfigDTO struct {
	Type   string            `json:"type"`   // "apikey", "bearer", "basic", "oauth"
	Config map[string]string `json:"config"` // Flexible auth configuration
}

// RateLimitConfigDTO for API rate limiting
type RateLimitConfigDTO struct {
	RequestsPerMinute int    `json:"requests_per_minute"`
	BurstLimit        int    `json:"burst_limit"`
	BackoffStrategy   string `json:"backoff_strategy"`
	RetryAttempts     int    `json:"retry_attempts"`
}

// CacheConfigDTO for response caching
type CacheConfigDTO struct {
	TTL          int    `json:"ttl_seconds"`
	Strategy     string `json:"strategy"`     // "memory", "disk", "redis"
	Invalidation string `json:"invalidation"` // "ttl", "manual", "conditional"
}

// WidgetMetadataDTO provides additional widget information
type WidgetMetadataDTO struct {
	Category        string               `json:"category"`
	Tags            []string             `json:"tags"`
	Complexity      string               `json:"complexity"`       // "simple", "medium", "complex"
	UpdateFrequency string               `json:"update_frequency"` // "realtime", "frequent", "hourly", "daily"
	Dependencies    []string             `json:"dependencies,omitempty"`
	Performance     *PerformanceHintsDTO `json:"performance,omitempty"`
}

// PerformanceHintsDTO for widget optimization
type PerformanceHintsDTO struct {
	PreferredCacheTTL int      `json:"preferred_cache_ttl"`
	MemoryUsage       string   `json:"memory_usage"`  // "low", "medium", "high"
	CPUIntensity      string   `json:"cpu_intensity"` // "low", "medium", "high"
	NetworkUsage      string   `json:"network_usage"` // "low", "medium", "high"
	Optimizations     []string `json:"optimizations,omitempty"`
}

// ValidationResultDTO from widget validation agent
type ValidationResultDTO struct {
	Valid          bool                   `json:"valid"`
	Errors         []ValidationErrorDTO   `json:"errors,omitempty"`
	Warnings       []ValidationWarningDTO `json:"warnings,omitempty"`
	Suggestions    []string               `json:"suggestions,omitempty"`
	Compatibility  *CompatibilityCheckDTO `json:"compatibility,omitempty"`
	TestResults    []TestResultDTO        `json:"test_results,omitempty"`
	SecurityChecks []SecurityCheckDTO     `json:"security_checks,omitempty"`
}

// CompatibilityCheckDTO ensures widget compatibility
type CompatibilityCheckDTO struct {
	TemplateCompatible  bool     `json:"template_compatible"`
	APICompatible       bool     `json:"api_compatible"`
	FrameworkVersion    string   `json:"framework_version"`
	RequiredFeatures    []string `json:"required_features,omitempty"`
	ConflictingFeatures []string `json:"conflicting_features,omitempty"`
}

// TestResultDTO from automated testing
type TestResultDTO struct {
	TestName string  `json:"test_name"`
	Passed   bool    `json:"passed"`
	Duration int     `json:"duration_ms"`
	Message  string  `json:"message,omitempty"`
	Details  string  `json:"details,omitempty"`
	Coverage float64 `json:"coverage,omitempty"`
}

// SecurityCheckDTO for widget security validation
type SecurityCheckDTO struct {
	CheckType  string `json:"check_type"` // "xss", "injection", "auth", "data_exposure"
	Passed     bool   `json:"passed"`
	Risk       string `json:"risk"` // "low", "medium", "high", "critical"
	Message    string `json:"message"`
	Mitigation string `json:"mitigation,omitempty"`
}

// WidgetDocumentationDTO generated by documentation agent
type WidgetDocumentationDTO struct {
	UserGuide       string                    `json:"user_guide"`
	TechnicalSpecs  string                    `json:"technical_specs"`
	APIReference    string                    `json:"api_reference,omitempty"`
	Examples        []DocumentationExampleDTO `json:"examples,omitempty"`
	Troubleshooting []TroubleshootingTipDTO   `json:"troubleshooting,omitempty"`
	Changelog       []ChangelogEntryDTO       `json:"changelog,omitempty"`
}

// DocumentationExampleDTO provides usage examples
type DocumentationExampleDTO struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Code        string      `json:"code,omitempty"`
	Data        interface{} `json:"data,omitempty"`
	Screenshot  string      `json:"screenshot,omitempty"`
}

// TroubleshootingTipDTO for common issues
type TroubleshootingTipDTO struct {
	Issue      string   `json:"issue"`
	Symptoms   []string `json:"symptoms"`
	Solution   string   `json:"solution"`
	Prevention string   `json:"prevention,omitempty"`
	References []string `json:"references,omitempty"`
}

// ChangelogEntryDTO tracks widget evolution
type ChangelogEntryDTO struct {
	Version         string    `json:"version"`
	Date            time.Time `json:"date"`
	Changes         []string  `json:"changes"`
	BreakingChanges []string  `json:"breaking_changes,omitempty"`
	Migration       string    `json:"migration,omitempty"`
}

// TemplateConfigDTO represents a suggested widget template
type TemplateConfigDTO struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// APIValidationResultDTO holds API validation information
type APIValidationResultDTO struct {
	IsValid    bool                   `json:"is_valid"`
	URL        string                 `json:"url"`
	StatusCode int                    `json:"status_code,omitempty"`
	Message    string                 `json:"message,omitempty"`
	Schema     map[string]interface{} `json:"schema,omitempty"`
	Headers    map[string]string      `json:"headers,omitempty"`
	Errors     []string               `json:"errors,omitempty"`
	Warnings   []string               `json:"warnings,omitempty"`
}

// MappingSuggestionDTO represents an alternative mapping suggestion
type MappingSuggestionDTO struct {
	Field       string      `json:"field"`
	JSONPath    string      `json:"json_path"`
	Value       interface{} `json:"value"`
	Confidence  float64     `json:"confidence"`
	Description string      `json:"description"`
}

// NaturalLanguageRequest represents a natural language widget generation request
type NaturalLanguageRequest struct {
	Description string                 `json:"description" validate:"required,min=1,max=1000"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Examples    []string               `json:"examples,omitempty"`
}

// OpenAPISpecRequest represents an OpenAPI specification parsing request
type OpenAPISpecRequest struct {
	Specification interface{} `json:"specification" validate:"required"`
	WidgetType    string      `json:"widget_type,omitempty"`
	Endpoint      string      `json:"endpoint,omitempty"`
}
