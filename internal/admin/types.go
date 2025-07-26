package admin

import (
	"time"

	"github.com/bartosz/homeboard/internal/config"
)

// API Request/Response Models

// ConfigRequest represents a configuration update request
type ConfigRequest struct {
	RefreshInterval int             `json:"refresh_interval" validate:"min=1,max=1440"`
	ServerPort      int             `json:"server_port" validate:"min=1024,max=65535"`
	Title           string          `json:"title" validate:"required,max=100"`
	Theme           config.Theme    `json:"theme" validate:"required"`
	Widgets         []config.Widget `json:"widgets" validate:"dive"`
}

// ConfigResponse represents a configuration API response
type ConfigResponse struct {
	Config     *config.Config   `json:"config,omitempty"`
	Validation ValidationResult `json:"validation,omitempty"`
	Timestamp  time.Time        `json:"timestamp"`
}

// WidgetRequest represents a widget creation/update request
type WidgetRequest struct {
	Name       string                 `json:"name" validate:"required,max=50"`
	Script     string                 `json:"script" validate:"required"`
	Parameters map[string]interface{} `json:"parameters"`
	Timeout    int                    `json:"timeout" validate:"min=1,max=300"`
	Enabled    bool                   `json:"enabled"`
}

// WidgetTestRequest represents a widget test request
type WidgetTestRequest struct {
	Parameters map[string]interface{} `json:"parameters"`
	Timeout    int                    `json:"timeout,omitempty"`
}

// WidgetTestResponse represents a widget test result
type WidgetTestResponse struct {
	Success          bool          `json:"success"`
	Output           string        `json:"output"`
	Error            string        `json:"error,omitempty"`
	ExecutionTime    time.Duration `json:"execution_time"`
	Timestamp        time.Time     `json:"timestamp"`
	ValidationErrors []string      `json:"validation_errors,omitempty"`
}

// WidgetStatus represents widget runtime status
type WidgetStatus struct {
	Widget         config.Widget `json:"widget"`
	Status         string        `json:"status"` // "active", "disabled", "error", "idle"
	LastExecution  time.Time     `json:"last_execution"`
	ExecutionTime  time.Duration `json:"execution_time"`
	ErrorMessage   string        `json:"error_message,omitempty"`
	SuccessRate    float64       `json:"success_rate"`
	ExecutionCount int64         `json:"execution_count"`
}

// SystemStatus represents overall system status
type SystemStatus struct {
	Status        string        `json:"status"` // "running", "stopped", "error"
	Uptime        time.Duration `json:"uptime"`
	Version       string        `json:"version"`
	Metrics       SystemMetrics `json:"metrics"`
	ActiveWidgets int           `json:"active_widgets"`
	LastRefresh   time.Time     `json:"last_refresh"`
	ConfigFile    string        `json:"config_file"`
	ServerPort    int           `json:"server_port"`
}

// SystemMetrics represents system performance metrics
type SystemMetrics struct {
	CPUUsage       float64 `json:"cpu_usage"`
	MemoryUsage    float64 `json:"memory_usage"`
	DiskUsage      float64 `json:"disk_usage"`
	RequestCount   int64   `json:"request_count"`
	ErrorCount     int64   `json:"error_count"`
	AverageLatency float64 `json:"average_latency"`
	TotalUptime    int64   `json:"total_uptime"`
}

// ValidationResult represents configuration validation results
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []string          `json:"warnings,omitempty"`
}

// ValidationError represents a specific validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Value   interface{} `json:"value,omitempty"`
}

// ErrorResponse represents API error responses
type ErrorResponse struct {
	Error     string    `json:"error"`
	Details   string    `json:"details,omitempty"`
	Code      string    `json:"code,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Component string                 `json:"component"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// BackupInfo represents configuration backup information
type BackupInfo struct {
	ID          string    `json:"id"`
	Filename    string    `json:"filename"`
	CreatedAt   time.Time `json:"created_at"`
	Size        int64     `json:"size"`
	Description string    `json:"description"`
}

// WebSocket Message Types

// AdminMessage represents a WebSocket message
type AdminMessage struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// WebSocket message type constants
const (
	WSTypeConfigUpdate  = "config_update"
	WSTypeWidgetStatus  = "widget_status"
	WSTypeSystemMetrics = "system_metrics"
	WSTypeLogEntry      = "log_entry"
	WSTypeError         = "error"
	WSTypeNotification  = "notification"
)

// WidgetStatusUpdate represents a widget status change
type WidgetStatusUpdate struct {
	WidgetID      string        `json:"widget_id"`
	WidgetName    string        `json:"widget_name"`
	Status        string        `json:"status"`
	LastExecution time.Time     `json:"last_execution"`
	ExecutionTime time.Duration `json:"execution_time"`
	Error         string        `json:"error,omitempty"`
	Output        string        `json:"output,omitempty"`
}

// SystemMetricsUpdate represents system metrics update
type SystemMetricsUpdate struct {
	Timestamp time.Time     `json:"timestamp"`
	Metrics   SystemMetrics `json:"metrics"`
}

// NotificationMessage represents a notification
type NotificationMessage struct {
	ID       string               `json:"id"`
	Type     string               `json:"type"` // "info", "warning", "error", "success"
	Title    string               `json:"title"`
	Message  string               `json:"message"`
	Duration int                  `json:"duration"` // Duration in seconds, 0 for persistent
	Actions  []NotificationAction `json:"actions,omitempty"`
}

// NotificationAction represents an action in a notification
type NotificationAction struct {
	Label  string `json:"label"`
	Action string `json:"action"`
	Style  string `json:"style"` // "primary", "secondary", "danger"
}

// Widget Test Runner Types

// TestResult represents the result of a widget test
type TestResult struct {
	Success          bool                   `json:"success"`
	Output           string                 `json:"output"`
	Error            string                 `json:"error,omitempty"`
	ExecutionTime    time.Duration          `json:"execution_time"`
	Timestamp        time.Time              `json:"timestamp"`
	ValidationErrors []string               `json:"validation_errors,omitempty"`
	Parameters       map[string]interface{} `json:"parameters"`
	Environment      map[string]string      `json:"environment,omitempty"`
	ExitCode         int                    `json:"exit_code"`
}

// Performance Test Types

// PerformanceMetrics represents performance test results
type PerformanceMetrics struct {
	RequestCount     int64         `json:"request_count"`
	AverageLatency   time.Duration `json:"average_latency"`
	MinLatency       time.Duration `json:"min_latency"`
	MaxLatency       time.Duration `json:"max_latency"`
	P95Latency       time.Duration `json:"p95_latency"`
	ErrorRate        float64       `json:"error_rate"`
	ThroughputPerSec float64       `json:"throughput_per_sec"`
	MemoryUsage      int64         `json:"memory_usage"`
	CPUUsage         float64       `json:"cpu_usage"`
}

// LoadTestConfig represents load test configuration
type LoadTestConfig struct {
	Duration    time.Duration  `json:"duration"`
	Concurrency int            `json:"concurrency"`
	RequestRate int            `json:"request_rate"`
	TargetURL   string         `json:"target_url"`
	TestType    string         `json:"test_type"` // "load", "stress", "spike"
	Scenarios   []TestScenario `json:"scenarios"`
}

// TestScenario represents a test scenario
type TestScenario struct {
	Name       string                 `json:"name"`
	Weight     float64                `json:"weight"`
	Endpoint   string                 `json:"endpoint"`
	Method     string                 `json:"method"`
	Headers    map[string]string      `json:"headers,omitempty"`
	Body       string                 `json:"body,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Validation []ValidationRule       `json:"validation,omitempty"`
}

// ValidationRule represents a test validation rule
type ValidationRule struct {
	Type     string      `json:"type"` // "status_code", "response_time", "content"
	Expected interface{} `json:"expected"`
	Operator string      `json:"operator"` // "eq", "lt", "gt", "contains"
}

// Configuration Backup Types

// BackupMetadata represents backup metadata
type BackupMetadata struct {
	ID          string            `json:"id"`
	Filename    string            `json:"filename"`
	CreatedAt   time.Time         `json:"created_at"`
	Size        int64             `json:"size"`
	Checksum    string            `json:"checksum"`
	Description string            `json:"description"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// RestoreOptions represents restore operation options
type RestoreOptions struct {
	BackupID      string `json:"backup_id"`
	CreateBackup  bool   `json:"create_backup"`  // Create backup before restore
	ValidateFirst bool   `json:"validate_first"` // Validate backup before restore
	DryRun        bool   `json:"dry_run"`        // Test restore without applying
}

// Theme Management Types

// ThemeDefinition represents a theme configuration
type ThemeDefinition struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Author      string            `json:"author"`
	Version     string            `json:"version"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Theme       config.Theme      `json:"theme"`
	Preview     string            `json:"preview,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Variables   map[string]string `json:"variables,omitempty"`
}

// Widget Library Types

// WidgetTemplate represents a widget template
type WidgetTemplate struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	Author       string                `json:"author"`
	Version      string                `json:"version"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	Script       string                `json:"script"`
	Parameters   []ParameterDefinition `json:"parameters"`
	Dependencies []string              `json:"dependencies"`
	Examples     []ConfigExample       `json:"examples"`
	Category     string                `json:"category"`
	Tags         []string              `json:"tags,omitempty"`
	Icon         string                `json:"icon,omitempty"`
	Screenshots  []string              `json:"screenshots,omitempty"`
}

// ParameterDefinition represents a widget parameter definition
type ParameterDefinition struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // "string", "number", "boolean", "object", "array"
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Validation  string      `json:"validation,omitempty"`
	Options     []string    `json:"options,omitempty"` // For enum types
	Min         *float64    `json:"min,omitempty"`     // For number types
	Max         *float64    `json:"max,omitempty"`     // For number types
	Pattern     string      `json:"pattern,omitempty"` // For string validation
}

// ConfigExample represents a configuration example
type ConfigExample struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Screenshot  string                 `json:"screenshot,omitempty"`
}
