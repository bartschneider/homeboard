package application

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/bartosz/homeboard/internal/api/dto"
	"github.com/bartosz/homeboard/internal/domain/widget"
)

// ImprovedMockWidgetRepository implements widget.Repository with proper ID management
type ImprovedMockWidgetRepository struct {
	widgets   map[widget.WidgetID]*widget.Widget
	widgetIDs map[*widget.Widget]widget.WidgetID
	nextID    widget.WidgetID
}

func NewImprovedMockWidgetRepository() *ImprovedMockWidgetRepository {
	return &ImprovedMockWidgetRepository{
		widgets:   make(map[widget.WidgetID]*widget.Widget),
		widgetIDs: make(map[*widget.Widget]widget.WidgetID),
		nextID:    1,
	}
}

func (r *ImprovedMockWidgetRepository) Save(w *widget.Widget) error {
	// Check if widget already has an ID
	if existingID, exists := r.widgetIDs[w]; exists {
		// Update existing widget
		r.widgets[existingID] = w
		return nil
	}

	// Assign new ID using reflection to set the private field
	newID := r.nextID
	r.nextID++

	// Use reflection to set the widget ID (since it's a private field)
	// This simulates what a real database would do
	w = r.setWidgetID(w, newID)

	r.widgets[newID] = w
	r.widgetIDs[w] = newID

	return nil
}

// setWidgetID simulates setting the ID that would be done by a database
func (r *ImprovedMockWidgetRepository) setWidgetID(w *widget.Widget, id widget.WidgetID) *widget.Widget {
	// Use reflection to set the private id field
	v := reflect.ValueOf(w).Elem()
	idField := v.FieldByName("id")

	if idField.IsValid() && idField.CanAddr() {
		// Make the field writable using unsafe
		idField = reflect.NewAt(idField.Type(), unsafe.Pointer(idField.UnsafeAddr())).Elem()
		idField.Set(reflect.ValueOf(id))
	}

	return w
}

func (r *ImprovedMockWidgetRepository) FindByID(id widget.WidgetID) (*widget.Widget, error) {
	w, exists := r.widgets[id]
	if !exists {
		return nil, widget.NewDomainError(widget.ErrWidgetNotFound, fmt.Sprintf("widget with ID %d not found", id))
	}
	return w, nil
}

func (r *ImprovedMockWidgetRepository) FindAll() ([]*widget.Widget, error) {
	widgets := make([]*widget.Widget, 0, len(r.widgets))
	for _, w := range r.widgets {
		widgets = append(widgets, w)
	}
	return widgets, nil
}

func (r *ImprovedMockWidgetRepository) FindByName(name string) (*widget.Widget, error) {
	for _, w := range r.widgets {
		if w.Name() == name {
			return w, nil
		}
	}
	return nil, nil
}

func (r *ImprovedMockWidgetRepository) Delete(id widget.WidgetID) error {
	w, exists := r.widgets[id]
	if !exists {
		return widget.NewDomainError(widget.ErrWidgetNotFound, fmt.Sprintf("widget with ID %d not found", id))
	}

	delete(r.widgets, id)
	delete(r.widgetIDs, w)
	return nil
}

func (r *ImprovedMockWidgetRepository) Count() (int, error) {
	return len(r.widgets), nil
}

func (r *ImprovedMockWidgetRepository) GetWidgetID(w *widget.Widget) widget.WidgetID {
	if id, exists := r.widgetIDs[w]; exists {
		return id
	}
	return 0
}

// Test helper functions
func createImprovedTestService() (*WidgetService, *ImprovedMockWidgetRepository, *MockWidgetValidator) {
	repo := NewImprovedMockWidgetRepository()
	validator := NewMockWidgetValidator()
	logger := &SimpleLogger{}
	service := NewWidgetService(repo, validator, logger)
	return service, repo, validator
}

// Improved tests
func TestImprovedWidgetService_CreateWidget_Success(t *testing.T) {
	// Given
	service, repo, _ := createImprovedTestService()
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

	// Verify widget was saved to repository
	widgets, err := repo.FindAll()
	if err != nil {
		t.Fatalf("Error finding widgets: %v", err)
	}

	if len(widgets) != 1 {
		t.Errorf("Expected 1 widget in repository, got %d", len(widgets))
	}

	// Verify we can retrieve the widget by ID
	retrievedWidget, err := repo.FindByID(widget.WidgetID(response.ID))
	if err != nil {
		t.Fatalf("Error retrieving widget: %v", err)
	}

	if retrievedWidget.Name() != request.Name {
		t.Errorf("Expected retrieved widget name %s, got %s", request.Name, retrievedWidget.Name())
	}
}

func TestImprovedWidgetService_GetWidget_Success(t *testing.T) {
	// Given
	service, _, _ := createImprovedTestService()
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

func TestImprovedWidgetService_UpdateWidget_Success(t *testing.T) {
	// Given
	service, _, _ := createImprovedTestService()
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

func TestImprovedWidgetService_DeleteWidget_Success(t *testing.T) {
	// Given
	service, repo, _ := createImprovedTestService()
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

func TestImprovedWidgetService_ListWidgets_WithMultipleWidgets(t *testing.T) {
	// Given
	service, _, _ := createImprovedTestService()

	// Create multiple widgets
	request1 := createTestWidgetRequest()
	request1.Name = "Widget 1"

	request2 := createTestWidgetRequest()
	request2.Name = "Widget 2"

	request3 := createTestWidgetRequest()
	request3.Name = "Widget 3"

	_, err := service.CreateWidget(request1)
	if err != nil {
		t.Fatalf("Failed to create widget 1: %v", err)
	}

	_, err = service.CreateWidget(request2)
	if err != nil {
		t.Fatalf("Failed to create widget 2: %v", err)
	}

	_, err = service.CreateWidget(request3)
	if err != nil {
		t.Fatalf("Failed to create widget 3: %v", err)
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

	if len(response.Widgets) != 3 {
		t.Errorf("Expected 3 widgets, got %d", len(response.Widgets))
	}

	if response.Pagination.Total != 3 {
		t.Errorf("Expected total 3, got %d", response.Pagination.Total)
	}

	if response.Pagination.Page != 1 {
		t.Errorf("Expected page 1, got %d", response.Pagination.Page)
	}
}

func TestImprovedWidgetService_ErrorHandling_NotFound(t *testing.T) {
	// Given
	service, _, _ := createImprovedTestService()
	nonExistentID := widget.WidgetID(999)

	// When - Try to get non-existent widget
	_, err := service.GetWidget(nonExistentID)

	// Then
	if err == nil {
		t.Fatal("Expected not found error, got nil")
	}

	// Check if it's a domain error with the right code
	domainErr, ok := err.(*widget.DomainError)
	if !ok || domainErr.Code != widget.ErrRepositoryFailure {
		t.Errorf("Expected repository failure error, got: %v", err)
	}
}

func TestImprovedWidgetService_DuplicateName_Handling(t *testing.T) {
	// Given
	service, _, _ := createImprovedTestService()
	request := createTestWidgetRequest()

	// Create first widget
	_, err := service.CreateWidget(request)
	if err != nil {
		t.Fatalf("Failed to create first widget: %v", err)
	}

	// When - Try to create second widget with same name
	_, err = service.CreateWidget(request)

	// Then
	if err == nil {
		t.Fatal("Expected duplicate name error, got nil")
	}

	domainErr, ok := err.(*widget.DomainError)
	if !ok {
		t.Fatalf("Expected DomainError, got: %T", err)
	}

	if domainErr.Code != widget.ErrWidgetAlreadyExists {
		t.Errorf("Expected ErrWidgetAlreadyExists, got: %s", domainErr.Code)
	}
}

// Benchmark tests for performance validation
func BenchmarkImprovedWidgetService_CreateWidget(b *testing.B) {
	service, _, _ := createImprovedTestService()

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

func BenchmarkImprovedWidgetService_GetWidget(b *testing.B) {
	service, _, _ := createImprovedTestService()

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
