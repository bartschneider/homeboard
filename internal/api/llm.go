package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bartosz/homeboard/internal/db"
)

// LLMService handles interactions with the Gemini API
type LLMService struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewLLMService creates a new LLM service
func NewLLMService(apiKey string) *LLMService {
	return &LLMService{
		apiKey:  apiKey,
		baseURL: "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GeminiRequest represents a request to the Gemini API
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

// GeminiContent represents content in a Gemini request
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part of Gemini content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse represents a response from the Gemini API
type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

// GeminiCandidate represents a candidate response
type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

// AnalyzeAPIData analyzes API data using the Gemini LLM
func (s *LLMService) AnalyzeAPIData(request db.LLMAnalyzeRequest) (*db.LLMAnalyzeResponse, error) {
	// First, fetch sample data from the API
	sampleData, err := s.fetchAPIData(request.APIURL, request.APIHeaders)
	if err != nil {
		return &db.LLMAnalyzeResponse{
			Error:       fmt.Sprintf("Failed to fetch API data: %v", err),
			APIData:     nil,
			DataMapping: s.generateFallbackMapping(request.WidgetTemplate),
			Reasoning:   "Could not fetch API data. Using fallback mapping based on template.",
		}, nil
	}

	// If LLM is not configured, use smart fallback analysis
	if !s.IsConfigured() {
		mapping := s.smartAnalyzeData(sampleData, request.WidgetTemplate)
		return &db.LLMAnalyzeResponse{
			APIData:     sampleData,
			DataMapping: mapping,
			Reasoning:   "LLM not configured. Using smart field detection based on data structure and template requirements.",
			Confidence:  0.7, // Medium confidence for smart analysis
		}, nil
	}

	// Create the prompt for Gemini
	prompt := s.buildAnalysisPrompt(request.WidgetTemplate, sampleData)

	// Call Gemini API
	geminiResponse, err := s.callGeminiAPI(prompt)
	if err != nil {
		// Fallback to smart analysis if LLM fails
		mapping := s.smartAnalyzeData(sampleData, request.WidgetTemplate)
		return &db.LLMAnalyzeResponse{
			APIData:     sampleData,
			DataMapping: mapping,
			Reasoning:   fmt.Sprintf("LLM API failed (%v). Using smart field detection as fallback.", err),
			Confidence:  0.6, // Lower confidence for fallback
		}, nil
	}

	// Parse the response
	mapping, suggestions, confidence, err := s.parseGeminiResponse(geminiResponse)
	if err != nil {
		// Fallback to smart analysis if parsing fails
		fallbackMapping := s.smartAnalyzeData(sampleData, request.WidgetTemplate)
		return &db.LLMAnalyzeResponse{
			APIData:     sampleData,
			DataMapping: fallbackMapping,
			Reasoning:   fmt.Sprintf("LLM response parsing failed (%v). Using smart field detection as fallback.", err),
			Confidence:  0.6,
		}, nil
	}

	return &db.LLMAnalyzeResponse{
		APIData:     sampleData,
		DataMapping: mapping,
		Suggestions: suggestions,
		Reasoning:   "AI successfully analyzed your API data and generated optimal field mappings.",
		Confidence:  confidence,
	}, nil
}

// fetchAPIData fetches sample data from the API
func (s *LLMService) fetchAPIData(apiURL string, headers map[string]string) (interface{}, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Add a user agent
	req.Header.Set("User-Agent", "E-Paper-Dashboard/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// buildAnalysisPrompt creates a detailed prompt for the Gemini API
func (s *LLMService) buildAnalysisPrompt(templateType string, apiData interface{}) string {
	dataJSON, _ := json.MarshalIndent(apiData, "", "  ")

	// Get template information
	template := s.getTemplateInfo(templateType)

	prompt := fmt.Sprintf(`You are an expert at mapping API data to widget templates for an e-paper dashboard.

Task: Analyze the provided API response and create a mapping configuration for a "%s" widget template.

Widget Template: %s
Template Fields: %s

API Response Data:
%s

Requirements:
1. Create a JSON mapping object that maps template fields to JSON paths in the API data
2. Use JSONPath notation (e.g., "data.temperature", "items[0].name", "weather[0].description")
3. Prioritize the most relevant and useful data from the API response
4. Ensure all required fields are mapped
5. Provide alternative suggestions if multiple good options exist
6. Consider data types and formatting

Response Format (JSON only, no markdown):
{
  "dataMapping": {
    "field_name": "json.path.to.data",
    ...
  },
  "suggestions": [
    {
      "field": "field_name",
      "jsonPath": "alternative.path",
      "value": "sample_value",
      "confidence": 0.8,
      "description": "Why this mapping makes sense"
    }
  ],
  "confidence": 0.9
}

Focus on creating practical, useful mappings that will display meaningful information on an e-paper screen.`,
		templateType, template.Description, s.formatTemplateFields(template.Fields), string(dataJSON))

	return prompt
}

// getTemplateInfo returns template information by type
func (s *LLMService) getTemplateInfo(templateType string) db.WidgetTemplate {
	templates := GetWidgetTemplates()
	for _, template := range templates {
		if template.Type == templateType {
			return template
		}
	}

	// Return a basic template if not found
	return db.WidgetTemplate{
		Type:        templateType,
		Description: "Generic widget template",
		Fields: []db.WidgetTemplateField{
			{Key: "title", Label: "Title", Type: "text", Required: true},
			{Key: "value", Label: "Value", Type: "text", Required: true},
		},
	}
}

// formatTemplateFields formats template fields for the prompt
func (s *LLMService) formatTemplateFields(fields []db.WidgetTemplateField) string {
	var fieldDescriptions []string
	for _, field := range fields {
		required := ""
		if field.Required {
			required = " (required)"
		}
		fieldDescriptions = append(fieldDescriptions,
			fmt.Sprintf("- %s (%s): %s%s", field.Key, field.Type, field.Description, required))
	}
	return strings.Join(fieldDescriptions, "\n")
}

// callGeminiAPI makes a request to the Gemini API
func (s *LLMService) callGeminiAPI(prompt string) (*GeminiResponse, error) {
	requestBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s?key=%s", s.baseURL, s.apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Gemini API error %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, err
	}

	return &geminiResp, nil
}

// parseGeminiResponse parses the response from Gemini
func (s *LLMService) parseGeminiResponse(resp *GeminiResponse) (
	map[string]interface{},
	[]db.MappingSuggestion,
	float64,
	error,
) {
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, nil, 0, fmt.Errorf("empty response from Gemini")
	}

	responseText := resp.Candidates[0].Content.Parts[0].Text

	// Clean up the response text (remove markdown formatting if present)
	responseText = strings.TrimSpace(responseText)
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	// Parse the JSON response
	var parsedResponse struct {
		DataMapping map[string]interface{} `json:"dataMapping"`
		Suggestions []db.MappingSuggestion `json:"suggestions"`
		Confidence  float64                `json:"confidence"`
	}

	if err := json.Unmarshal([]byte(responseText), &parsedResponse); err != nil {
		return nil, nil, 0, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return parsedResponse.DataMapping, parsedResponse.Suggestions, parsedResponse.Confidence, nil
}

// IsConfigured returns true if the LLM service has an API key
func (s *LLMService) IsConfigured() bool {
	return s.apiKey != ""
}

// smartAnalyzeData performs intelligent field mapping without LLM
func (s *LLMService) smartAnalyzeData(data interface{}, templateType string) map[string]interface{} {
	template := s.getTemplateInfo(templateType)
	mapping := make(map[string]interface{})

	// Convert data to map for analysis
	dataMap, ok := s.convertToMap(data)
	if !ok {
		return s.generateFallbackMapping(templateType)
	}

	// Smart field matching based on template requirements
	for _, field := range template.Fields {
		bestMatch := s.findBestFieldMatch(field, dataMap)
		if bestMatch != "" {
			mapping[field.Key] = bestMatch
		}
	}

	// Fill missing required fields with intelligent defaults
	s.fillMissingFields(mapping, template)

	return mapping
}

// generateFallbackMapping creates basic mapping based on template
func (s *LLMService) generateFallbackMapping(templateType string) map[string]interface{} {
	template := s.getTemplateInfo(templateType)
	mapping := make(map[string]interface{})

	// Provide sensible defaults for each template type
	for _, field := range template.Fields {
		switch field.Key {
		case "title", "name", "label":
			mapping[field.Key] = "title || name || label"
		case "value", "amount", "count":
			mapping[field.Key] = "value || amount || count"
		case "description", "summary":
			mapping[field.Key] = "description || summary || info"
		case "status":
			mapping[field.Key] = "status || state || condition"
		case "timestamp", "time", "date":
			mapping[field.Key] = "timestamp || time || date || created_at"
		case "unit", "units":
			mapping[field.Key] = "unit || units || currency"
		case "progress", "percentage":
			mapping[field.Key] = "progress || percentage || percent"
		default:
			mapping[field.Key] = fmt.Sprintf("data.%s", field.Key)
		}
	}

	return mapping
}

// convertToMap converts interface{} to map[string]interface{} for analysis
func (s *LLMService) convertToMap(data interface{}) (map[string]interface{}, bool) {
	switch v := data.(type) {
	case map[string]interface{}:
		return v, true
	case []interface{}:
		if len(v) > 0 {
			if first, ok := v[0].(map[string]interface{}); ok {
				return first, true
			}
		}
		return nil, false
	default:
		return nil, false
	}
}

// findBestFieldMatch finds the best matching field in data for a template field
func (s *LLMService) findBestFieldMatch(field db.WidgetTemplateField, dataMap map[string]interface{}) string {
	// Direct key match (exact)
	if _, exists := dataMap[field.Key]; exists {
		return field.Key
	}

	// Synonyms for common field types
	synonyms := map[string][]string{
		"title":       {"name", "label", "heading", "subject"},
		"value":       {"amount", "count", "number", "data", "result"},
		"description": {"summary", "info", "details", "text", "body"},
		"status":      {"state", "condition", "health", "level"},
		"timestamp":   {"time", "date", "created_at", "updated_at", "datetime"},
		"unit":        {"units", "currency", "symbol", "suffix"},
		"progress":    {"percentage", "percent", "completion", "ratio"},
	}

	// Check synonyms
	if syns, exists := synonyms[field.Key]; exists {
		for _, syn := range syns {
			if _, exists := dataMap[syn]; exists {
				return syn
			}
		}
	}

	// Fuzzy matching for partial matches
	for key := range dataMap {
		if s.fuzzyMatch(field.Key, key) {
			return key
		}
	}

	return ""
}

// fuzzyMatch performs basic fuzzy matching between two strings
func (s *LLMService) fuzzyMatch(target, candidate string) bool {
	target = strings.ToLower(target)
	candidate = strings.ToLower(candidate)

	// Check if target is substring of candidate or vice versa
	if strings.Contains(candidate, target) || strings.Contains(target, candidate) {
		return true
	}

	// Check for common patterns
	patterns := map[string][]string{
		"temp":     {"temperature", "temp_c", "temp_f"},
		"humidity": {"humid", "moisture", "rh"},
		"pressure": {"press", "atm", "bar"},
		"speed":    {"velocity", "rate", "mph", "kmh"},
	}

	for pattern, matches := range patterns {
		if strings.Contains(target, pattern) {
			for _, match := range matches {
				if strings.Contains(candidate, match) {
					return true
				}
			}
		}
	}

	return false
}

// fillMissingFields fills in missing required fields with best guesses
func (s *LLMService) fillMissingFields(mapping map[string]interface{}, template db.WidgetTemplate) {
	for _, field := range template.Fields {
		if _, exists := mapping[field.Key]; !exists && field.Required {
			// Use placeholder as fallback
			if field.Placeholder != "" {
				mapping[field.Key] = field.Placeholder
			} else {
				mapping[field.Key] = fmt.Sprintf("data.%s", field.Key)
			}
		}
	}
}
