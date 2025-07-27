package dto

import (
	"time"
)

// CreateWidgetRequest represents a request to create a new widget
type CreateWidgetRequest struct {
	Name         string            `json:"name" validate:"required,min=1,max=255"`
	TemplateType string            `json:"template_type" validate:"required,oneof=key_value title_subtitle_value icon_list metric_grid weather_current time_display status_list chart_simple text_block image_caption"`
	DataSource   string            `json:"data_source" validate:"required,oneof=api rss"`
	APIURL       string            `json:"api_url,omitempty" validate:"omitempty,url"`
	APIHeaders   map[string]string `json:"api_headers,omitempty"`
	DataMapping  DataMappingDTO    `json:"data_mapping,omitempty"`
	RSSConfig    *RSSConfigDTO     `json:"rss_config,omitempty"`
	Description  string            `json:"description,omitempty" validate:"max=500"`
	Timeout      int               `json:"timeout,omitempty" validate:"min=1,max=300"`
}

// UpdateWidgetRequest represents a request to update an existing widget
type UpdateWidgetRequest struct {
	Name         *string           `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	TemplateType *string           `json:"template_type,omitempty" validate:"omitempty,oneof=key_value title_subtitle_value icon_list metric_grid weather_current time_display status_list chart_simple text_block image_caption"`
	DataSource   *string           `json:"data_source,omitempty" validate:"omitempty,oneof=api rss"`
	APIURL       *string           `json:"api_url,omitempty" validate:"omitempty,url"`
	APIHeaders   map[string]string `json:"api_headers,omitempty"`
	DataMapping  *DataMappingDTO   `json:"data_mapping,omitempty"`
	RSSConfig    *RSSConfigDTO     `json:"rss_config,omitempty"`
	Description  *string           `json:"description,omitempty" validate:"omitempty,max=500"`
	Timeout      *int              `json:"timeout,omitempty" validate:"omitempty,min=1,max=300"`
	Enabled      *bool             `json:"enabled,omitempty"`
}

// WidgetResponse represents the response when returning widget data
type WidgetResponse struct {
	ID           int               `json:"id"`
	Name         string            `json:"name"`
	TemplateType string            `json:"template_type"`
	DataSource   string            `json:"data_source"`
	APIURL       string            `json:"api_url,omitempty"`
	APIHeaders   map[string]string `json:"api_headers,omitempty"`
	DataMapping  DataMappingDTO    `json:"data_mapping,omitempty"`
	RSSConfig    *RSSConfigDTO     `json:"rss_config,omitempty"`
	Description  string            `json:"description,omitempty"`
	Timeout      int               `json:"timeout"`
	Enabled      bool              `json:"enabled"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// WidgetListResponse represents a paginated list of widgets
type WidgetListResponse struct {
	Widgets    []WidgetSummaryResponse `json:"widgets"`
	Pagination PaginationResponse      `json:"pagination"`
}

// WidgetSummaryResponse represents a simplified widget for list views
type WidgetSummaryResponse struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	TemplateType string    `json:"template_type"`
	DataSource   string    `json:"data_source"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
}

// DataMappingDTO represents data mapping configuration
type DataMappingDTO struct {
	Fields map[string]FieldMappingDTO `json:"fields"`
}

// FieldMappingDTO represents mapping for a single field
type FieldMappingDTO struct {
	JSONPath   string      `json:"json_path" validate:"required"`
	Type       string      `json:"type" validate:"required,oneof=string number boolean array object"`
	Required   bool        `json:"required"`
	Default    interface{} `json:"default,omitempty"`
	Transform  string      `json:"transform,omitempty"`
	Validation string      `json:"validation,omitempty"`
}

// RSSConfigDTO represents RSS feed configuration
type RSSConfigDTO struct {
	FeedURL       string `json:"feed_url" validate:"required,url"`
	MaxItems      int    `json:"max_items" validate:"min=1,max=100"`
	CacheMinutes  int    `json:"cache_minutes" validate:"min=1,max=1440"`
	ItemFilter    string `json:"item_filter,omitempty" validate:"omitempty,oneof=latest today week"`
	IncludeImage  bool   `json:"include_image"`
	IncludeAuthor bool   `json:"include_author"`
	DateFormat    string `json:"date_format,omitempty"`
}

// WidgetPreviewRequest represents a request to preview widget data
type WidgetPreviewRequest struct {
	TemplateType string            `json:"template_type" validate:"required"`
	APIURL       string            `json:"api_url,omitempty" validate:"omitempty,url"`
	APIHeaders   map[string]string `json:"api_headers,omitempty"`
	DataMapping  DataMappingDTO    `json:"data_mapping,omitempty"`
	RSSConfig    *RSSConfigDTO     `json:"rss_config,omitempty"`
}

// WidgetPreviewResponse represents the response for widget preview
type WidgetPreviewResponse struct {
	Success      bool        `json:"success"`
	Data         interface{} `json:"data,omitempty"`
	RenderedHTML string      `json:"rendered_html,omitempty"`
	Error        *ErrorDTO   `json:"error,omitempty"`
	Warnings     []string    `json:"warnings,omitempty"`
}

// WidgetValidationRequest represents a request to validate widget configuration
type WidgetValidationRequest struct {
	Name         string            `json:"name"`
	TemplateType string            `json:"template_type"`
	DataSource   string            `json:"data_source"`
	APIURL       string            `json:"api_url,omitempty"`
	APIHeaders   map[string]string `json:"api_headers,omitempty"`
	DataMapping  DataMappingDTO    `json:"data_mapping,omitempty"`
	RSSConfig    *RSSConfigDTO     `json:"rss_config,omitempty"`
}

// WidgetValidationResponse represents the response for widget validation
type WidgetValidationResponse struct {
	Valid    bool                   `json:"valid"`
	Errors   []ValidationErrorDTO   `json:"errors,omitempty"`
	Warnings []ValidationWarningDTO `json:"warnings,omitempty"`
}

// ValidationErrorDTO represents a validation error
type ValidationErrorDTO struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidationWarningDTO represents a validation warning
type ValidationWarningDTO struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Impact  string `json:"impact"` // low, medium, high
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page    int    `json:"page" validate:"min=1"`
	Limit   int    `json:"limit" validate:"min=1,max=100"`
	SortBy  string `json:"sort_by,omitempty" validate:"omitempty,oneof=id name created_at updated_at"`
	SortDir string `json:"sort_dir,omitempty" validate:"omitempty,oneof=asc desc"`
	Search  string `json:"search,omitempty" validate:"omitempty,max=255"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// Common response DTOs

// ErrorDTO represents an error response
type ErrorDTO struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessDTO represents a success response
type SuccessDTO struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
