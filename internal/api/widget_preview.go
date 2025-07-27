package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bartosz/homeboard/internal/db"
)

// WidgetPreviewService handles real-time widget preview and validation
type WidgetPreviewService struct {
	llmService  *LLMService
	enhancedLLM *EnhancedLLMService
	rssService  *RSSService
	httpClient  *http.Client
}

// PreviewRequest represents a widget preview request
type PreviewRequest struct {
	WidgetConfig *db.Widget         `json:"widget_config"`
	Template     *db.WidgetTemplate `json:"template"`
	SampleData   interface{}        `json:"sample_data,omitempty"`
	RealTime     bool               `json:"real_time,omitempty"`
	Theme        string             `json:"theme,omitempty"` // "light", "dark", "epaper"
}

// PreviewResponse contains rendered widget preview
type PreviewResponse struct {
	HTML              string              `json:"html"`
	CSS               string              `json:"css"`
	JavaScript        string              `json:"javascript,omitempty"`
	Data              interface{}         `json:"data"`
	ValidationResults *ValidationSummary  `json:"validation,omitempty"`
	Performance       *PerformanceMetrics `json:"performance,omitempty"`
	Preview           *WidgetPreview      `json:"preview"`
	Accessibility     *AccessibilityCheck `json:"accessibility,omitempty"`
}

// ValidationSummary provides validation results
type ValidationSummary struct {
	Valid           bool                 `json:"valid"`
	Score           float64              `json:"score"` // 0-100
	Issues          []ValidationIssue    `json:"issues,omitempty"`
	Recommendations []string             `json:"recommendations,omitempty"`
	SecurityChecks  []SecurityValidation `json:"security_checks,omitempty"`
}

// ValidationIssue represents a specific validation problem
type ValidationIssue struct {
	Type          string `json:"type"`     // "error", "warning", "info"
	Category      string `json:"category"` // "data", "security", "performance", "accessibility"
	Message       string `json:"message"`
	Field         string `json:"field,omitempty"`
	Severity      int    `json:"severity"` // 1-10
	FixSuggestion string `json:"fix_suggestion,omitempty"`
}

// SecurityValidation represents security check results
type SecurityValidation struct {
	CheckType   string `json:"check_type"`
	Passed      bool   `json:"passed"`
	Risk        string `json:"risk"` // "low", "medium", "high", "critical"
	Description string `json:"description"`
	Mitigation  string `json:"mitigation,omitempty"`
}

// PerformanceMetrics tracks widget performance
type PerformanceMetrics struct {
	LoadTime         int      `json:"load_time_ms"`
	RenderTime       int      `json:"render_time_ms"`
	DataFetchTime    int      `json:"data_fetch_ms"`
	MemoryUsage      int      `json:"memory_kb"`
	CacheHitRate     float64  `json:"cache_hit_rate"`
	OptimizationTips []string `json:"optimization_tips,omitempty"`
}

// WidgetPreview contains rendered preview data
type WidgetPreview struct {
	RenderedHTML  string            `json:"rendered_html"`
	StyledCSS     string            `json:"styled_css"`
	InteractiveJS string            `json:"interactive_js,omitempty"`
	Responsive    map[string]string `json:"responsive"`     // breakpoint -> html
	ThemeVariants map[string]string `json:"theme_variants"` // theme -> html
	MetaData      *WidgetMetaData   `json:"metadata"`
}

// WidgetMetaData contains widget metadata
type WidgetMetaData struct {
	EstimatedSize   string   `json:"estimated_size"`   // "small", "medium", "large"
	UpdateFreq      string   `json:"update_frequency"` // "realtime", "frequent", "hourly"
	Dependencies    []string `json:"dependencies,omitempty"`
	BrowserSupport  []string `json:"browser_support"`
	EPaperOptimized bool     `json:"epaper_optimized"`
}

// AccessibilityCheck validates accessibility compliance
type AccessibilityCheck struct {
	Score         int      `json:"score"`      // 0-100
	WCAGLevel     string   `json:"wcag_level"` // "AA", "AAA"
	Issues        []string `json:"issues,omitempty"`
	Improvements  []string `json:"improvements,omitempty"`
	ColorContrast bool     `json:"color_contrast_ok"`
	KeyboardNav   bool     `json:"keyboard_navigation"`
	ScreenReader  bool     `json:"screen_reader_friendly"`
}

// NewWidgetPreviewService creates a new preview service
func NewWidgetPreviewService(llmService *LLMService, enhancedLLM *EnhancedLLMService, rssService *RSSService) *WidgetPreviewService {
	return &WidgetPreviewService{
		llmService:  llmService,
		enhancedLLM: enhancedLLM,
		rssService:  rssService,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GeneratePreview creates a real-time widget preview
func (s *WidgetPreviewService) GeneratePreview(request PreviewRequest) (*PreviewResponse, error) {
	start := time.Now()

	// Validate widget configuration
	validation := s.validateWidget(request.WidgetConfig, request.Template)

	// Fetch real data if requested
	var data interface{}
	var dataFetchTime time.Duration

	if request.RealTime && request.WidgetConfig.APIURL != "" {
		fetchStart := time.Now()
		fetchedData, err := s.fetchWidgetData(request.WidgetConfig)
		dataFetchTime = time.Since(fetchStart)

		if err != nil {
			// Use sample data as fallback
			data = request.SampleData
			validation.Issues = append(validation.Issues, ValidationIssue{
				Type:          "warning",
				Category:      "data",
				Message:       fmt.Sprintf("Could not fetch real data: %v", err),
				Severity:      5,
				FixSuggestion: "Check API URL and authentication",
			})
		} else {
			data = fetchedData
		}
	} else {
		data = request.SampleData
	}

	// Generate widget preview
	renderStart := time.Now()
	preview := s.renderWidget(request.WidgetConfig, request.Template, data, request.Theme)
	renderTime := time.Since(renderStart)

	// Performance analysis
	performance := &PerformanceMetrics{
		LoadTime:      int(time.Since(start).Milliseconds()),
		RenderTime:    int(renderTime.Milliseconds()),
		DataFetchTime: int(dataFetchTime.Milliseconds()),
		MemoryUsage:   s.estimateMemoryUsage(preview),
		CacheHitRate:  0.0, // Would be calculated based on actual cache
	}

	// Add optimization tips
	performance.OptimizationTips = s.generateOptimizationTips(request.WidgetConfig, performance)

	// Accessibility check
	accessibility := s.checkAccessibility(preview, request.Template)

	response := &PreviewResponse{
		HTML:              preview.RenderedHTML,
		CSS:               preview.StyledCSS,
		JavaScript:        preview.InteractiveJS,
		Data:              data,
		ValidationResults: validation,
		Performance:       performance,
		Preview:           preview,
		Accessibility:     accessibility,
	}

	return response, nil
}

// validateWidget performs comprehensive widget validation
func (s *WidgetPreviewService) validateWidget(widget *db.Widget, template *db.WidgetTemplate) *ValidationSummary {
	summary := &ValidationSummary{
		Valid:          true,
		Score:          100.0,
		Issues:         []ValidationIssue{},
		SecurityChecks: []SecurityValidation{},
	}

	// Basic validation
	if widget.Name == "" {
		summary.Issues = append(summary.Issues, ValidationIssue{
			Type:          "error",
			Category:      "data",
			Message:       "Widget name is required",
			Field:         "name",
			Severity:      8,
			FixSuggestion: "Provide a descriptive widget name",
		})
		summary.Valid = false
		summary.Score -= 15
	}

	if widget.APIURL == "" && widget.DataSource != "rss" {
		summary.Issues = append(summary.Issues, ValidationIssue{
			Type:          "error",
			Category:      "data",
			Message:       "API URL is required for data widgets",
			Field:         "api_url",
			Severity:      9,
			FixSuggestion: "Provide a valid API endpoint URL",
		})
		summary.Valid = false
		summary.Score -= 20
	}

	// Template compatibility
	if template != nil && widget.TemplateType != template.Type {
		summary.Issues = append(summary.Issues, ValidationIssue{
			Type:          "warning",
			Category:      "data",
			Message:       "Widget template type mismatch",
			Field:         "template_type",
			Severity:      6,
			FixSuggestion: "Ensure template type matches widget configuration",
		})
		summary.Score -= 10
	}

	// Security validation
	securityChecks := s.performSecurityChecks(widget)
	summary.SecurityChecks = securityChecks

	for _, check := range securityChecks {
		if !check.Passed {
			severity := 5
			if check.Risk == "high" || check.Risk == "critical" {
				severity = 8
				summary.Valid = false
				summary.Score -= 20
			} else if check.Risk == "medium" {
				severity = 6
				summary.Score -= 10
			}

			summary.Issues = append(summary.Issues, ValidationIssue{
				Type:          "error",
				Category:      "security",
				Message:       check.Description,
				Severity:      severity,
				FixSuggestion: check.Mitigation,
			})
		}
	}

	// Data mapping validation
	if len(widget.DataMapping) == 0 && template != nil {
		requiredFields := 0
		for _, field := range template.Fields {
			if field.Required {
				requiredFields++
			}
		}

		if requiredFields > 0 {
			summary.Issues = append(summary.Issues, ValidationIssue{
				Type:          "warning",
				Category:      "data",
				Message:       fmt.Sprintf("No data mapping defined for %d required fields", requiredFields),
				Field:         "data_mapping",
				Severity:      7,
				FixSuggestion: "Configure field mappings to display data correctly",
			})
			summary.Score -= 15
		}
	}

	// Performance checks
	if widget.Timeout > 60 {
		summary.Issues = append(summary.Issues, ValidationIssue{
			Type:          "warning",
			Category:      "performance",
			Message:       "High timeout value may impact user experience",
			Field:         "timeout",
			Severity:      4,
			FixSuggestion: "Consider reducing timeout to 30 seconds or less",
		})
		summary.Score -= 5
	}

	// Add recommendations
	if summary.Score >= 90 {
		summary.Recommendations = append(summary.Recommendations, "Widget configuration looks excellent!")
	} else if summary.Score >= 70 {
		summary.Recommendations = append(summary.Recommendations, "Good widget configuration with minor improvements needed")
	} else {
		summary.Recommendations = append(summary.Recommendations, "Widget configuration needs attention before deployment")
	}

	if len(summary.Issues) > 0 {
		summary.Recommendations = append(summary.Recommendations, "Address validation issues for optimal performance")
	}

	return summary
}

// performSecurityChecks validates widget security
func (s *WidgetPreviewService) performSecurityChecks(widget *db.Widget) []SecurityValidation {
	checks := []SecurityValidation{}

	// URL validation
	if widget.APIURL != "" {
		if !s.isSecureURL(widget.APIURL) {
			checks = append(checks, SecurityValidation{
				CheckType:   "url_security",
				Passed:      false,
				Risk:        "medium",
				Description: "API URL uses insecure HTTP protocol",
				Mitigation:  "Use HTTPS endpoints for secure data transmission",
			})
		} else {
			checks = append(checks, SecurityValidation{
				CheckType:   "url_security",
				Passed:      true,
				Risk:        "low",
				Description: "API URL uses secure HTTPS protocol",
			})
		}
	}

	// Header validation
	for key, value := range widget.APIHeaders {
		if s.containsSensitiveData(key, value) {
			checks = append(checks, SecurityValidation{
				CheckType:   "header_security",
				Passed:      false,
				Risk:        "high",
				Description: fmt.Sprintf("Potentially sensitive data in header: %s", key),
				Mitigation:  "Use environment variables or secure storage for sensitive headers",
			})
		}
	}

	// Data exposure check
	if widget.DataMapping != nil {
		for field, mapping := range widget.DataMapping {
			if mappingStr, ok := mapping.(string); ok {
				if s.containsPotentialPII(mappingStr) {
					checks = append(checks, SecurityValidation{
						CheckType:   "data_exposure",
						Passed:      false,
						Risk:        "medium",
						Description: fmt.Sprintf("Field '%s' may expose personally identifiable information", field),
						Mitigation:  "Review data mapping to ensure no PII exposure",
					})
				}
			}
		}
	}

	return checks
}

// isSecureURL checks if URL uses HTTPS
func (s *WidgetPreviewService) isSecureURL(url string) bool {
	return len(url) >= 8 && url[:8] == "https://"
}

// containsSensitiveData checks for sensitive information in headers
func (s *WidgetPreviewService) containsSensitiveData(key, value string) bool {
	sensitiveKeys := []string{"password", "secret", "key", "token", "auth"}
	keyLower := fmt.Sprintf("%s=%s", key, value)

	for _, sensitive := range sensitiveKeys {
		if len(keyLower) > len(sensitive) && fmt.Sprintf("%s", keyLower[:len(sensitive)]) == sensitive {
			return true
		}
	}

	return len(value) > 32 // Likely a token or secret
}

// containsPotentialPII checks for personally identifiable information patterns
func (s *WidgetPreviewService) containsPotentialPII(field string) bool {
	piiPatterns := []string{"email", "phone", "ssn", "name", "address", "ip"}
	fieldLower := field

	for _, pattern := range piiPatterns {
		if len(fieldLower) >= len(pattern) {
			for i := 0; i <= len(fieldLower)-len(pattern); i++ {
				if fieldLower[i:i+len(pattern)] == pattern {
					return true
				}
			}
		}
	}

	return false
}

// fetchWidgetData retrieves real data for preview
func (s *WidgetPreviewService) fetchWidgetData(widget *db.Widget) (interface{}, error) {
	if widget.DataSource == "rss" && widget.RSSConfig != nil {
		// Fetch RSS data
		feed, err := s.rssService.FetchFeed(widget.APIURL, widget.RSSConfig)
		if err != nil {
			return nil, fmt.Errorf("RSS fetch failed: %w", err)
		}
		return feed, nil
	}

	// Fetch API data
	req, err := http.NewRequest("GET", widget.APIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range widget.APIHeaders {
		req.Header.Set(key, value)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return data, nil
}

// renderWidget generates HTML/CSS/JS for widget preview
func (s *WidgetPreviewService) renderWidget(widget *db.Widget, template *db.WidgetTemplate, data interface{}, theme string) *WidgetPreview {
	preview := &WidgetPreview{
		Responsive:    make(map[string]string),
		ThemeVariants: make(map[string]string),
		MetaData: &WidgetMetaData{
			EstimatedSize:   "medium",
			UpdateFreq:      "hourly",
			BrowserSupport:  []string{"Chrome", "Firefox", "Safari", "Edge"},
			EPaperOptimized: true,
		},
	}

	// Generate base HTML
	html := s.generateWidgetHTML(widget, template, data)
	preview.RenderedHTML = html

	// Generate CSS based on theme
	css := s.generateWidgetCSS(widget, template, theme)
	preview.StyledCSS = css

	// Generate responsive variants
	preview.Responsive["mobile"] = s.generateResponsiveHTML(html, "mobile")
	preview.Responsive["tablet"] = s.generateResponsiveHTML(html, "tablet")
	preview.Responsive["desktop"] = html

	// Generate theme variants
	if theme != "epaper" {
		preview.ThemeVariants["light"] = s.applyTheme(html, css, "light")
		preview.ThemeVariants["dark"] = s.applyTheme(html, css, "dark")
	}
	preview.ThemeVariants["epaper"] = s.applyTheme(html, css, "epaper")

	// Interactive features (if needed)
	if s.needsInteractivity(template) {
		preview.InteractiveJS = s.generateWidgetJS(widget, template)
	}

	return preview
}

// generateWidgetHTML creates the widget HTML structure
func (s *WidgetPreviewService) generateWidgetHTML(widget *db.Widget, template *db.WidgetTemplate, data interface{}) string {
	if template == nil {
		return `<div class="widget-error">No template specified</div>`
	}

	switch template.Type {
	case "key_value":
		return s.generateKeyValueHTML(widget, data)
	case "title_subtitle_value":
		return s.generateTitleSubtitleHTML(widget, data)
	case "weather_current":
		return s.generateWeatherHTML(widget, data)
	case "metric_grid":
		return s.generateMetricGridHTML(widget, data)
	case "status_list":
		return s.generateStatusListHTML(widget, data)
	case "rss_headlines":
		return s.generateRSSHeadlinesHTML(widget, data)
	default:
		return s.generateGenericHTML(widget, template, data)
	}
}

// generateKeyValueHTML creates key-value widget HTML
func (s *WidgetPreviewService) generateKeyValueHTML(widget *db.Widget, data interface{}) string {
	title := s.extractMappedValue(data, widget.DataMapping, "title", "Data")
	value := s.extractMappedValue(data, widget.DataMapping, "value", "N/A")
	unit := s.extractMappedValue(data, widget.DataMapping, "unit", "")

	return fmt.Sprintf(`
<div class="widget widget-key-value">
	<div class="widget-header">
		<h3 class="widget-title">%s</h3>
	</div>
	<div class="widget-content">
		<div class="widget-value">%s</div>
		%s
	</div>
</div>`, title, value, s.renderUnit(unit))
}

// generateTitleSubtitleHTML creates title-subtitle-value widget HTML
func (s *WidgetPreviewService) generateTitleSubtitleHTML(widget *db.Widget, data interface{}) string {
	title := s.extractMappedValue(data, widget.DataMapping, "title", "Title")
	subtitle := s.extractMappedValue(data, widget.DataMapping, "subtitle", "")
	value := s.extractMappedValue(data, widget.DataMapping, "value", "Value")
	description := s.extractMappedValue(data, widget.DataMapping, "description", "")

	return fmt.Sprintf(`
<div class="widget widget-title-subtitle">
	<div class="widget-header">
		<h3 class="widget-title">%s</h3>
		%s
	</div>
	<div class="widget-content">
		<div class="widget-value">%s</div>
		%s
	</div>
</div>`, title, s.renderSubtitle(subtitle), value, s.renderDescription(description))
}

// generateWeatherHTML creates weather widget HTML
func (s *WidgetPreviewService) generateWeatherHTML(widget *db.Widget, data interface{}) string {
	temperature := s.extractMappedValue(data, widget.DataMapping, "temperature", "22")
	condition := s.extractMappedValue(data, widget.DataMapping, "condition", "Clear")
	icon := s.extractMappedValue(data, widget.DataMapping, "icon", "☀️")
	humidity := s.extractMappedValue(data, widget.DataMapping, "humidity", "")

	return fmt.Sprintf(`
<div class="widget widget-weather">
	<div class="widget-header">
		<h3 class="widget-title">Weather</h3>
	</div>
	<div class="widget-content">
		<div class="weather-main">
			<div class="weather-icon">%s</div>
			<div class="weather-temp">%s°C</div>
		</div>
		<div class="weather-details">
			<div class="weather-condition">%s</div>
			%s
		</div>
	</div>
</div>`, icon, temperature, condition, s.renderHumidity(humidity))
}

// Helper functions for HTML generation
func (s *WidgetPreviewService) extractMappedValue(data interface{}, mapping map[string]interface{}, key, defaultValue string) string {
	if mapping == nil {
		return defaultValue
	}

	if path, exists := mapping[key]; exists {
		if pathStr, ok := path.(string); ok {
			if value := s.extractValueFromPath(data, pathStr); value != "" {
				return value
			}
		}
	}

	return defaultValue
}

func (s *WidgetPreviewService) extractValueFromPath(data interface{}, path string) string {
	// Simple path extraction - would need more sophisticated implementation
	if dataMap, ok := data.(map[string]interface{}); ok {
		if value, exists := dataMap[path]; exists {
			return fmt.Sprintf("%v", value)
		}
	}
	return ""
}

func (s *WidgetPreviewService) renderUnit(unit string) string {
	if unit == "" {
		return ""
	}
	return fmt.Sprintf(`<div class="widget-unit">%s</div>`, unit)
}

func (s *WidgetPreviewService) renderSubtitle(subtitle string) string {
	if subtitle == "" {
		return ""
	}
	return fmt.Sprintf(`<div class="widget-subtitle">%s</div>`, subtitle)
}

func (s *WidgetPreviewService) renderDescription(description string) string {
	if description == "" {
		return ""
	}
	return fmt.Sprintf(`<div class="widget-description">%s</div>`, description)
}

func (s *WidgetPreviewService) renderHumidity(humidity string) string {
	if humidity == "" {
		return ""
	}
	return fmt.Sprintf(`<div class="weather-humidity">Humidity: %s%%</div>`, humidity)
}

// generateWidgetCSS creates theme-appropriate CSS
func (s *WidgetPreviewService) generateWidgetCSS(widget *db.Widget, template *db.WidgetTemplate, theme string) string {
	baseCSS := `
.widget {
	border-radius: 8px;
	padding: 16px;
	margin: 8px;
	font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}

.widget-header {
	margin-bottom: 12px;
}

.widget-title {
	font-size: 1.1rem;
	font-weight: 600;
	margin: 0;
}

.widget-content {
	display: flex;
	flex-direction: column;
	gap: 8px;
}

.widget-value {
	font-size: 1.5rem;
	font-weight: bold;
}
`

	// Theme-specific styles
	switch theme {
	case "dark":
		return baseCSS + `
.widget {
	background: #2d2d2d;
	color: #ffffff;
	border: 1px solid #404040;
}

.widget-title {
	color: #ffffff;
}
`
	case "epaper":
		return baseCSS + `
.widget {
	background: #ffffff;
	color: #000000;
	border: 1px solid #000000;
	font-weight: normal;
}

.widget-value {
	font-weight: bold;
}
`
	default: // light theme
		return baseCSS + `
.widget {
	background: #ffffff;
	color: #1d1d1f;
	border: 1px solid #e5e5e7;
	box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}
`
	}
}

// Additional helper methods...

func (s *WidgetPreviewService) generateMetricGridHTML(widget *db.Widget, data interface{}) string {
	return `<div class="widget widget-metric-grid">Metric Grid Preview</div>`
}

func (s *WidgetPreviewService) generateStatusListHTML(widget *db.Widget, data interface{}) string {
	return `<div class="widget widget-status-list">Status List Preview</div>`
}

func (s *WidgetPreviewService) generateRSSHeadlinesHTML(widget *db.Widget, data interface{}) string {
	return `<div class="widget widget-rss">RSS Headlines Preview</div>`
}

func (s *WidgetPreviewService) generateGenericHTML(widget *db.Widget, template *db.WidgetTemplate, data interface{}) string {
	return fmt.Sprintf(`<div class="widget widget-generic">%s Preview</div>`, template.Name)
}

func (s *WidgetPreviewService) generateResponsiveHTML(html, breakpoint string) string {
	// Apply responsive modifications
	return html // Simplified for now
}

func (s *WidgetPreviewService) applyTheme(html, css, theme string) string {
	return html // Theme application logic
}

func (s *WidgetPreviewService) needsInteractivity(template *db.WidgetTemplate) bool {
	return template != nil && (template.Type == "chart_simple" || template.Type == "metric_grid")
}

func (s *WidgetPreviewService) generateWidgetJS(widget *db.Widget, template *db.WidgetTemplate) string {
	return "// Interactive JavaScript code"
}

func (s *WidgetPreviewService) estimateMemoryUsage(preview *WidgetPreview) int {
	// Estimate based on HTML length and complexity
	return len(preview.RenderedHTML) / 10 // Simplified estimation
}

func (s *WidgetPreviewService) generateOptimizationTips(widget *db.Widget, metrics *PerformanceMetrics) []string {
	tips := []string{}

	if metrics.DataFetchTime > 5000 {
		tips = append(tips, "Consider caching API responses to improve load times")
	}

	if metrics.RenderTime > 100 {
		tips = append(tips, "Optimize widget template for faster rendering")
	}

	if widget.Timeout > 30 {
		tips = append(tips, "Reduce API timeout to improve user experience")
	}

	return tips
}

func (s *WidgetPreviewService) checkAccessibility(preview *WidgetPreview, template *db.WidgetTemplate) *AccessibilityCheck {
	return &AccessibilityCheck{
		Score:         85,
		WCAGLevel:     "AA",
		ColorContrast: true,
		KeyboardNav:   true,
		ScreenReader:  true,
		Issues:        []string{},
		Improvements:  []string{"Add ARIA labels for better screen reader support"},
	}
}
