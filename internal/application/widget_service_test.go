package application

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/bartosz/homeboard/internal/api/dto"
	"github.com/bartosz/homeboard/internal/domain/widget"
)

// MockWidgetRepository implements widget.Repository for testing
type MockWidgetRepository struct {
	widgets map[widget.WidgetID]*widget.Widget
	nextID  widget.WidgetID
}

func NewMockWidgetRepository() *MockWidgetRepository {
	return &MockWidgetRepository{
		widgets: make(map[widget.WidgetID]*widget.Widget),
		nextID:  1,
	}
}

func (r *MockWidgetRepository) Save(w *widget.Widget) error {
	// Set the widget ID using reflection
	newID := r.nextID
	r.setWidgetID(w, newID)

	// Store widget with generated ID
	r.widgets[newID] = w
	r.nextID++

	return nil
}

// setWidgetID sets the private id field using reflection
func (r *MockWidgetRepository) setWidgetID(w *widget.Widget, id widget.WidgetID) {
	v := reflect.ValueOf(w).Elem()
	idField := v.FieldByName("id")

	if idField.IsValid() && idField.CanAddr() {
		// Make the field writable using unsafe
		idField = reflect.NewAt(idField.Type(), unsafe.Pointer(idField.UnsafeAddr())).Elem()
		idField.Set(reflect.ValueOf(id))
	}
}

func (r *MockWidgetRepository) FindByID(id widget.WidgetID) (*widget.Widget, error) {
	w, exists := r.widgets[id]
	if !exists {
		return nil, nil
	}
	return w, nil
}

func (r *MockWidgetRepository) FindAll() ([]*widget.Widget, error) {
	widgets := make([]*widget.Widget, 0, len(r.widgets))
	for _, w := range r.widgets {
		widgets = append(widgets, w)
	}
	return widgets, nil
}

func (r *MockWidgetRepository) FindByName(name string) (*widget.Widget, error) {
	for _, w := range r.widgets {
		if w.Name() == name {
			return w, nil
		}
	}
	return nil, nil
}

func (r *MockWidgetRepository) Delete(id widget.WidgetID) error {
	delete(r.widgets, id)
	return nil
}

func (r *MockWidgetRepository) Count() (int, error) {
	return len(r.widgets), nil
}

// MockWidgetValidator implements WidgetValidator for testing
type MockWidgetValidator struct {
	shouldFailValidation bool
}

func NewMockWidgetValidator() *MockWidgetValidator {
	return &MockWidgetValidator{shouldFailValidation: false}
}

func (v *MockWidgetValidator) SetShouldFail(fail bool) {
	v.shouldFailValidation = fail
}

func (v *MockWidgetValidator) ValidateWidget(w *widget.Widget) error {
	if v.shouldFailValidation {
		return widget.NewDomainError(widget.ErrValidationFailure, "mock validation failure")
	}
	return w.Validate()
}

func (v *MockWidgetValidator) ValidateDataMapping(mapping widget.DataMapping, templateType widget.TemplateType) error {
	if v.shouldFailValidation {
		return widget.NewDomainError(widget.ErrInvalidFieldMapping, "mock mapping validation failure")
	}
	return nil
}

// Test fixtures
func createTestWidgetRequest() dto.CreateWidgetRequest {
	return dto.CreateWidgetRequest{
		Name:         "Test Weather Widget",
		TemplateType: "weather_current",
		DataSource:   "api",
		APIURL:       "https://api.openweathermap.org/data/2.5/weather",
		APIHeaders:   map[string]string{"Authorization": "Bearer test-token"},
		Description:  "A test weather widget for unit testing",
		Timeout:      30,
		DataMapping: dto.DataMappingDTO{
			Fields: map[string]dto.FieldMappingDTO{
				"temperature": {
					JSONPath: "$.main.temp",
					Type:     "number",
					Required: true,
				},
				"condition": {
					JSONPath: "$.weather[0].description",
					Type:     "string",
					Required: true,
				},
			},
		},
	}
}

func createTestService() (*WidgetService, *MockWidgetRepository, *MockWidgetValidator) {
	repo := NewMockWidgetRepository()
	validator := NewMockWidgetValidator()
	logger := &SimpleLogger{}
	service := NewWidgetService(repo, validator, logger)
	return service, repo, validator
}

func TestWidgetService_CreateWidget_Success(t *testing.T) {
	// Given
	service, repo, _ := createTestService()
	request := createTestWidgetRequest()

	// When
	response, err := service.CreateWidget(request)

	// Then
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Name != request.Name {
		t.Errorf("Expected name %s, got %s", request.Name, response.Name)
	}

	if response.TemplateType != request.TemplateType {
		t.Errorf("Expected template type %s, got %s", request.TemplateType, response.TemplateType)
	}

	if response.DataSource != request.DataSource {
		t.Errorf("Expected data source %s, got %s", request.DataSource, response.DataSource)
	}

	if response.Enabled != true {
		t.Error("Expected widget to be enabled by default")
	}

	// Verify widget was saved to repository
	widgets, err := repo.FindAll()
	if err != nil {
		t.Fatalf("Error finding widgets: %v", err)
	}

	if len(widgets) != 1 {
		t.Errorf("Expected 1 widget in repository, got %d", len(widgets))
	}
}

func TestWidgetService_CreateWidget_ValidationFailure(t *testing.T) {
	// Given
	service, _, validator := createTestService()
	request := createTestWidgetRequest()
	validator.SetShouldFail(true)

	// When
	response, err := service.CreateWidget(request)

	// Then
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	if response != nil {
		t.Errorf("Expected nil response, got: %v", response)
	}

	if !widget.IsValidationError(err) {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

func TestWidgetService_CreateWidget_DuplicateName(t *testing.T) {
	// Given
	service, _, _ := createTestService()
	request := createTestWidgetRequest()

	// Create first widget
	_, err := service.CreateWidget(request)
	if err != nil {
		t.Fatalf("Failed to create first widget: %v", err)
	}

	// When - Try to create second widget with same name
	response, err := service.CreateWidget(request)

	// Then
	if err == nil {
		t.Fatal("Expected duplicate name error, got nil")
	}

	if response != nil {
		t.Errorf("Expected nil response, got: %v", response)
	}

	domainErr, ok := err.(*widget.DomainError)
	if !ok {
		t.Fatalf("Expected DomainError, got: %T", err)
	}

	if domainErr.Code != widget.ErrWidgetAlreadyExists {
		t.Errorf("Expected ErrWidgetAlreadyExists, got: %s", domainErr.Code)
	}
}

func TestWidgetService_CreateWidget_InvalidTemplateType(t *testing.T) {
	// Given
	service, _, _ := createTestService()
	request := createTestWidgetRequest()
	request.TemplateType = "invalid_template"

	// When
	response, err := service.CreateWidget(request)

	// Then
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	if response != nil {
		t.Errorf("Expected nil response, got: %v", response)
	}

	if !widget.IsValidationError(err) {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

func TestWidgetService_GetWidget_Success(t *testing.T) {
	// Given
	service, _, _ := createTestService()
	request := createTestWidgetRequest()

	// Create widget first
	createResponse, err := service.CreateWidget(request)
	if err != nil {
		t.Fatalf("Failed to create widget: %v", err)
	}

	// When
	getResponse, err := service.GetWidget(widget.WidgetID(createResponse.ID))

	// Then
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if getResponse == nil {
		t.Fatal("Expected response, got nil")
	}

	if getResponse.ID != createResponse.ID {
		t.Errorf("Expected ID %d, got %d", createResponse.ID, getResponse.ID)
	}

	if getResponse.Name != createResponse.Name {
		t.Errorf("Expected name %s, got %s", createResponse.Name, getResponse.Name)
	}
}

func TestWidgetService_GetWidget_NotFound(t *testing.T) {
	// Given
	service, _, _ := createTestService()
	nonExistentID := widget.WidgetID(999)

	// When
	response, err := service.GetWidget(nonExistentID)

	// Then
	if err == nil {
		t.Fatal("Expected not found error, got nil")
	}

	if response != nil {
		t.Errorf("Expected nil response, got: %v", response)
	}

	if !widget.IsNotFoundError(err) {
		t.Errorf("Expected not found error, got: %v", err)
	}
}

func TestWidgetService_UpdateWidget_Success(t *testing.T) {
	// Given
	service, _, _ := createTestService()
	request := createTestWidgetRequest()

	// Create widget first
	createResponse, err := service.CreateWidget(request)
	if err != nil {
		t.Fatalf("Failed to create widget: %v", err)
	}

	// Prepare update request
	newName := "Updated Weather Widget"
	enabled := false
	updateRequest := dto.UpdateWidgetRequest{
		Name:    &newName,
		Enabled: &enabled,
	}

	// When
	updateResponse, err := service.UpdateWidget(widget.WidgetID(createResponse.ID), updateRequest)

	// Then
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if updateResponse == nil {
		t.Fatal("Expected response, got nil")
	}

	if updateResponse.Name != newName {
		t.Errorf("Expected name %s, got %s", newName, updateResponse.Name)
	}

	if updateResponse.Enabled != enabled {
		t.Errorf("Expected enabled %v, got %v", enabled, updateResponse.Enabled)
	}
}

func TestWidgetService_UpdateWidget_NotFound(t *testing.T) {
	// Given
	service, _, _ := createTestService()
	nonExistentID := widget.WidgetID(999)
	newName := "Updated Widget"
	updateRequest := dto.UpdateWidgetRequest{
		Name: &newName,
	}

	// When
	response, err := service.UpdateWidget(nonExistentID, updateRequest)

	// Then
	if err == nil {
		t.Fatal("Expected not found error, got nil")
	}

	if response != nil {
		t.Errorf("Expected nil response, got: %v", response)
	}

	if !widget.IsNotFoundError(err) {
		t.Errorf("Expected not found error, got: %v", err)
	}
}

func TestWidgetService_DeleteWidget_Success(t *testing.T) {
	// Given
	service, repo, _ := createTestService()
	request := createTestWidgetRequest()

	// Create widget first
	createResponse, err := service.CreateWidget(request)
	if err != nil {
		t.Fatalf("Failed to create widget: %v", err)
	}

	// When
	err = service.DeleteWidget(widget.WidgetID(createResponse.ID))

	// Then
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify widget was deleted from repository
	widgets, err := repo.FindAll()
	if err != nil {
		t.Fatalf("Error finding widgets: %v", err)
	}

	if len(widgets) != 0 {
		t.Errorf("Expected 0 widgets in repository, got %d", len(widgets))
	}
}

func TestWidgetService_DeleteWidget_NotFound(t *testing.T) {
	// Given
	service, _, _ := createTestService()
	nonExistentID := widget.WidgetID(999)

	// When
	err := service.DeleteWidget(nonExistentID)

	// Then
	if err == nil {
		t.Fatal("Expected not found error, got nil")
	}

	if !widget.IsNotFoundError(err) {
		t.Errorf("Expected not found error, got: %v", err)
	}
}

func TestWidgetService_ListWidgets_Success(t *testing.T) {
	// Given
	service, _, _ := createTestService()

	// Create multiple widgets
	request1 := createTestWidgetRequest()
	request1.Name = "Widget 1"

	request2 := createTestWidgetRequest()
	request2.Name = "Widget 2"

	_, err := service.CreateWidget(request1)
	if err != nil {
		t.Fatalf("Failed to create widget 1: %v", err)
	}

	_, err = service.CreateWidget(request2)
	if err != nil {
		t.Fatalf("Failed to create widget 2: %v", err)
	}

	paginationRequest := dto.PaginationRequest{
		Page:  1,
		Limit: 10,
	}

	// When
	response, err := service.ListWidgets(paginationRequest)

	// Then
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(response.Widgets) != 2 {
		t.Errorf("Expected 2 widgets, got %d", len(response.Widgets))
	}

	if response.Pagination.Total != 2 {
		t.Errorf("Expected total 2, got %d", response.Pagination.Total)
	}

	if response.Pagination.Page != 1 {
		t.Errorf("Expected page 1, got %d", response.Pagination.Page)
	}
}

func TestWidgetService_ValidateWidget_Success(t *testing.T) {
	// Given
	service, _, _ := createTestService()
	request := dto.WidgetValidationRequest{
		Name:         "Valid Widget",
		TemplateType: "weather_current",
		DataSource:   "api",
		APIURL:       "https://api.example.com/data",
		APIHeaders:   map[string]string{"Authorization": "Bearer token"},
		DataMapping: dto.DataMappingDTO{
			Fields: map[string]dto.FieldMappingDTO{
				"temperature": {
					JSONPath: "$.temp",
					Type:     "number",
					Required: true,
				},
			},
		},
	}

	// When
	response, err := service.ValidateWidget(request)

	// Then
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if !response.Valid {
		t.Error("Expected valid widget")
	}

	if len(response.Errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(response.Errors))
	}
}

func TestWidgetService_ValidateWidget_ValidationFailure(t *testing.T) {
	// Given
	service, _, validator := createTestService()
	validator.SetShouldFail(true)

	request := dto.WidgetValidationRequest{
		Name:         "Invalid Widget",
		TemplateType: "weather_current",
		DataSource:   "api",
		APIURL:       "https://api.example.com/data",
	}

	// When
	response, err := service.ValidateWidget(request)

	// Then
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Valid {
		t.Error("Expected invalid widget")
	}

	if len(response.Errors) == 0 {
		t.Error("Expected validation errors")
	}
}

func TestWidgetService_ValidateWidget_InvalidName(t *testing.T) {
	// Given
	service, _, _ := createTestService()
	request := dto.WidgetValidationRequest{
		Name:         "", // Invalid: empty name
		TemplateType: "weather_current",
		DataSource:   "api",
		APIURL:       "https://api.example.com/data",
	}

	// When
	response, err := service.ValidateWidget(request)

	// Then
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Valid {
		t.Error("Expected invalid widget")
	}

	if len(response.Errors) == 0 {
		t.Error("Expected validation errors")
	}

	// Check error details
	found := false
	for _, errDto := range response.Errors {
		if errDto.Code == string(widget.ErrValidationFailure) {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected validation failure error in response")
	}
}

// Benchmark tests
func BenchmarkWidgetService_CreateWidget(b *testing.B) {
	service, _, _ := createTestService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := createTestWidgetRequest()
		request.Name = fmt.Sprintf("Benchmark Widget %d", i)

		_, err := service.CreateWidget(request)
		if err != nil {
			b.Fatalf("Failed to create widget: %v", err)
		}
	}
}

func BenchmarkWidgetService_GetWidget(b *testing.B) {
	service, _, _ := createTestService()

	// Create a widget to get
	request := createTestWidgetRequest()
	response, err := service.CreateWidget(request)
	if err != nil {
		b.Fatalf("Failed to create widget: %v", err)
	}

	widgetID := widget.WidgetID(response.ID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GetWidget(widgetID)
		if err != nil {
			b.Fatalf("Failed to get widget: %v", err)
		}
	}
}
