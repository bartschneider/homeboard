package application

import (
	"testing"

	"github.com/bartosz/homeboard/internal/api/dto"
)

func TestWidgetService_BasicFunctionality(t *testing.T) {
	// Test widget service creation and basic operations
	service, _, _ := createTestService()

	if service == nil {
		t.Fatal("Expected widget service to be created")
	}

	// Test validation functionality
	request := dto.WidgetValidationRequest{
		Name:         "Test Widget",
		TemplateType: "weather_current",
		DataSource:   "api",
		APIURL:       "https://api.example.com",
	}

	response, err := service.ValidateWidget(request)
	if err != nil {
		t.Fatalf("Expected no error during validation, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected validation response")
	}

	if !response.Valid {
		t.Error("Expected valid widget configuration")
	}
}

func TestWidgetService_ValidationErrors(t *testing.T) {
	service, _, validator := createTestService()
	validator.SetShouldFail(true)

	request := dto.WidgetValidationRequest{
		Name:         "Invalid Widget",
		TemplateType: "invalid_type",
		DataSource:   "api",
		APIURL:       "invalid-url",
	}

	response, err := service.ValidateWidget(request)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response.Valid {
		t.Error("Expected invalid widget")
	}

	if len(response.Errors) == 0 {
		t.Error("Expected validation errors")
	}
}

func TestWidgetService_ListEmptyWidgets(t *testing.T) {
	service, _, _ := createTestService()

	pagination := dto.PaginationRequest{
		Page:  1,
		Limit: 10,
	}

	response, err := service.ListWidgets(pagination)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response")
	}

	if len(response.Widgets) != 0 {
		t.Errorf("Expected empty widget list, got %d widgets", len(response.Widgets))
	}

	if response.Pagination.Total != 0 {
		t.Errorf("Expected total 0, got %d", response.Pagination.Total)
	}
}
