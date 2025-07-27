package widget

import (
	"testing"
	"time"
)

func TestWidget_Creation_Success(t *testing.T) {
	// Given
	name := "Test Weather Widget"
	templateType := WeatherCurrent
	dataSource := APIDataSource

	// When
	widget, err := NewWidget(name, templateType, dataSource)

	// Then
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if widget == nil {
		t.Fatal("Expected widget, got nil")
	}

	if widget.Name() != name {
		t.Errorf("Expected name %s, got %s", name, widget.Name())
	}

	if widget.TemplateType() != templateType {
		t.Errorf("Expected template type %s, got %s", templateType, widget.TemplateType())
	}

	if widget.DataSource() != dataSource {
		t.Errorf("Expected data source %s, got %s", dataSource, widget.DataSource())
	}

	if !widget.Enabled() {
		t.Error("Expected widget to be enabled by default")
	}
}

func TestWidget_Creation_InvalidName(t *testing.T) {
	// Given - only empty names and overly long names should fail per validation logic
	invalidNames := []string{"", string(make([]byte, 256))}

	for _, name := range invalidNames {
		// When
		widget, err := NewWidget(name, WeatherCurrent, APIDataSource)

		// Then
		if err == nil {
			t.Errorf("Expected error for invalid name '%s', got nil", name)
		}

		if widget != nil {
			t.Errorf("Expected nil widget for invalid name '%s', got widget", name)
		}
	}
}

func TestWidget_Enable_Disable(t *testing.T) {
	// Given
	widget, err := NewWidget("Test Widget", WeatherCurrent, APIDataSource)
	if err != nil {
		t.Fatalf("Failed to create widget: %v", err)
	}

	// When - Disable
	widget.Disable()

	// Then
	if widget.Enabled() {
		t.Error("Expected widget to be disabled")
	}

	// When - Enable
	widget.Enable()

	// Then
	if !widget.Enabled() {
		t.Error("Expected widget to be enabled")
	}
}

func TestWidget_CreatedAt_UpdatedAt(t *testing.T) {
	// Given
	before := time.Now()

	// When
	widget, err := NewWidget("Test Widget", WeatherCurrent, APIDataSource)
	if err != nil {
		t.Fatalf("Failed to create widget: %v", err)
	}

	after := time.Now()

	// Then
	if widget.CreatedAt().Before(before) || widget.CreatedAt().After(after) {
		t.Errorf("Expected CreatedAt to be between %v and %v, got %v", before, after, widget.CreatedAt())
	}

	if widget.UpdatedAt().Before(before) || widget.UpdatedAt().After(after) {
		t.Errorf("Expected UpdatedAt to be between %v and %v, got %v", before, after, widget.UpdatedAt())
	}
}

func TestTemplateType_IsValid(t *testing.T) {
	// Given
	validTemplateTypes := []TemplateType{
		KeyValue,
		TitleSubtitleValue,
		WeatherCurrent,
		TimeDisplay,
		StatusList,
	}

	invalidTemplateTypes := []TemplateType{
		"",
		"invalid",
		"weather_invalid",
		"custom_template",
	}

	// Test valid template types
	for _, templateType := range validTemplateTypes {
		if !templateType.IsValid() {
			t.Errorf("Expected template type %s to be valid", templateType)
		}
	}

	// Test invalid template types
	for _, templateType := range invalidTemplateTypes {
		if templateType.IsValid() {
			t.Errorf("Expected template type %s to be invalid", templateType)
		}
	}
}

func TestDataSource_IsValid(t *testing.T) {
	// Given
	validDataSources := []DataSource{APIDataSource, RSSDataSource}
	invalidDataSources := []DataSource{"", "database", "file", "custom"}

	// Test valid data sources
	for _, dataSource := range validDataSources {
		if !dataSource.IsValid() {
			t.Errorf("Expected data source %s to be valid", dataSource)
		}
	}

	// Test invalid data sources
	for _, dataSource := range invalidDataSources {
		if dataSource.IsValid() {
			t.Errorf("Expected data source %s to be invalid", dataSource)
		}
	}
}

func TestDomainError_Types(t *testing.T) {
	// Test error type detection
	validationErr := NewDomainError(ErrValidationFailure, "validation failed")
	notFoundErr := NewDomainError(ErrWidgetNotFound, "widget not found")

	if !IsValidationError(validationErr) {
		t.Error("Expected validation error to be detected")
	}

	if !IsNotFoundError(notFoundErr) {
		t.Error("Expected not found error to be detected")
	}

	if IsValidationError(notFoundErr) {
		t.Error("Expected not found error not to be validation error")
	}

	if IsNotFoundError(validationErr) {
		t.Error("Expected validation error not to be not found error")
	}
}

func TestWidget_Validation(t *testing.T) {
	// Given
	widget, err := NewWidget("Test Widget", WeatherCurrent, APIDataSource)
	if err != nil {
		t.Fatalf("Failed to create widget: %v", err)
	}

	// Set up required API configuration for validation
	err = widget.SetAPIConfiguration("https://api.example.com/data", map[string]string{"Authorization": "Bearer test"})
	if err != nil {
		t.Fatalf("Failed to set API configuration: %v", err)
	}

	// When
	err = widget.Validate()

	// Then
	if err != nil {
		t.Errorf("Expected no error for valid widget, got: %v", err)
	}
}
