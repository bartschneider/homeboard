package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bartosz/homeboard/internal/db"
)

// EnhancedRSSService extends RSS functionality with AI-powered features
type EnhancedRSSService struct {
	rssService *RSSService
	llmService *LLMService
	httpClient *http.Client
}

// RSSAnalysisRequest for AI-powered RSS analysis
type RSSAnalysisRequest struct {
	FeedURL        string                 `json:"feed_url"`
	AnalysisType   string                 `json:"analysis_type"` // "content", "structure", "categorization"
	UserIntent     string                 `json:"user_intent,omitempty"`
	Categories     []string               `json:"categories,omitempty"`
	ContentFilters []string               `json:"content_filters,omitempty"`
	LanguagePrefs  []string               `json:"language_preferences,omitempty"`
	Context        map[string]interface{} `json:"context,omitempty"`
}

// RSSAnalysisResponse contains AI analysis results
type RSSAnalysisResponse struct {
	FeedAnalysis      *FeedAnalysis       `json:"feed_analysis"`
	ContentSummary    *ContentSummary     `json:"content_summary"`
	Recommendations   *RSSRecommendations `json:"recommendations"`
	WidgetSuggestions []WidgetSuggestion  `json:"widget_suggestions"`
	AutoConfiguration *db.RSSConfig       `json:"auto_configuration"`
	QualityScore      float64             `json:"quality_score"`
	Confidence        float64             `json:"confidence"`
}

// FeedAnalysis provides detailed RSS feed analysis
type FeedAnalysis struct {
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Language        string                 `json:"language"`
	Category        string                 `json:"category"`
	UpdateFrequency string                 `json:"update_frequency"` // "hourly", "daily", "weekly"
	ItemCount       int                    `json:"item_count"`
	ContentTypes    []string               `json:"content_types"`
	Topics          []TopicAnalysis        `json:"topics"`
	Reliability     *ReliabilityMetrics    `json:"reliability"`
	ContentQuality  *ContentQualityMetrics `json:"content_quality"`
}

// ContentSummary provides content analysis
type ContentSummary struct {
	MainTopics    []string `json:"main_topics"`
	KeyTerms      []string `json:"key_terms"`
	Sentiment     string   `json:"sentiment"`      // "positive", "negative", "neutral", "mixed"
	ReadingLevel  string   `json:"reading_level"`  // "basic", "intermediate", "advanced"
	ContentLength string   `json:"content_length"` // "short", "medium", "long"
	MediaTypes    []string `json:"media_types"`
	SourceTypes   []string `json:"source_types"`
}

// TopicAnalysis represents topic categorization
type TopicAnalysis struct {
	Topic      string   `json:"topic"`
	Confidence float64  `json:"confidence"`
	ItemCount  int      `json:"item_count"`
	Keywords   []string `json:"keywords"`
}

// ReliabilityMetrics assess feed reliability
type ReliabilityMetrics struct {
	UpdateConsistency float64 `json:"update_consistency"` // 0-1
	ContentFreshness  float64 `json:"content_freshness"`  // 0-1
	SourceCredibility float64 `json:"source_credibility"` // 0-1
	TechnicalQuality  float64 `json:"technical_quality"`  // 0-1
	OverallScore      float64 `json:"overall_score"`      // 0-1
}

// ContentQualityMetrics assess content quality
type ContentQualityMetrics struct {
	Readability      float64 `json:"readability"`       // 0-1
	InformationDepth float64 `json:"information_depth"` // 0-1
	Uniqueness       float64 `json:"uniqueness"`        // 0-1
	Engagement       float64 `json:"engagement"`        // 0-1
	OverallQuality   float64 `json:"overall_quality"`   // 0-1
}

// RSSRecommendations provides optimization suggestions
type RSSRecommendations struct {
	OptimalConfig       *db.RSSConfig      `json:"optimal_config"`
	FilterSuggestions   []FilterSuggestion `json:"filter_suggestions"`
	DisplayOptions      []DisplayOption    `json:"display_options"`
	UpdateSchedule      *UpdateSchedule    `json:"update_schedule"`
	QualityImprovements []string           `json:"quality_improvements"`
}

// FilterSuggestion recommends content filtering
type FilterSuggestion struct {
	Type       string   `json:"type"` // "keyword", "category", "date", "author"
	Values     []string `json:"values"`
	Reasoning  string   `json:"reasoning"`
	Impact     string   `json:"impact"` // "high", "medium", "low"
	Confidence float64  `json:"confidence"`
}

// DisplayOption suggests presentation options
type DisplayOption struct {
	Template    string                 `json:"template"`
	Layout      string                 `json:"layout"`
	Fields      []string               `json:"fields"`
	Styling     map[string]interface{} `json:"styling"`
	Description string                 `json:"description"`
	Suitability float64                `json:"suitability"` // 0-1
}

// UpdateSchedule recommends refresh timing
type UpdateSchedule struct {
	Frequency     string   `json:"frequency"`      // "realtime", "every_15min", "hourly", "daily"
	PeakTimes     []string `json:"peak_times"`     // Hours when content is most active
	CacheDuration int      `json:"cache_duration"` // Recommended cache time in minutes
	Reasoning     string   `json:"reasoning"`
}

// WidgetSuggestion recommends widget configurations
type WidgetSuggestion struct {
	WidgetType    string                 `json:"widget_type"`
	TemplateType  string                 `json:"template_type"`
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	DataMapping   map[string]interface{} `json:"data_mapping"`
	Configuration map[string]interface{} `json:"configuration"`
	Suitability   float64                `json:"suitability"` // 0-1
	Benefits      []string               `json:"benefits"`
}

// NewEnhancedRSSService creates enhanced RSS service
func NewEnhancedRSSService(rssService *RSSService, llmService *LLMService) *EnhancedRSSService {
	return &EnhancedRSSService{
		rssService: rssService,
		llmService: llmService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// AnalyzeRSSFeed performs AI-powered RSS feed analysis
func (s *EnhancedRSSService) AnalyzeRSSFeed(request RSSAnalysisRequest) (*RSSAnalysisResponse, error) {
	// Fetch RSS feed data
	basicConfig := &db.RSSConfig{
		MaxItems:     20, // Get more items for analysis
		CacheMinutes: 5,  // Short cache for analysis
	}

	feed, err := s.rssService.FetchFeed(request.FeedURL, basicConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
	}

	// Perform AI analysis
	analysis, err := s.performAIAnalysis(feed, request)
	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	// Generate recommendations
	recommendations := s.generateRecommendations(feed, analysis, request)

	// Generate widget suggestions
	widgetSuggestions := s.generateWidgetSuggestions(feed, analysis, request)

	// Create auto-configuration
	autoConfig := s.generateAutoConfiguration(feed, analysis, request)

	response := &RSSAnalysisResponse{
		FeedAnalysis:      analysis,
		ContentSummary:    s.analyzeContent(feed),
		Recommendations:   recommendations,
		WidgetSuggestions: widgetSuggestions,
		AutoConfiguration: autoConfig,
		QualityScore:      s.calculateQualityScore(analysis),
		Confidence:        0.85, // Would be calculated based on analysis quality
	}

	return response, nil
}

// performAIAnalysis uses LLM for feed analysis
func (s *EnhancedRSSService) performAIAnalysis(feed *db.RSSFeed, request RSSAnalysisRequest) (*FeedAnalysis, error) {
	// Prepare analysis prompt
	prompt := s.buildAnalysisPrompt(feed, request)

	// Call LLM service (simplified - would use structured prompts in real implementation)
	analysisData, err := s.callLLMForAnalysis(prompt)
	if err != nil {
		return nil, err
	}

	// Parse and structure analysis results
	analysis := &FeedAnalysis{
		Title:           feed.Title,
		Description:     feed.Description,
		Language:        feed.Language,
		ItemCount:       len(feed.Items),
		Category:        s.inferCategory(feed),
		UpdateFrequency: s.analyzeUpdateFrequency(feed),
		ContentTypes:    s.identifyContentTypes(feed),
		Topics:          s.extractTopics(feed),
		Reliability:     s.assessReliability(feed),
		ContentQuality:  s.assessContentQuality(feed),
	}

	// Enhance with AI insights
	if analysisData != nil {
		s.enhanceWithAIInsights(analysis, analysisData)
	}

	return analysis, nil
}

// buildAnalysisPrompt creates LLM prompt for RSS analysis
func (s *EnhancedRSSService) buildAnalysisPrompt(feed *db.RSSFeed, request RSSAnalysisRequest) string {
	var builder strings.Builder

	builder.WriteString("Analyze this RSS feed for widget integration:\n\n")
	builder.WriteString(fmt.Sprintf("Feed Title: %s\n", feed.Title))
	builder.WriteString(fmt.Sprintf("Description: %s\n", feed.Description))
	builder.WriteString(fmt.Sprintf("Item Count: %d\n\n", len(feed.Items)))

	// Include sample items
	builder.WriteString("Sample Items:\n")
	maxSamples := 5
	if len(feed.Items) < maxSamples {
		maxSamples = len(feed.Items)
	}

	for i := 0; i < maxSamples; i++ {
		item := feed.Items[i]
		builder.WriteString(fmt.Sprintf("- Title: %s\n", item.Title))
		builder.WriteString(fmt.Sprintf("  Description: %s\n", s.truncateText(item.Description, 200)))
		builder.WriteString(fmt.Sprintf("  Date: %s\n\n", item.PubDate))
	}

	builder.WriteString("Analysis Requirements:\n")
	builder.WriteString(fmt.Sprintf("- Analysis Type: %s\n", request.AnalysisType))
	if request.UserIntent != "" {
		builder.WriteString(fmt.Sprintf("- User Intent: %s\n", request.UserIntent))
	}

	builder.WriteString("\nProvide analysis in JSON format with:\n")
	builder.WriteString("- Main topics and categories\n")
	builder.WriteString("- Content quality assessment\n")
	builder.WriteString("- Update frequency patterns\n")
	builder.WriteString("- Widget suitability recommendations\n")
	builder.WriteString("- Optimal configuration suggestions\n")

	return builder.String()
}

// callLLMForAnalysis makes LLM API call for analysis
func (s *EnhancedRSSService) callLLMForAnalysis(prompt string) (map[string]interface{}, error) {
	// This would use the actual LLM service
	// For now, return a mock response structure
	return map[string]interface{}{
		"category":   "technology",
		"topics":     []string{"software", "development", "AI"},
		"sentiment":  "positive",
		"quality":    "high",
		"frequency":  "daily",
		"confidence": 0.85,
	}, nil
}

// Helper methods for analysis

func (s *EnhancedRSSService) inferCategory(feed *db.RSSFeed) string {
	// Simple category inference based on title and description
	title := strings.ToLower(feed.Title)
	desc := strings.ToLower(feed.Description)

	categories := map[string][]string{
		"technology": {"tech", "software", "programming", "development", "AI", "computer"},
		"news":       {"news", "breaking", "latest", "update", "report"},
		"business":   {"business", "finance", "market", "economy", "startup"},
		"science":    {"science", "research", "study", "discovery", "innovation"},
		"lifestyle":  {"lifestyle", "health", "fitness", "food", "travel"},
	}

	for category, keywords := range categories {
		for _, keyword := range keywords {
			if strings.Contains(title, keyword) || strings.Contains(desc, keyword) {
				return category
			}
		}
	}

	return "general"
}

func (s *EnhancedRSSService) analyzeUpdateFrequency(feed *db.RSSFeed) string {
	if len(feed.Items) < 2 {
		return "unknown"
	}

	// Analyze time gaps between items
	avgGap := s.calculateAverageTimeBetweenItems(feed.Items)

	if avgGap < 3*time.Hour {
		return "frequent" // Multiple times per day
	} else if avgGap < 25*time.Hour {
		return "daily"
	} else if avgGap < 8*24*time.Hour {
		return "weekly"
	}

	return "infrequent"
}

func (s *EnhancedRSSService) calculateAverageTimeBetweenItems(items []db.RSSItem) time.Duration {
	if len(items) < 2 {
		return 0
	}

	var totalDuration time.Duration
	validPairs := 0

	for i := 0; i < len(items)-1; i++ {
		date1, err1 := s.parseItemDate(items[i].PubDate)
		date2, err2 := s.parseItemDate(items[i+1].PubDate)

		if err1 == nil && err2 == nil {
			if date1.After(date2) {
				totalDuration += date1.Sub(date2)
			} else {
				totalDuration += date2.Sub(date1)
			}
			validPairs++
		}
	}

	if validPairs == 0 {
		return 24 * time.Hour // Default to daily
	}

	return totalDuration / time.Duration(validPairs)
}

func (s *EnhancedRSSService) parseItemDate(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func (s *EnhancedRSSService) identifyContentTypes(feed *db.RSSFeed) []string {
	types := []string{}
	hasImages := false
	hasVideos := false
	hasText := false

	for _, item := range feed.Items {
		if item.ImageURL != "" {
			hasImages = true
		}
		if strings.Contains(strings.ToLower(item.Description), "video") {
			hasVideos = true
		}
		if len(item.Description) > 100 {
			hasText = true
		}
	}

	if hasText {
		types = append(types, "text")
	}
	if hasImages {
		types = append(types, "images")
	}
	if hasVideos {
		types = append(types, "videos")
	}

	if len(types) == 0 {
		types = append(types, "text")
	}

	return types
}

func (s *EnhancedRSSService) extractTopics(feed *db.RSSFeed) []TopicAnalysis {
	// Simple topic extraction - would use more sophisticated NLP in real implementation
	topicCounts := make(map[string]int)

	for _, item := range feed.Items {
		words := s.extractKeywords(item.Title + " " + item.Description)
		for _, word := range words {
			topicCounts[word]++
		}
	}

	topics := []TopicAnalysis{}
	for topic, count := range topicCounts {
		if count >= 2 { // Only include topics mentioned multiple times
			topics = append(topics, TopicAnalysis{
				Topic:      topic,
				Confidence: float64(count) / float64(len(feed.Items)),
				ItemCount:  count,
				Keywords:   []string{topic},
			})
		}
	}

	return topics
}

func (s *EnhancedRSSService) extractKeywords(text string) []string {
	// Simple keyword extraction
	words := strings.Fields(strings.ToLower(text))
	keywords := []string{}

	// Filter common words and extract meaningful keywords
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
	}

	for _, word := range words {
		cleaned := strings.Trim(word, ".,!?;:")
		if len(cleaned) > 3 && !stopWords[cleaned] {
			keywords = append(keywords, cleaned)
		}
	}

	return keywords
}

func (s *EnhancedRSSService) assessReliability(feed *db.RSSFeed) *ReliabilityMetrics {
	return &ReliabilityMetrics{
		UpdateConsistency: 0.8,  // Would analyze actual update patterns
		ContentFreshness:  0.9,  // Based on item dates
		SourceCredibility: 0.85, // Based on feed metadata and content analysis
		TechnicalQuality:  0.9,  // Based on RSS format quality
		OverallScore:      0.86,
	}
}

func (s *EnhancedRSSService) assessContentQuality(feed *db.RSSFeed) *ContentQualityMetrics {
	return &ContentQualityMetrics{
		Readability:      0.8,  // Would analyze text complexity
		InformationDepth: 0.75, // Based on content length and detail
		Uniqueness:       0.85, // Based on content similarity analysis
		Engagement:       0.7,  // Based on content style and structure
		OverallQuality:   0.77,
	}
}

func (s *EnhancedRSSService) enhanceWithAIInsights(analysis *FeedAnalysis, aiData map[string]interface{}) {
	// Enhance analysis with AI insights
	if category, ok := aiData["category"].(string); ok {
		analysis.Category = category
	}
	if frequency, ok := aiData["frequency"].(string); ok {
		analysis.UpdateFrequency = frequency
	}
}

func (s *EnhancedRSSService) analyzeContent(feed *db.RSSFeed) *ContentSummary {
	topics := []string{}
	keyTerms := []string{}

	// Extract main topics from analysis
	for _, item := range feed.Items[:min(5, len(feed.Items))] {
		words := s.extractKeywords(item.Title)
		topics = append(topics, words...)
	}

	// Deduplicate and get top topics
	topicCounts := make(map[string]int)
	for _, topic := range topics {
		topicCounts[topic]++
	}

	for topic, count := range topicCounts {
		if count >= 2 {
			keyTerms = append(keyTerms, topic)
		}
	}

	return &ContentSummary{
		MainTopics:    keyTerms[:min(5, len(keyTerms))],
		KeyTerms:      keyTerms,
		Sentiment:     "neutral", // Would use sentiment analysis
		ReadingLevel:  "intermediate",
		ContentLength: "medium",
		MediaTypes:    []string{"text"},
		SourceTypes:   []string{"articles"},
	}
}

func (s *EnhancedRSSService) generateRecommendations(feed *db.RSSFeed, analysis *FeedAnalysis, request RSSAnalysisRequest) *RSSRecommendations {
	return &RSSRecommendations{
		OptimalConfig: &db.RSSConfig{
			MaxItems:      10,
			CacheMinutes:  60, // Based on update frequency
			ItemFilter:    "latest",
			IncludeImage:  true,
			IncludeAuthor: true,
			DateFormat:    "Jan 2, 2006",
		},
		FilterSuggestions: []FilterSuggestion{
			{
				Type:       "keyword",
				Values:     []string{"breaking", "update"},
				Reasoning:  "Focus on important news updates",
				Impact:     "medium",
				Confidence: 0.8,
			},
		},
		DisplayOptions: []DisplayOption{
			{
				Template:    "rss_headlines",
				Layout:      "list",
				Fields:      []string{"title", "pub_date"},
				Description: "Simple headline list",
				Suitability: 0.9,
			},
		},
		UpdateSchedule: &UpdateSchedule{
			Frequency:     "hourly",
			PeakTimes:     []string{"09:00", "12:00", "17:00"},
			CacheDuration: 60,
			Reasoning:     "Based on typical news update patterns",
		},
		QualityImprovements: []string{
			"Add image support for better visual appeal",
			"Include author information for credibility",
			"Implement smart filtering for relevant content",
		},
	}
}

func (s *EnhancedRSSService) generateWidgetSuggestions(feed *db.RSSFeed, analysis *FeedAnalysis, request RSSAnalysisRequest) []WidgetSuggestion {
	suggestions := []WidgetSuggestion{
		{
			WidgetType:   "rss_headlines",
			TemplateType: "rss_headlines",
			Name:         feed.Title + " Headlines",
			Description:  "Display latest headlines from " + feed.Title,
			DataMapping: map[string]interface{}{
				"feed_title": "title",
				"items":      "items[*].title",
				"max_items":  5,
			},
			Configuration: map[string]interface{}{
				"show_dates": true,
				"max_length": 100,
				"style":      "compact",
			},
			Suitability: 0.9,
			Benefits:    []string{"Clean headline display", "Optimized for e-paper", "Automatic updates"},
		},
	}

	// Add content-specific suggestions
	if analysis.Category == "news" {
		suggestions = append(suggestions, WidgetSuggestion{
			WidgetType:   "rss_summary",
			TemplateType: "rss_summary",
			Name:         "News Summary",
			Description:  "Featured news article with summary",
			DataMapping: map[string]interface{}{
				"title":       "items[0].title",
				"description": "items[0].description",
				"pub_date":    "items[0].pub_date",
				"link":        "items[0].link",
			},
			Suitability: 0.85,
			Benefits:    []string{"In-depth content view", "Single article focus", "Rich metadata"},
		})
	}

	return suggestions
}

func (s *EnhancedRSSService) generateAutoConfiguration(feed *db.RSSFeed, analysis *FeedAnalysis, request RSSAnalysisRequest) *db.RSSConfig {
	config := &db.RSSConfig{
		MaxItems:      10,
		CacheMinutes:  30,
		IncludeImage:  true,
		IncludeAuthor: true,
		DateFormat:    "Jan 2, 2006",
	}

	// Adjust based on update frequency
	switch analysis.UpdateFrequency {
	case "frequent":
		config.CacheMinutes = 15
		config.MaxItems = 15
	case "daily":
		config.CacheMinutes = 60
		config.MaxItems = 10
	case "weekly":
		config.CacheMinutes = 240
		config.MaxItems = 5
	}

	// Adjust based on content type
	hasImages := false
	for _, contentType := range analysis.ContentTypes {
		if contentType == "images" {
			hasImages = true
			break
		}
	}
	config.IncludeImage = hasImages

	return config
}

func (s *EnhancedRSSService) calculateQualityScore(analysis *FeedAnalysis) float64 {
	if analysis.Reliability == nil || analysis.ContentQuality == nil {
		return 0.7 // Default score
	}

	return (analysis.Reliability.OverallScore + analysis.ContentQuality.OverallQuality) / 2
}

// Helper functions

func (s *EnhancedRSSService) truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
