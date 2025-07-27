package widget

import (
	"fmt"
	"time"
)

// WidgetID represents a unique widget identifier
type WidgetID int

// Widget represents the core widget domain model
type Widget struct {
	id            WidgetID
	name          string
	templateType  TemplateType
	dataSource    DataSource
	configuration Configuration
	metadata      Metadata
	enabled       bool
	createdAt     time.Time
	updatedAt     time.Time
}

// TemplateType represents widget template types
type TemplateType string

const (
	KeyValue           TemplateType = "key_value"
	TitleSubtitleValue TemplateType = "title_subtitle_value"
	IconList           TemplateType = "icon_list"
	MetricGrid         TemplateType = "metric_grid"
	WeatherCurrent     TemplateType = "weather_current"
	TimeDisplay        TemplateType = "time_display"
	StatusList         TemplateType = "status_list"
	ChartSimple        TemplateType = "chart_simple"
	TextBlock          TemplateType = "text_block"
	ImageCaption       TemplateType = "image_caption"
)

// DataSource represents widget data source types
type DataSource string

const (
	APIDataSource DataSource = "api"
	RSSDataSource DataSource = "rss"
)

// IsValid checks if the TemplateType is valid
func (t TemplateType) IsValid() bool {
	switch t {
	case KeyValue, TitleSubtitleValue, IconList, MetricGrid, WeatherCurrent,
		TimeDisplay, StatusList, ChartSimple, TextBlock, ImageCaption:
		return true
	default:
		return false
	}
}

// IsValid checks if the DataSource is valid
func (d DataSource) IsValid() bool {
	switch d {
	case APIDataSource, RSSDataSource:
		return true
	default:
		return false
	}
}

// Configuration represents widget configuration
type Configuration struct {
	apiConfig *APIConfiguration
	rssConfig *RSSConfiguration
	mapping   DataMapping
	timeout   time.Duration
}

// APIConfiguration represents API-specific configuration
type APIConfiguration struct {
	url     string
	headers map[string]string
	method  string
}

// RSSConfiguration represents RSS-specific configuration
type RSSConfiguration struct {
	feedURL       string
	maxItems      int
	cacheMinutes  int
	itemFilter    string
	includeImage  bool
	includeAuthor bool
	dateFormat    string
}

// DataMapping represents field mapping configuration
type DataMapping struct {
	fields map[string]FieldMapping
}

// FieldMapping represents mapping for a single field
type FieldMapping struct {
	jsonPath   string
	fieldType  FieldType
	required   bool
	defaultVal interface{}
	transform  string
	validation string
}

// FieldType represents field data types
type FieldType string

const (
	StringField  FieldType = "string"
	NumberField  FieldType = "number"
	BooleanField FieldType = "boolean"
	ArrayField   FieldType = "array"
	ObjectField  FieldType = "object"
)

// Metadata represents widget metadata
type Metadata struct {
	description string
	category    string
	tags        []string
	complexity  ComplexityLevel
}

// ComplexityLevel represents widget complexity
type ComplexityLevel string

const (
	SimpleComplexity  ComplexityLevel = "simple"
	MediumComplexity  ComplexityLevel = "medium"
	ComplexComplexity ComplexityLevel = "complex"
)

// Constructor
func NewWidget(name string, templateType TemplateType, dataSource DataSource) (*Widget, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}

	if err := validateTemplateType(templateType); err != nil {
		return nil, err
	}

	if err := validateDataSource(dataSource); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Widget{
		name:         name,
		templateType: templateType,
		dataSource:   dataSource,
		configuration: Configuration{
			timeout: 30 * time.Second,
		},
		enabled:   true,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// Getters
func (w *Widget) ID() WidgetID                 { return w.id }
func (w *Widget) Name() string                 { return w.name }
func (w *Widget) TemplateType() TemplateType   { return w.templateType }
func (w *Widget) DataSource() DataSource       { return w.dataSource }
func (w *Widget) Configuration() Configuration { return w.configuration }
func (w *Widget) Metadata() Metadata           { return w.metadata }
func (w *Widget) Enabled() bool                { return w.enabled }
func (w *Widget) CreatedAt() time.Time         { return w.createdAt }
func (w *Widget) UpdatedAt() time.Time         { return w.updatedAt }

// Business methods
func (w *Widget) UpdateName(name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	w.name = name
	w.updatedAt = time.Now()
	return nil
}

func (w *Widget) SetAPIConfiguration(url string, headers map[string]string) error {
	if w.dataSource != APIDataSource {
		return NewDomainError(ErrInvalidDataSource, "cannot set API configuration for non-API widget")
	}

	if err := validateURL(url); err != nil {
		return err
	}

	w.configuration.apiConfig = &APIConfiguration{
		url:     url,
		headers: headers,
		method:  "GET",
	}
	w.updatedAt = time.Now()
	return nil
}

func (w *Widget) SetRSSConfiguration(config RSSConfiguration) error {
	if w.dataSource != RSSDataSource {
		return NewDomainError(ErrInvalidDataSource, "cannot set RSS configuration for non-RSS widget")
	}

	if err := validateURL(config.feedURL); err != nil {
		return err
	}

	if config.maxItems <= 0 || config.maxItems > 100 {
		return NewDomainError(ErrInvalidConfiguration, "max items must be between 1 and 100")
	}

	w.configuration.rssConfig = &config
	w.updatedAt = time.Now()
	return nil
}

func (w *Widget) SetDataMapping(mapping DataMapping) error {
	// Validate mapping based on template type
	if err := w.validateMappingForTemplate(mapping); err != nil {
		return err
	}

	w.configuration.mapping = mapping
	w.updatedAt = time.Now()
	return nil
}

func (w *Widget) Enable() {
	w.enabled = true
	w.updatedAt = time.Now()
}

func (w *Widget) Disable() {
	w.enabled = false
	w.updatedAt = time.Now()
}

func (w *Widget) UpdateMetadata(description, category string, tags []string, complexity ComplexityLevel) {
	w.metadata = Metadata{
		description: description,
		category:    category,
		tags:        tags,
		complexity:  complexity,
	}
	w.updatedAt = time.Now()
}

// Validation methods
func (w *Widget) Validate() error {
	if err := validateName(w.name); err != nil {
		return err
	}

	if err := validateTemplateType(w.templateType); err != nil {
		return err
	}

	if err := validateDataSource(w.dataSource); err != nil {
		return err
	}

	// Validate configuration based on data source
	switch w.dataSource {
	case APIDataSource:
		if w.configuration.apiConfig == nil {
			return NewDomainError(ErrMissingConfiguration, "API configuration is required for API widgets")
		}
		if err := validateURL(w.configuration.apiConfig.url); err != nil {
			return err
		}
	case RSSDataSource:
		if w.configuration.rssConfig == nil {
			return NewDomainError(ErrMissingConfiguration, "RSS configuration is required for RSS widgets")
		}
		if err := validateURL(w.configuration.rssConfig.feedURL); err != nil {
			return err
		}
	}

	return nil
}

func (w *Widget) validateMappingForTemplate(mapping DataMapping) error {
	requiredFields := getRequiredFieldsForTemplate(w.templateType)

	for _, field := range requiredFields {
		if _, exists := mapping.fields[field]; !exists {
			return NewDomainError(ErrMissingRequiredField, fmt.Sprintf("required field '%s' is missing from mapping", field))
		}
	}

	return nil
}

// Helper functions
func validateName(name string) error {
	if name == "" {
		return NewDomainError(ErrInvalidName, "widget name cannot be empty")
	}
	if len(name) > 255 {
		return NewDomainError(ErrInvalidName, "widget name cannot exceed 255 characters")
	}
	return nil
}

func validateTemplateType(templateType TemplateType) error {
	validTypes := []TemplateType{
		KeyValue, TitleSubtitleValue, IconList, MetricGrid,
		WeatherCurrent, TimeDisplay, StatusList, ChartSimple,
		TextBlock, ImageCaption,
	}

	for _, valid := range validTypes {
		if templateType == valid {
			return nil
		}
	}

	return NewDomainError(ErrInvalidTemplateType, fmt.Sprintf("invalid template type: %s", templateType))
}

func validateDataSource(dataSource DataSource) error {
	if dataSource != APIDataSource && dataSource != RSSDataSource {
		return NewDomainError(ErrInvalidDataSource, fmt.Sprintf("invalid data source: %s", dataSource))
	}
	return nil
}

func validateURL(url string) error {
	if url == "" {
		return NewDomainError(ErrInvalidURL, "URL cannot be empty")
	}
	// Add more URL validation as needed
	return nil
}

func getRequiredFieldsForTemplate(templateType TemplateType) []string {
	switch templateType {
	case KeyValue:
		return []string{"key", "value"}
	case TitleSubtitleValue:
		return []string{"title", "subtitle", "value"}
	case WeatherCurrent:
		return []string{"temperature", "condition"}
	case TimeDisplay:
		return []string{"time"}
	default:
		return []string{}
	}
}

// Domain repository interface
type Repository interface {
	Save(widget *Widget) error
	FindByID(id WidgetID) (*Widget, error)
	FindAll() ([]*Widget, error)
	FindByName(name string) (*Widget, error)
	Delete(id WidgetID) error
	Count() (int, error)
}

// Domain service interface
type Service interface {
	CreateWidget(name string, templateType TemplateType, dataSource DataSource) (*Widget, error)
	GetWidget(id WidgetID) (*Widget, error)
	UpdateWidget(id WidgetID, updates WidgetUpdates) (*Widget, error)
	DeleteWidget(id WidgetID) error
	ListWidgets(filter ListFilter) ([]*Widget, error)
	ValidateWidget(widget *Widget) error
}

// WidgetUpdates represents updates to a widget
type WidgetUpdates struct {
	Name        *string
	Enabled     *bool
	APIConfig   *APIConfiguration
	RSSConfig   *RSSConfiguration
	DataMapping *DataMapping
	Metadata    *Metadata
}

// ListFilter represents filtering options for widget lists
type ListFilter struct {
	DataSource   *DataSource
	TemplateType *TemplateType
	Enabled      *bool
	Search       string
	Limit        int
	Offset       int
}
