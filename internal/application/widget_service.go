package application

import (
	"fmt"
	"log"

	"github.com/bartosz/homeboard/internal/api/dto"
	"github.com/bartosz/homeboard/internal/domain/widget"
)

// WidgetService implements the application service for widget operations
type WidgetService struct {
	repo      widget.Repository
	validator WidgetValidator
	logger    Logger
}

// WidgetValidator interface for widget validation
type WidgetValidator interface {
	ValidateWidget(w *widget.Widget) error
	ValidateDataMapping(mapping widget.DataMapping, templateType widget.TemplateType) error
}

// Logger interface for application logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, err error, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// NewWidgetService creates a new widget service
func NewWidgetService(repo widget.Repository, validator WidgetValidator, logger Logger) *WidgetService {
	return &WidgetService{
		repo:      repo,
		validator: validator,
		logger:    logger,
	}
}

// CreateWidget creates a new widget from a DTO
func (s *WidgetService) CreateWidget(req dto.CreateWidgetRequest) (*dto.WidgetResponse, error) {
	s.logger.Info("Creating widget", "name", req.Name, "template_type", req.TemplateType)

	// Convert DTO to domain objects
	templateType := widget.TemplateType(req.TemplateType)
	dataSource := widget.DataSource(req.DataSource)

	// Create domain widget
	w, err := widget.NewWidget(req.Name, templateType, dataSource)
	if err != nil {
		s.logger.Error("Failed to create widget domain object", err, "name", req.Name)
		return nil, err
	}

	// Set configuration based on data source
	if err := s.setWidgetConfiguration(w, req); err != nil {
		s.logger.Error("Failed to set widget configuration", err, "widget_id", w.ID())
		return nil, err
	}

	// Set metadata
	if req.Description != "" {
		complexity := s.determineComplexity(req)
		w.UpdateMetadata(req.Description, "", []string{}, complexity)
	}

	// Validate the widget
	if err := s.validator.ValidateWidget(w); err != nil {
		s.logger.Error("Widget validation failed", err, "widget_id", w.ID())
		return nil, widget.NewDomainErrorWithCause(widget.ErrValidationFailure, "widget validation failed", err)
	}

	// Check for existing widget with same name
	existing, _ := s.repo.FindByName(req.Name)
	if existing != nil {
		return nil, widget.NewDomainError(widget.ErrWidgetAlreadyExists, fmt.Sprintf("widget with name '%s' already exists", req.Name))
	}

	// Save the widget
	if err := s.repo.Save(w); err != nil {
		s.logger.Error("Failed to save widget", err, "widget_id", w.ID())
		return nil, widget.NewDomainErrorWithCause(widget.ErrRepositoryFailure, "failed to save widget", err)
	}

	s.logger.Info("Widget created successfully", "widget_id", w.ID(), "name", w.Name())
	return s.toWidgetResponse(w), nil
}

// GetWidget retrieves a widget by ID
func (s *WidgetService) GetWidget(id widget.WidgetID) (*dto.WidgetResponse, error) {
	s.logger.Debug("Getting widget", "widget_id", id)

	w, err := s.repo.FindByID(id)
	if err != nil {
		s.logger.Error("Failed to find widget", err, "widget_id", id)
		return nil, widget.NewDomainErrorWithCause(widget.ErrRepositoryFailure, "failed to find widget", err)
	}
	if w == nil {
		return nil, widget.NewDomainError(widget.ErrWidgetNotFound, fmt.Sprintf("widget with ID %d not found", id))
	}

	return s.toWidgetResponse(w), nil
}

// UpdateWidget updates an existing widget
func (s *WidgetService) UpdateWidget(id widget.WidgetID, req dto.UpdateWidgetRequest) (*dto.WidgetResponse, error) {
	s.logger.Info("Updating widget", "widget_id", id)

	// Get existing widget
	w, err := s.repo.FindByID(id)
	if err != nil {
		s.logger.Error("Failed to find widget for update", err, "widget_id", id)
		return nil, widget.NewDomainErrorWithCause(widget.ErrRepositoryFailure, "failed to find widget", err)
	}
	if w == nil {
		return nil, widget.NewDomainError(widget.ErrWidgetNotFound, fmt.Sprintf("widget with ID %d not found", id))
	}

	// Apply updates
	if err := s.applyWidgetUpdates(w, req); err != nil {
		s.logger.Error("Failed to apply widget updates", err, "widget_id", id)
		return nil, err
	}

	// Validate updated widget
	if err := s.validator.ValidateWidget(w); err != nil {
		s.logger.Error("Updated widget validation failed", err, "widget_id", id)
		return nil, widget.NewDomainErrorWithCause(widget.ErrValidationFailure, "widget validation failed", err)
	}

	// Save updated widget
	if err := s.repo.Save(w); err != nil {
		s.logger.Error("Failed to save updated widget", err, "widget_id", id)
		return nil, widget.NewDomainErrorWithCause(widget.ErrRepositoryFailure, "failed to save widget", err)
	}

	s.logger.Info("Widget updated successfully", "widget_id", id)
	return s.toWidgetResponse(w), nil
}

// DeleteWidget deletes a widget by ID
func (s *WidgetService) DeleteWidget(id widget.WidgetID) error {
	s.logger.Info("Deleting widget", "widget_id", id)

	// Check if widget exists
	w, err := s.repo.FindByID(id)
	if err != nil {
		s.logger.Error("Failed to find widget for deletion", err, "widget_id", id)
		return widget.NewDomainErrorWithCause(widget.ErrRepositoryFailure, "failed to find widget", err)
	}
	if w == nil {
		return widget.NewDomainError(widget.ErrWidgetNotFound, fmt.Sprintf("widget with ID %d not found", id))
	}

	// Delete the widget
	if err := s.repo.Delete(id); err != nil {
		s.logger.Error("Failed to delete widget", err, "widget_id", id)
		return widget.NewDomainErrorWithCause(widget.ErrRepositoryFailure, "failed to delete widget", err)
	}

	s.logger.Info("Widget deleted successfully", "widget_id", id)
	return nil
}

// ListWidgets returns a paginated list of widgets
func (s *WidgetService) ListWidgets(req dto.PaginationRequest) (*dto.WidgetListResponse, error) {
	s.logger.Debug("Listing widgets", "page", req.Page, "limit", req.Limit)

	// Create domain filter from DTO
	filter := widget.ListFilter{
		Search: req.Search,
		Limit:  req.Limit,
		Offset: (req.Page - 1) * req.Limit,
	}

	// Get widgets from repository
	widgets, err := s.repo.FindAll() // TODO: Update repository to support filtering
	if err != nil {
		s.logger.Error("Failed to list widgets", err)
		return nil, widget.NewDomainErrorWithCause(widget.ErrRepositoryFailure, "failed to list widgets", err)
	}

	// Get total count
	total, err := s.repo.Count()
	if err != nil {
		s.logger.Error("Failed to count widgets", err)
		return nil, widget.NewDomainErrorWithCause(widget.ErrRepositoryFailure, "failed to count widgets", err)
	}

	// Apply client-side filtering and pagination (TODO: move to repository)
	filteredWidgets := s.applyClientSideFilter(widgets, filter)

	// Convert to DTOs
	summaries := make([]dto.WidgetSummaryResponse, len(filteredWidgets))
	for i, w := range filteredWidgets {
		summaries[i] = s.toWidgetSummary(w)
	}

	// Calculate pagination
	totalPages := (total + req.Limit - 1) / req.Limit
	pagination := dto.PaginationResponse{
		Page:       req.Page,
		Limit:      req.Limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}

	return &dto.WidgetListResponse{
		Widgets:    summaries,
		Pagination: pagination,
	}, nil
}

// ValidateWidget validates a widget configuration without saving
func (s *WidgetService) ValidateWidget(req dto.WidgetValidationRequest) (*dto.WidgetValidationResponse, error) {
	s.logger.Debug("Validating widget", "name", req.Name)

	// Convert DTO to domain objects
	templateType := widget.TemplateType(req.TemplateType)
	dataSource := widget.DataSource(req.DataSource)

	// Create temporary widget for validation
	w, err := widget.NewWidget(req.Name, templateType, dataSource)
	if err != nil {
		return &dto.WidgetValidationResponse{
			Valid: false,
			Errors: []dto.ValidationErrorDTO{{
				Field:   "general",
				Message: err.Error(),
				Code:    string(widget.ErrValidationFailure),
			}},
		}, nil
	}

	// Set configuration
	createReq := dto.CreateWidgetRequest{
		Name:         req.Name,
		TemplateType: req.TemplateType,
		DataSource:   req.DataSource,
		APIURL:       req.APIURL,
		APIHeaders:   req.APIHeaders,
		DataMapping:  req.DataMapping,
		RSSConfig:    req.RSSConfig,
	}

	if err := s.setWidgetConfiguration(w, createReq); err != nil {
		return &dto.WidgetValidationResponse{
			Valid: false,
			Errors: []dto.ValidationErrorDTO{{
				Field:   "configuration",
				Message: err.Error(),
				Code:    string(widget.ErrInvalidConfiguration),
			}},
		}, nil
	}

	// Validate
	if err := s.validator.ValidateWidget(w); err != nil {
		return &dto.WidgetValidationResponse{
			Valid: false,
			Errors: []dto.ValidationErrorDTO{{
				Field:   "general",
				Message: err.Error(),
				Code:    string(widget.ErrValidationFailure),
			}},
		}, nil
	}

	return &dto.WidgetValidationResponse{
		Valid: true,
	}, nil
}

// Helper methods

func (s *WidgetService) setWidgetConfiguration(w *widget.Widget, req dto.CreateWidgetRequest) error {
	switch w.DataSource() {
	case widget.APIDataSource:
		if req.APIURL == "" {
			return widget.NewDomainError(widget.ErrMissingConfiguration, "API URL is required for API widgets")
		}
		return w.SetAPIConfiguration(req.APIURL, req.APIHeaders)

	case widget.RSSDataSource:
		if req.RSSConfig == nil {
			return widget.NewDomainError(widget.ErrMissingConfiguration, "RSS configuration is required for RSS widgets")
		}
		rssConfig := widget.RSSConfiguration{
			// Convert DTO to domain object
			// Implementation details...
		}
		return w.SetRSSConfiguration(rssConfig)
	}

	return nil
}

func (s *WidgetService) applyWidgetUpdates(w *widget.Widget, req dto.UpdateWidgetRequest) error {
	if req.Name != nil {
		if err := w.UpdateName(*req.Name); err != nil {
			return err
		}
	}

	if req.Enabled != nil {
		if *req.Enabled {
			w.Enable()
		} else {
			w.Disable()
		}
	}

	// Apply other updates...
	return nil
}

func (s *WidgetService) toWidgetResponse(w *widget.Widget) *dto.WidgetResponse {
	return &dto.WidgetResponse{
		ID:           int(w.ID()),
		Name:         w.Name(),
		TemplateType: string(w.TemplateType()),
		DataSource:   string(w.DataSource()),
		Enabled:      w.Enabled(),
		CreatedAt:    w.CreatedAt(),
		UpdatedAt:    w.UpdatedAt(),
		// Add other fields...
	}
}

func (s *WidgetService) toWidgetSummary(w *widget.Widget) dto.WidgetSummaryResponse {
	return dto.WidgetSummaryResponse{
		ID:           int(w.ID()),
		Name:         w.Name(),
		TemplateType: string(w.TemplateType()),
		DataSource:   string(w.DataSource()),
		Enabled:      w.Enabled(),
		CreatedAt:    w.CreatedAt(),
	}
}

func (s *WidgetService) determineComplexity(req dto.CreateWidgetRequest) widget.ComplexityLevel {
	// Simple heuristic for complexity
	if req.DataMapping.Fields != nil && len(req.DataMapping.Fields) > 5 {
		return widget.ComplexComplexity
	}
	if req.DataMapping.Fields != nil && len(req.DataMapping.Fields) > 2 {
		return widget.MediumComplexity
	}
	return widget.SimpleComplexity
}

func (s *WidgetService) applyClientSideFilter(widgets []*widget.Widget, filter widget.ListFilter) []*widget.Widget {
	// TODO: Move this logic to repository layer
	var filtered []*widget.Widget

	for _, w := range widgets {
		// Apply search filter
		if filter.Search != "" {
			// Simple search implementation
			continue
		}

		// Apply other filters...
		filtered = append(filtered, w)
	}

	// Apply pagination
	start := filter.Offset
	end := start + filter.Limit
	if start >= len(filtered) {
		return []*widget.Widget{}
	}
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[start:end]
}

// SimpleLogger provides a basic logger implementation
type SimpleLogger struct{}

func (l *SimpleLogger) Info(msg string, fields ...interface{}) {
	log.Printf("INFO: %s %v", msg, fields)
}

func (l *SimpleLogger) Error(msg string, err error, fields ...interface{}) {
	log.Printf("ERROR: %s: %v %v", msg, err, fields)
}

func (l *SimpleLogger) Debug(msg string, fields ...interface{}) {
	log.Printf("DEBUG: %s %v", msg, fields)
}
