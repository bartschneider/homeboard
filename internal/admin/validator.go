package admin

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bartosz/homeboard/internal/config"
)

// ConfigValidator validates configuration requests
type ConfigValidator struct {
	rules map[string]ValidationRule
}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		rules: make(map[string]ValidationRule),
	}
}

// ValidateConfig validates a complete configuration request
func (v *ConfigValidator) ValidateConfig(req *ConfigRequest) ValidationResult {
	result := ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []string{},
	}

	// Validate basic configuration fields
	v.validateRefreshInterval(req.RefreshInterval, &result)
	v.validateServerPort(req.ServerPort, &result)
	v.validateTitle(req.Title, &result)
	v.validateTheme(&req.Theme, &result)

	// Validate widgets
	v.validateWidgets(req.Widgets, &result)

	// Set overall validity
	result.Valid = len(result.Errors) == 0

	return result
}

// ValidateWidget validates a single widget request
func (v *ConfigValidator) ValidateWidget(req *WidgetRequest) error {
	result := ValidationResult{
		Valid:  true,
		Errors: []ValidationError{},
	}

	v.validateWidgetName(req.Name, &result)
	v.validateWidgetScript(req.Script, &result)
	v.validateWidgetTimeout(req.Timeout, &result)
	v.validateWidgetParameters(req.Parameters, &result)

	if !result.Valid {
		return fmt.Errorf("widget validation failed: %s", v.formatErrors(result.Errors))
	}

	return nil
}

// Individual validation methods

func (v *ConfigValidator) validateRefreshInterval(interval int, result *ValidationResult) {
	if interval < 1 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "refresh_interval",
			Message: "Refresh interval must be at least 1 minute",
			Code:    "MIN_VALUE",
			Value:   interval,
		})
	}

	if interval > 1440 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "refresh_interval",
			Message: "Refresh interval cannot exceed 24 hours (1440 minutes)",
			Code:    "MAX_VALUE",
			Value:   interval,
		})
	}

	if interval < 5 {
		result.Warnings = append(result.Warnings, "Refresh intervals below 5 minutes may impact e-paper display lifespan")
	}
}

func (v *ConfigValidator) validateServerPort(port int, result *ValidationResult) {
	if port < 1024 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "server_port",
			Message: "Server port must be 1024 or higher",
			Code:    "MIN_VALUE",
			Value:   port,
		})
	}

	if port > 65535 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "server_port",
			Message: "Server port cannot exceed 65535",
			Code:    "MAX_VALUE",
			Value:   port,
		})
	}

	// Check for common reserved ports
	reservedPorts := []int{22, 23, 25, 53, 80, 110, 143, 443, 993, 995}
	for _, reserved := range reservedPorts {
		if port == reserved {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Port %d is commonly reserved for other services", port))
			break
		}
	}
}

func (v *ConfigValidator) validateTitle(title string, result *ValidationResult) {
	if title == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "title",
			Message: "Title is required",
			Code:    "REQUIRED",
			Value:   title,
		})
		return
	}

	if len(title) > 100 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "title",
			Message: "Title cannot exceed 100 characters",
			Code:    "MAX_LENGTH",
			Value:   title,
		})
	}

	// Check for potentially problematic characters
	if strings.ContainsAny(title, "<>\"'&") {
		result.Warnings = append(result.Warnings, "Title contains HTML special characters that may cause display issues")
	}
}

func (v *ConfigValidator) validateTheme(theme *config.Theme, result *ValidationResult) {
	// Validate font family
	if theme.FontFamily == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "theme.font_family",
			Message: "Font family is required",
			Code:    "REQUIRED",
			Value:   theme.FontFamily,
		})
	}

	// Validate font size
	if theme.FontSize == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "theme.font_size",
			Message: "Font size is required",
			Code:    "REQUIRED",
			Value:   theme.FontSize,
		})
	} else {
		// Check if font size is a valid CSS value
		fontSizePattern := regexp.MustCompile(`^\d+(\.\d+)?(px|em|rem|pt|%)$`)
		if !fontSizePattern.MatchString(theme.FontSize) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "theme.font_size",
				Message: "Font size must be a valid CSS value (e.g., '16px', '1.2em')",
				Code:    "INVALID_FORMAT",
				Value:   theme.FontSize,
			})
		}
	}

	// Validate colors
	v.validateColor(theme.Background, "theme.background", result)
	v.validateColor(theme.Foreground, "theme.foreground", result)

	// E-paper specific warnings
	if theme.Background != "#ffffff" && theme.Background != "#000000" {
		result.Warnings = append(result.Warnings, "E-paper displays work best with pure white (#ffffff) or black (#000000) backgrounds")
	}

	if theme.Foreground != "#000000" && theme.Foreground != "#ffffff" {
		result.Warnings = append(result.Warnings, "E-paper displays work best with pure black (#000000) or white (#ffffff) text")
	}
}

func (v *ConfigValidator) validateColor(color, field string, result *ValidationResult) {
	if color == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   field,
			Message: "Color is required",
			Code:    "REQUIRED",
			Value:   color,
		})
		return
	}

	// Validate hex color format
	hexPattern := regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)
	if !hexPattern.MatchString(color) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   field,
			Message: "Color must be a valid hex color (e.g., '#ffffff' or '#fff')",
			Code:    "INVALID_FORMAT",
			Value:   color,
		})
	}
}

func (v *ConfigValidator) validateWidgets(widgets []config.Widget, result *ValidationResult) {
	if len(widgets) == 0 {
		result.Warnings = append(result.Warnings, "No widgets configured - dashboard will be empty")
		return
	}

	if len(widgets) > 20 {
		result.Warnings = append(result.Warnings, "More than 20 widgets may impact dashboard performance")
	}

	// Track widget names for uniqueness
	widgetNames := make(map[string]int)

	for i, widget := range widgets {
		fieldPrefix := fmt.Sprintf("widgets[%d]", i)

		v.validateWidgetName(widget.Name, result, fieldPrefix)
		v.validateWidgetScript(widget.Script, result, fieldPrefix)
		v.validateWidgetTimeout(widget.Timeout, result, fieldPrefix)
		v.validateWidgetParameters(widget.Parameters, result, fieldPrefix)

		// Check for duplicate names
		if count, exists := widgetNames[widget.Name]; exists {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fmt.Sprintf("%s.name", fieldPrefix),
				Message: fmt.Sprintf("Widget name '%s' is already used by widget[%d]", widget.Name, count),
				Code:    "DUPLICATE_NAME",
				Value:   widget.Name,
			})
		} else {
			widgetNames[widget.Name] = i
		}
	}
}

func (v *ConfigValidator) validateWidgetName(name string, result *ValidationResult, fieldPrefix ...string) {
	field := "name"
	if len(fieldPrefix) > 0 {
		field = fmt.Sprintf("%s.name", fieldPrefix[0])
	}

	if name == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   field,
			Message: "Widget name is required",
			Code:    "REQUIRED",
			Value:   name,
		})
		return
	}

	if len(name) > 50 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   field,
			Message: "Widget name cannot exceed 50 characters",
			Code:    "MAX_LENGTH",
			Value:   name,
		})
	}

	// Check for valid identifier characters
	namePattern := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\s_-]*$`)
	if !namePattern.MatchString(name) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   field,
			Message: "Widget name must start with alphanumeric character and contain only letters, numbers, spaces, hyphens, and underscores",
			Code:    "INVALID_FORMAT",
			Value:   name,
		})
	}
}

func (v *ConfigValidator) validateWidgetScript(script string, result *ValidationResult, fieldPrefix ...string) {
	field := "script"
	if len(fieldPrefix) > 0 {
		field = fmt.Sprintf("%s.script", fieldPrefix[0])
	}

	if script == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   field,
			Message: "Widget script path is required",
			Code:    "REQUIRED",
			Value:   script,
		})
		return
	}

	// Validate file extension
	ext := filepath.Ext(script)
	if ext != ".py" {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Widget script '%s' does not have .py extension", script))
	}

	// Check if file exists (relative to project root)
	if !filepath.IsAbs(script) {
		// For relative paths, we can't validate existence without knowing the working directory
		// Add a warning instead
		if !strings.HasPrefix(script, "widgets/") {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Widget script '%s' is not in the standard 'widgets/' directory", script))
		}
	} else {
		// For absolute paths, check if file exists
		if _, err := os.Stat(script); os.IsNotExist(err) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   field,
				Message: fmt.Sprintf("Widget script file does not exist: %s", script),
				Code:    "FILE_NOT_FOUND",
				Value:   script,
			})
		}
	}

	// Security check for path traversal
	if strings.Contains(script, "..") {
		result.Errors = append(result.Errors, ValidationError{
			Field:   field,
			Message: "Widget script path cannot contain '..' (path traversal)",
			Code:    "SECURITY_VIOLATION",
			Value:   script,
		})
	}
}

func (v *ConfigValidator) validateWidgetTimeout(timeout int, result *ValidationResult, fieldPrefix ...string) {
	field := "timeout"
	if len(fieldPrefix) > 0 {
		field = fmt.Sprintf("%s.timeout", fieldPrefix[0])
	}

	if timeout < 1 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   field,
			Message: "Widget timeout must be at least 1 second",
			Code:    "MIN_VALUE",
			Value:   timeout,
		})
	}

	if timeout > 300 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   field,
			Message: "Widget timeout cannot exceed 300 seconds (5 minutes)",
			Code:    "MAX_VALUE",
			Value:   timeout,
		})
	}

	if timeout > 60 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Widget timeout of %d seconds is quite long and may impact dashboard responsiveness", timeout))
	}
}

func (cv *ConfigValidator) validateWidgetParameters(params map[string]interface{}, result *ValidationResult, fieldPrefix ...string) {
	field := "parameters"
	if len(fieldPrefix) > 0 {
		field = fmt.Sprintf("%s.parameters", fieldPrefix[0])
	}

	// Parameters are optional, so nil/empty is valid
	if params == nil {
		return
	}

	// Check for reasonable parameter count
	if len(params) > 50 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Widget has %d parameters - consider simplifying configuration", len(params)))
	}

	// Validate parameter keys
	for key, value := range params {
		paramField := fmt.Sprintf("%s.%s", field, key)

		// Check key format
		keyPattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
		if !keyPattern.MatchString(key) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   paramField,
				Message: "Parameter key must start with a letter and contain only letters, numbers, and underscores",
				Code:    "INVALID_FORMAT",
				Value:   key,
			})
		}

		// Validate parameter value types
		cv.validateParameterValue(value, paramField, result)
	}
}

func (cv *ConfigValidator) validateParameterValue(value interface{}, field string, result *ValidationResult) {
	switch v := value.(type) {
	case string:
		if len(v) > 1000 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Parameter '%s' has a very long string value (%d characters)", field, len(v)))
		}
	case map[string]interface{}:
		// Recursively validate nested objects
		for key, nestedValue := range v {
			nestedField := fmt.Sprintf("%s.%s", field, key)
			cv.validateParameterValue(nestedValue, nestedField, result)
		}
	case []interface{}:
		if len(v) > 100 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Parameter '%s' has a very large array (%d items)", field, len(v)))
		}
		// Validate array elements
		for i, item := range v {
			itemField := fmt.Sprintf("%s[%d]", field, i)
			cv.validateParameterValue(item, itemField, result)
		}
	case float64:
		// JSON numbers are parsed as float64
		// Check for reasonable ranges to prevent overflow
		if v > 1e15 || v < -1e15 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Parameter '%s' has an extremely large number value", field))
		}
	case bool:
		// Booleans are always valid
	case nil:
		// Null values are valid
	default:
		result.Warnings = append(result.Warnings, fmt.Sprintf("Parameter '%s' has an unusual type: %T", field, value))
	}
}

// Helper methods

func (v *ConfigValidator) formatErrors(errors []ValidationError) string {
	if len(errors) == 0 {
		return "unknown validation error"
	}

	messages := make([]string, len(errors))
	for i, err := range errors {
		messages[i] = fmt.Sprintf("%s: %s", err.Field, err.Message)
	}

	return strings.Join(messages, "; ")
}
