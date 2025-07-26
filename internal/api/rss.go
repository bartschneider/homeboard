package api

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/bartosz/homeboard/internal/db"
)

// RSSService handles RSS feed fetching and parsing
type RSSService struct {
	httpClient *http.Client
	cache      map[string]*cachedFeed
}

// cachedFeed represents a cached RSS feed with expiration
type cachedFeed struct {
	feed      *db.RSSFeed
	expiresAt time.Time
}

// RSS XML structures for parsing
type rssXML struct {
	XMLName xml.Name   `xml:"rss"`
	Channel channelXML `xml:"channel"`
}

type channelXML struct {
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	Link        string    `xml:"link"`
	Language    string    `xml:"language"`
	LastBuild   string    `xml:"lastBuildDate"`
	Items       []itemXML `xml:"item"`
}

type itemXML struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Link        string `xml:"link"`
	Author      string `xml:"author"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"`
	Category    string `xml:"category"`
	Enclosure   struct {
		URL  string `xml:"url,attr"`
		Type string `xml:"type,attr"`
	} `xml:"enclosure"`
}

// NewRSSService creates a new RSS service
func NewRSSService() *RSSService {
	return &RSSService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache: make(map[string]*cachedFeed),
	}
}

// FetchFeed fetches and parses an RSS feed
func (s *RSSService) FetchFeed(feedURL string, config *db.RSSConfig) (*db.RSSFeed, error) {
	// Check cache first
	if cached, exists := s.cache[feedURL]; exists && time.Now().Before(cached.expiresAt) {
		return s.filterAndLimitItems(cached.feed, config), nil
	}

	// Fetch RSS content
	req, err := http.NewRequest("GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "E-Paper-Dashboard/1.0 RSS Reader")
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS feed returned status %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read RSS response: %w", err)
	}

	// Parse RSS XML
	feed, err := s.parseRSSXML(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS XML: %w", err)
	}

	// Cache the feed
	cacheMinutes := config.CacheMinutes
	if cacheMinutes <= 0 {
		cacheMinutes = 30 // default cache
	}
	s.cache[feedURL] = &cachedFeed{
		feed:      feed,
		expiresAt: time.Now().Add(time.Duration(cacheMinutes) * time.Minute),
	}

	return s.filterAndLimitItems(feed, config), nil
}

// parseRSSXML parses RSS XML content into RSSFeed struct
func (s *RSSService) parseRSSXML(xmlData []byte) (*db.RSSFeed, error) {
	var rss rssXML
	err := xml.Unmarshal(xmlData, &rss)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal RSS XML: %w", err)
	}

	feed := &db.RSSFeed{
		Title:       strings.TrimSpace(rss.Channel.Title),
		Description: strings.TrimSpace(rss.Channel.Description),
		Link:        strings.TrimSpace(rss.Channel.Link),
		Language:    strings.TrimSpace(rss.Channel.Language),
		LastBuild:   strings.TrimSpace(rss.Channel.LastBuild),
		Items:       make([]db.RSSItem, 0, len(rss.Channel.Items)),
	}

	for _, xmlItem := range rss.Channel.Items {
		item := db.RSSItem{
			Title:       strings.TrimSpace(xmlItem.Title),
			Description: s.stripHTMLTags(strings.TrimSpace(xmlItem.Description)),
			Link:        strings.TrimSpace(xmlItem.Link),
			Author:      strings.TrimSpace(xmlItem.Author),
			PubDate:     strings.TrimSpace(xmlItem.PubDate),
			GUID:        strings.TrimSpace(xmlItem.GUID),
			Category:    strings.TrimSpace(xmlItem.Category),
		}

		// Extract image URL from enclosure if available
		if xmlItem.Enclosure.URL != "" && strings.HasPrefix(xmlItem.Enclosure.Type, "image/") {
			item.ImageURL = xmlItem.Enclosure.URL
		}

		feed.Items = append(feed.Items, item)
	}

	// Sort items by publication date (newest first)
	s.sortItemsByDate(feed.Items)

	return feed, nil
}

// filterAndLimitItems applies filters and limits to RSS items
func (s *RSSService) filterAndLimitItems(feed *db.RSSFeed, config *db.RSSConfig) *db.RSSFeed {
	filtered := &db.RSSFeed{
		Title:       feed.Title,
		Description: feed.Description,
		Link:        feed.Link,
		Language:    feed.Language,
		LastBuild:   feed.LastBuild,
		Items:       make([]db.RSSItem, 0),
	}

	for _, item := range feed.Items {
		// Apply item filter
		if config.ItemFilter != "" && !s.matchesFilter(item, config.ItemFilter) {
			continue
		}

		// Remove unwanted fields based on config
		if !config.IncludeImage {
			item.ImageURL = ""
		}
		if !config.IncludeAuthor {
			item.Author = ""
		}

		// Format date if custom format specified
		if config.DateFormat != "" {
			item.PubDate = s.formatDate(item.PubDate, config.DateFormat)
		}

		filtered.Items = append(filtered.Items, item)

		// Apply max items limit
		maxItems := config.MaxItems
		if maxItems <= 0 {
			maxItems = 10 // default limit
		}
		if len(filtered.Items) >= maxItems {
			break
		}
	}

	return filtered
}

// matchesFilter checks if an RSS item matches the specified filter
func (s *RSSService) matchesFilter(item db.RSSItem, filter string) bool {
	switch strings.ToLower(filter) {
	case "latest":
		return true // No additional filtering for latest
	case "today":
		return s.isFromToday(item.PubDate)
	case "thisweek":
		return s.isFromThisWeek(item.PubDate)
	default:
		// Custom filter - check if filter text appears in title or description
		filterLower := strings.ToLower(filter)
		return strings.Contains(strings.ToLower(item.Title), filterLower) ||
			strings.Contains(strings.ToLower(item.Description), filterLower)
	}
}

// isFromToday checks if the publication date is from today
func (s *RSSService) isFromToday(pubDate string) bool {
	date, err := s.parseDate(pubDate)
	if err != nil {
		return false
	}
	now := time.Now()
	return date.Year() == now.Year() && date.YearDay() == now.YearDay()
}

// isFromThisWeek checks if the publication date is from this week
func (s *RSSService) isFromThisWeek(pubDate string) bool {
	date, err := s.parseDate(pubDate)
	if err != nil {
		return false
	}
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	return date.After(weekStart)
}

// parseDate parses various RSS date formats
func (s *RSSService) parseDate(dateStr string) (time.Time, error) {
	// Common RSS date formats
	formats := []string{
		time.RFC1123Z,               // Mon, 02 Jan 2006 15:04:05 -0700
		time.RFC1123,                // Mon, 02 Jan 2006 15:04:05 MST
		time.RFC822Z,                // 02 Jan 06 15:04 -0700
		time.RFC822,                 // 02 Jan 06 15:04 MST
		"2006-01-02T15:04:05Z07:00", // ISO 8601
		"2006-01-02T15:04:05Z",      // ISO 8601 UTC
		"2006-01-02 15:04:05",       // Simple format
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// formatDate formats a date string according to the specified format
func (s *RSSService) formatDate(dateStr, format string) string {
	date, err := s.parseDate(dateStr)
	if err != nil {
		return dateStr // Return original if parsing fails
	}
	return date.Format(format)
}

// sortItemsByDate sorts RSS items by publication date (newest first)
func (s *RSSService) sortItemsByDate(items []db.RSSItem) {
	sort.Slice(items, func(i, j int) bool {
		dateI, errI := s.parseDate(items[i].PubDate)
		dateJ, errJ := s.parseDate(items[j].PubDate)

		if errI != nil && errJ != nil {
			return false // Keep original order if both dates are invalid
		}
		if errI != nil {
			return false // Put invalid dates at the end
		}
		if errJ != nil {
			return true // Put invalid dates at the end
		}

		return dateI.After(dateJ) // Newest first
	})
}

// stripHTMLTags removes HTML tags from text content
func (s *RSSService) stripHTMLTags(html string) string {
	// Simple HTML tag removal (for basic cleanup)
	result := html
	for {
		start := strings.Index(result, "<")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], ">")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	return strings.TrimSpace(result)
}

// ValidateFeedURL validates an RSS feed URL by attempting to fetch and parse it
func (s *RSSService) ValidateFeedURL(feedURL string) error {
	config := &db.RSSConfig{
		MaxItems:     1,
		CacheMinutes: 0, // Don't cache validation requests
	}

	_, err := s.FetchFeed(feedURL, config)
	return err
}
