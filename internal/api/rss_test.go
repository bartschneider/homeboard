package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bartosz/homeboard/internal/db"
)

// Mock RSS feed XML for testing
const mockRSSXML = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Test News Feed</title>
		<description>A test RSS feed for unit testing</description>
		<link>https://example.com</link>
		<language>en-us</language>
		<lastBuildDate>Mon, 15 Jan 2024 10:30:00 +0000</lastBuildDate>
		<item>
			<title>Breaking News: Technology Breakthrough</title>
			<description>Scientists have made a significant breakthrough in quantum computing technology that could revolutionize the industry.</description>
			<link>https://example.com/article1</link>
			<author>Jane Smith</author>
			<pubDate>Mon, 15 Jan 2024 10:00:00 +0000</pubDate>
			<guid>article-1</guid>
			<category>Technology</category>
			<enclosure url="https://example.com/image1.jpg" type="image/jpeg" />
		</item>
		<item>
			<title>Market Update: Tech Stocks Rally</title>
			<description>Technology stocks showed strong performance in today's trading session, with several companies posting significant gains.</description>
			<link>https://example.com/article2</link>
			<author>John Doe</author>
			<pubDate>Mon, 15 Jan 2024 09:30:00 +0000</pubDate>
			<guid>article-2</guid>
			<category>Finance</category>
		</item>
		<item>
			<title>Weather Alert: Storm Approaching</title>
			<description>Meteorologists warn of a severe storm system approaching the region with high winds and heavy rainfall expected.</description>
			<link>https://example.com/article3</link>
			<author>Weather Team</author>
			<pubDate>Mon, 15 Jan 2024 08:00:00 +0000</pubDate>
			<guid>article-3</guid>
			<category>Weather</category>
		</item>
	</channel>
</rss>`

const invalidRSSXML = `<?xml version="1.0" encoding="UTF-8"?>
<invalid>
	<not-rss>This is not a valid RSS feed</not-rss>
</invalid>`

func TestNewRSSService(t *testing.T) {
	service := NewRSSService()
	if service == nil {
		t.Fatal("NewRSSService() returned nil")
	}
	if service.httpClient == nil {
		t.Error("RSS service should have an HTTP client")
	}
	if service.cache == nil {
		t.Error("RSS service should have a cache")
	}
}

func TestParseRSSXML(t *testing.T) {
	service := NewRSSService()

	feed, err := service.parseRSSXML([]byte(mockRSSXML))
	if err != nil {
		t.Fatalf("Failed to parse valid RSS XML: %v", err)
	}

	// Test feed metadata
	if feed.Title != "Test News Feed" {
		t.Errorf("Expected title 'Test News Feed', got '%s'", feed.Title)
	}
	if feed.Description != "A test RSS feed for unit testing" {
		t.Errorf("Expected specific description, got '%s'", feed.Description)
	}
	if feed.Link != "https://example.com" {
		t.Errorf("Expected link 'https://example.com', got '%s'", feed.Link)
	}
	if feed.Language != "en-us" {
		t.Errorf("Expected language 'en-us', got '%s'", feed.Language)
	}

	// Test items count
	if len(feed.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(feed.Items))
	}

	// Test first item details
	item := feed.Items[0]
	if item.Title != "Breaking News: Technology Breakthrough" {
		t.Errorf("Expected specific title, got '%s'", item.Title)
	}
	if item.Author != "Jane Smith" {
		t.Errorf("Expected author 'Jane Smith', got '%s'", item.Author)
	}
	if item.Category != "Technology" {
		t.Errorf("Expected category 'Technology', got '%s'", item.Category)
	}
	if item.ImageURL != "https://example.com/image1.jpg" {
		t.Errorf("Expected image URL, got '%s'", item.ImageURL)
	}
}

func TestParseInvalidRSSXML(t *testing.T) {
	service := NewRSSService()

	_, err := service.parseRSSXML([]byte(invalidRSSXML))
	if err == nil {
		t.Error("Expected error when parsing invalid RSS XML")
	}
}

func TestFetchFeedWithMockServer(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockRSSXML))
	}))
	defer server.Close()

	service := NewRSSService()
	config := &db.RSSConfig{
		MaxItems:     10,
		CacheMinutes: 30,
	}

	feed, err := service.FetchFeed(server.URL, config)
	if err != nil {
		t.Fatalf("Failed to fetch feed: %v", err)
	}

	if feed.Title != "Test News Feed" {
		t.Errorf("Expected title 'Test News Feed', got '%s'", feed.Title)
	}
	if len(feed.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(feed.Items))
	}
}

func TestFetchFeedWithServerError(t *testing.T) {
	// Create mock HTTP server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	service := NewRSSService()
	config := &db.RSSConfig{
		MaxItems:     10,
		CacheMinutes: 30,
	}

	_, err := service.FetchFeed(server.URL, config)
	if err == nil {
		t.Error("Expected error when server returns 500")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Expected error to mention status code 500, got: %v", err)
	}
}

func TestFilterAndLimitItems(t *testing.T) {
	service := NewRSSService()

	// Create test feed
	feed := &db.RSSFeed{
		Title: "Test Feed",
		Items: []db.RSSItem{
			{Title: "Item 1", Description: "Tech news", Author: "Author 1", ImageURL: "image1.jpg"},
			{Title: "Item 2", Description: "Sports news", Author: "Author 2", ImageURL: "image2.jpg"},
			{Title: "Item 3", Description: "Weather news", Author: "Author 3", ImageURL: "image3.jpg"},
		},
	}

	// Test max items limit
	config := &db.RSSConfig{
		MaxItems:      2,
		IncludeImage:  true,
		IncludeAuthor: true,
	}

	filtered := service.filterAndLimitItems(feed, config)
	if len(filtered.Items) != 2 {
		t.Errorf("Expected 2 items after filtering, got %d", len(filtered.Items))
	}

	// Test image exclusion
	config.IncludeImage = false
	filtered = service.filterAndLimitItems(feed, config)
	for _, item := range filtered.Items {
		if item.ImageURL != "" {
			t.Error("Expected image URLs to be removed when IncludeImage is false")
		}
	}

	// Test author exclusion
	config.IncludeAuthor = false
	filtered = service.filterAndLimitItems(feed, config)
	for _, item := range filtered.Items {
		if item.Author != "" {
			t.Error("Expected authors to be removed when IncludeAuthor is false")
		}
	}
}

func TestMatchesFilter(t *testing.T) {
	service := NewRSSService()

	item := db.RSSItem{
		Title:       "Technology News Update",
		Description: "Latest developments in artificial intelligence",
		PubDate:     time.Now().Format(time.RFC1123Z),
	}

	// Test custom filter (case insensitive)
	if !service.matchesFilter(item, "technology") {
		t.Error("Expected item to match 'technology' filter")
	}
	if !service.matchesFilter(item, "artificial") {
		t.Error("Expected item to match 'artificial' filter")
	}
	if service.matchesFilter(item, "sports") {
		t.Error("Expected item NOT to match 'sports' filter")
	}

	// Test latest filter (should always match)
	if !service.matchesFilter(item, "latest") {
		t.Error("Expected item to match 'latest' filter")
	}
}

func TestParseDate(t *testing.T) {
	service := NewRSSService()

	testCases := []struct {
		input    string
		expected bool // whether parsing should succeed
	}{
		{"Mon, 15 Jan 2024 10:00:00 +0000", true},
		{"Mon, 15 Jan 2024 10:00:00 GMT", true},
		{"15 Jan 24 10:00 +0000", true},
		{"2024-01-15T10:00:00Z", true},
		{"2024-01-15 10:00:00", true},
		{"invalid date format", false},
		{"", false},
	}

	for _, tc := range testCases {
		_, err := service.parseDate(tc.input)
		if tc.expected && err != nil {
			t.Errorf("Expected date '%s' to parse successfully, got error: %v", tc.input, err)
		}
		if !tc.expected && err == nil {
			t.Errorf("Expected date '%s' to fail parsing, but it succeeded", tc.input)
		}
	}
}

func TestFormatDate(t *testing.T) {
	service := NewRSSService()

	input := "Mon, 15 Jan 2024 10:00:00 +0000"
	format := "2006-01-02"

	result := service.formatDate(input, format)
	if result != "2024-01-15" {
		t.Errorf("Expected formatted date '2024-01-15', got '%s'", result)
	}

	// Test with invalid date (should return original)
	invalidInput := "invalid date"
	result = service.formatDate(invalidInput, format)
	if result != invalidInput {
		t.Errorf("Expected original date to be returned for invalid input, got '%s'", result)
	}
}

func TestStripHTMLTags(t *testing.T) {
	service := NewRSSService()

	testCases := []struct {
		input    string
		expected string
	}{
		{"<p>Hello <b>world</b></p>", "Hello world"},
		{"Plain text", "Plain text"},
		{"<div><span>Nested</span> tags</div>", "Nested tags"},
		{"<img src='test.jpg'>Image tag", "Image tag"},
		{"", ""},
	}

	for _, tc := range testCases {
		result := service.stripHTMLTags(tc.input)
		if result != tc.expected {
			t.Errorf("stripHTMLTags('%s') = '%s', expected '%s'", tc.input, result, tc.expected)
		}
	}
}

func TestRSSCaching(t *testing.T) {
	// Create mock server
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockRSSXML))
	}))
	defer server.Close()

	service := NewRSSService()
	config := &db.RSSConfig{
		MaxItems:     10,
		CacheMinutes: 1, // 1 minute cache
	}

	// First call should hit the server
	_, err := service.FetchFeed(server.URL, config)
	if err != nil {
		t.Fatalf("First fetch failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 server call, got %d", callCount)
	}

	// Second call should use cache
	_, err = service.FetchFeed(server.URL, config)
	if err != nil {
		t.Fatalf("Second fetch failed: %v", err)
	}
	if callCount != 1 {
		t.Errorf("Expected 1 server call (cached), got %d", callCount)
	}
}

func TestValidateFeedURL(t *testing.T) {
	// Create mock server with valid RSS
	validServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockRSSXML))
	}))
	defer validServer.Close()

	// Create mock server with invalid RSS
	invalidServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not rss"))
	}))
	defer invalidServer.Close()

	service := NewRSSService()

	// Test valid feed
	err := service.ValidateFeedURL(validServer.URL)
	if err != nil {
		t.Errorf("Expected valid feed to pass validation, got error: %v", err)
	}

	// Test invalid feed
	err = service.ValidateFeedURL(invalidServer.URL)
	if err == nil {
		t.Error("Expected invalid feed to fail validation")
	}

	// Test non-existent URL
	err = service.ValidateFeedURL("http://non-existent-url.example")
	if err == nil {
		t.Error("Expected non-existent URL to fail validation")
	}
}

// Benchmark tests
func BenchmarkParseRSSXML(b *testing.B) {
	service := NewRSSService()
	xmlData := []byte(mockRSSXML)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.parseRSSXML(xmlData)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}

func BenchmarkStripHTMLTags(b *testing.B) {
	service := NewRSSService()
	html := "<p>This is a <b>test</b> with <i>HTML</i> tags that need to be <a href='#'>stripped</a>.</p>"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.stripHTMLTags(html)
	}
}
