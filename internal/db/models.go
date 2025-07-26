package db

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Client represents a registered client device
type Client struct {
	ID                  int       `json:"id" db:"id"`
	IPAddress           string    `json:"ip_address" db:"ip_address"`
	LastSeen            time.Time `json:"last_seen" db:"last_seen"`
	AssignedDashboardID *int      `json:"assigned_dashboard_id" db:"assigned_dashboard_id"`
	Name                string    `json:"name" db:"name"`             // Optional client name
	UserAgent           string    `json:"user_agent" db:"user_agent"` // Client's user agent
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// Widget represents a widget configuration
type Widget struct {
	ID           int                    `json:"id" db:"id"`
	Name         string                 `json:"name" db:"name"`
	TemplateType string                 `json:"template_type" db:"template_type"`
	DataSource   string                 `json:"data_source" db:"data_source"` // "api" or "rss"
	APIURL       string                 `json:"api_url" db:"api_url"`
	APIHeaders   map[string]string      `json:"api_headers" db:"api_headers"`
	DataMapping  map[string]interface{} `json:"data_mapping" db:"data_mapping"`
	RSSConfig    *RSSConfig             `json:"rss_config,omitempty" db:"rss_config"`
	Description  string                 `json:"description" db:"description"`
	Timeout      int                    `json:"timeout" db:"timeout"` // in seconds
	Enabled      bool                   `json:"enabled" db:"enabled"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// Dashboard represents a dashboard layout
type Dashboard struct {
	ID          int               `json:"id" db:"id"`
	Name        string            `json:"name" db:"name"`
	Description string            `json:"description" db:"description"`
	IsDefault   bool              `json:"is_default" db:"is_default"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
	Widgets     []DashboardWidget `json:"widgets,omitempty"` // Loaded separately
}

// DashboardWidget represents the join table between dashboards and widgets
type DashboardWidget struct {
	ID           int `json:"id" db:"id"`
	DashboardID  int `json:"dashboard_id" db:"dashboard_id"`
	WidgetID     int `json:"widget_id" db:"widget_id"`
	DisplayOrder int `json:"display_order" db:"display_order"`
	// Position and size for future grid layout
	GridX      int `json:"grid_x" db:"grid_x"`
	GridY      int `json:"grid_y" db:"grid_y"`
	GridWidth  int `json:"grid_width" db:"grid_width"`
	GridHeight int `json:"grid_height" db:"grid_height"`

	// Embedded widget details (populated when loading dashboard)
	Widget *Widget `json:"widget,omitempty"`
}

// WidgetTemplate represents available widget templates
type WidgetTemplate struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Preview     string                 `json:"preview"`
	Fields      []WidgetTemplateField  `json:"fields"`
	Example     map[string]interface{} `json:"example"`
}

// WidgetTemplateField represents a field in a widget template
type WidgetTemplateField struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Type        string `json:"type"` // text, number, boolean, array
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Placeholder string `json:"placeholder"`
}

// LLMAnalyzeRequest represents the request for LLM analysis
type LLMAnalyzeRequest struct {
	APIURL         string            `json:"apiUrl"`
	WidgetTemplate string            `json:"widgetTemplate"`
	APIHeaders     map[string]string `json:"apiHeaders,omitempty"`
	SampleData     interface{}       `json:"sampleData,omitempty"`
}

// LLMAnalyzeResponse represents the response from LLM analysis
type LLMAnalyzeResponse struct {
	APIData     interface{}            `json:"apiData,omitempty"`
	DataMapping map[string]interface{} `json:"dataMapping"`
	Suggestions []MappingSuggestion    `json:"suggestions,omitempty"`
	Reasoning   string                 `json:"reasoning,omitempty"`
	Confidence  float64                `json:"confidence,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// MappingSuggestion represents an alternative mapping suggestion
type MappingSuggestion struct {
	Field       string      `json:"field"`
	JSONPath    string      `json:"jsonPath"`
	Value       interface{} `json:"value"`
	Confidence  float64     `json:"confidence"`
	Description string      `json:"description"`
}

// RSSConfig represents RSS feed configuration
type RSSConfig struct {
	FeedURL       string `json:"feed_url"`
	MaxItems      int    `json:"max_items"`      // Maximum number of items to fetch (default: 10)
	CacheMinutes  int    `json:"cache_minutes"`  // Cache duration in minutes (default: 30)
	ItemFilter    string `json:"item_filter"`    // Optional filter for items (e.g., "latest", "today")
	IncludeImage  bool   `json:"include_image"`  // Whether to include item images
	IncludeAuthor bool   `json:"include_author"` // Whether to include author information
	DateFormat    string `json:"date_format"`    // Custom date format for display
}

// RSSItem represents a parsed RSS feed item
type RSSItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Author      string `json:"author,omitempty"`
	PubDate     string `json:"pub_date"`
	GUID        string `json:"guid,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	Category    string `json:"category,omitempty"`
}

// RSSFeed represents a parsed RSS feed
type RSSFeed struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	Language    string    `json:"language,omitempty"`
	LastBuild   string    `json:"last_build,omitempty"`
	Items       []RSSItem `json:"items"`
}

// Helper methods for database serialization

// APIHeadersJSON converts APIHeaders map to JSON string for database storage
func (w *Widget) APIHeadersJSON() (string, error) {
	if w.APIHeaders == nil {
		return "{}", nil
	}
	data, err := json.Marshal(w.APIHeaders)
	return string(data), err
}

// SetAPIHeadersFromJSON sets APIHeaders from JSON string
func (w *Widget) SetAPIHeadersFromJSON(jsonStr string) error {
	if jsonStr == "" {
		w.APIHeaders = make(map[string]string)
		return nil
	}
	return json.Unmarshal([]byte(jsonStr), &w.APIHeaders)
}

// DataMappingJSON converts DataMapping to JSON string for database storage
func (w *Widget) DataMappingJSON() (string, error) {
	if w.DataMapping == nil {
		return "{}", nil
	}
	data, err := json.Marshal(w.DataMapping)
	return string(data), err
}

// SetDataMappingFromJSON sets DataMapping from JSON string
func (w *Widget) SetDataMappingFromJSON(jsonStr string) error {
	if jsonStr == "" {
		w.DataMapping = make(map[string]interface{})
		return nil
	}
	return json.Unmarshal([]byte(jsonStr), &w.DataMapping)
}

// RSSConfigJSON converts RSSConfig to JSON string for database storage
func (w *Widget) RSSConfigJSON() (string, error) {
	if w.RSSConfig == nil {
		return "", nil
	}
	data, err := json.Marshal(w.RSSConfig)
	return string(data), err
}

// SetRSSConfigFromJSON sets RSSConfig from JSON string
func (w *Widget) SetRSSConfigFromJSON(jsonStr string) error {
	if jsonStr == "" {
		w.RSSConfig = nil
		return nil
	}
	var config RSSConfig
	err := json.Unmarshal([]byte(jsonStr), &config)
	if err != nil {
		return err
	}
	w.RSSConfig = &config
	return nil
}

// IsValidTemplateType checks if the template type is valid
func (w *Widget) IsValidTemplateType() bool {
	validTypes := []string{
		"key_value",
		"title_subtitle_value",
		"icon_list",
		"metric_grid",
		"weather_current",
		"time_display",
		"status_list",
		"chart_simple",
		"text_block",
		"image_caption",
	}

	for _, validType := range validTypes {
		if w.TemplateType == validType {
			return true
		}
	}
	return false
}

// Validate validates widget configuration
func (w *Widget) Validate() error {
	if w.Name == "" {
		return ErrWidgetNameRequired
	}
	if w.APIURL == "" {
		return ErrWidgetAPIURLRequired
	}
	if w.TemplateType == "" {
		return ErrWidgetTemplateTypeRequired
	}
	if !w.IsValidTemplateType() {
		return ErrWidgetInvalidTemplateType
	}
	if w.Timeout <= 0 {
		w.Timeout = 30 // Default timeout
	}
	return nil
}

// Validate validates dashboard configuration
func (d *Dashboard) Validate() error {
	if d.Name == "" {
		return ErrDashboardNameRequired
	}
	return nil
}

// Validate validates client configuration
func (c *Client) Validate() error {
	if c.IPAddress == "" {
		return ErrClientIPRequired
	}
	return nil
}

// Custom errors
var (
	ErrWidgetNameRequired         = &ValidationError{Field: "name", Message: "Widget name is required"}
	ErrWidgetAPIURLRequired       = &ValidationError{Field: "api_url", Message: "API URL is required"}
	ErrWidgetTemplateTypeRequired = &ValidationError{Field: "template_type", Message: "Template type is required"}
	ErrWidgetInvalidTemplateType  = &ValidationError{Field: "template_type", Message: "Invalid template type"}
	ErrDashboardNameRequired      = &ValidationError{Field: "name", Message: "Dashboard name is required"}
	ErrClientIPRequired           = &ValidationError{Field: "ip_address", Message: "Client IP address is required"}
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NullTime wraps sql.NullTime for JSON serialization
type NullTime struct {
	sql.NullTime
}

// MarshalJSON implements json.Marshaler interface
func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(nt.Time)
}

// UnmarshalJSON implements json.Unmarshaler interface
func (nt *NullTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nt.Valid = false
		return nil
	}
	var t time.Time
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}
	nt.Time = t
	nt.Valid = true
	return nil
}
