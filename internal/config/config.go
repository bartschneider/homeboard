package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents the main configuration structure
type Config struct {
	RefreshInterval int      `json:"refresh_interval"` // in minutes
	ServerPort      int      `json:"server_port"`
	Widgets         []Widget `json:"widgets"`
	Title           string   `json:"title"`
	Theme           Theme    `json:"theme"`
}

// Widget represents a single widget configuration
type Widget struct {
	Name       string                 `json:"name"`
	Script     string                 `json:"script"`
	Parameters map[string]interface{} `json:"parameters"`
	Timeout    int                    `json:"timeout"` // in seconds
	Enabled    bool                   `json:"enabled"`
}

// Theme represents display theme configuration
type Theme struct {
	FontFamily string `json:"font_family"`
	FontSize   string `json:"font_size"`
	Background string `json:"background"`
	Foreground string `json:"foreground"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		RefreshInterval: 15,
		ServerPort:      8080,
		Title:           "E-Paper Dashboard",
		Theme: Theme{
			FontFamily: "serif",
			FontSize:   "16px",
			Background: "#ffffff",
			Foreground: "#000000",
		},
		Widgets: []Widget{
			{
				Name:    "Clock",
				Script:  "widgets/clock.py",
				Enabled: true,
				Timeout: 10,
				Parameters: map[string]interface{}{
					"format":   "2006-01-02 15:04:05",
					"timezone": "Local",
				},
			},
			{
				Name:    "System Status",
				Script:  "widgets/system.py",
				Enabled: true,
				Timeout: 15,
				Parameters: map[string]interface{}{
					"show_cpu":    true,
					"show_memory": true,
					"show_disk":   true,
				},
			},
		},
	}
}

// LoadConfig loads configuration from a JSON file
func LoadConfig(configPath string) (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config if it doesn't exist
		config := DefaultConfig()
		if err := SaveConfig(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return config, nil
	}

	// Read existing config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Validate and set defaults for missing fields
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to a JSON file
func SaveConfig(config *Config, configPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// validateConfig validates and sets defaults for configuration
func validateConfig(config *Config) error {
	// Set defaults for missing values
	if config.RefreshInterval <= 0 {
		config.RefreshInterval = 15
	}
	if config.ServerPort <= 0 {
		config.ServerPort = 8080
	}
	if config.Title == "" {
		config.Title = "E-Paper Dashboard"
	}
	if config.Theme.FontFamily == "" {
		config.Theme.FontFamily = "serif"
	}
	if config.Theme.FontSize == "" {
		config.Theme.FontSize = "16px"
	}
	if config.Theme.Background == "" {
		config.Theme.Background = "#ffffff"
	}
	if config.Theme.Foreground == "" {
		config.Theme.Foreground = "#000000"
	}

	// Validate widgets
	for i := range config.Widgets {
		widget := &config.Widgets[i]
		if widget.Name == "" {
			return fmt.Errorf("widget %d: name cannot be empty", i)
		}
		if widget.Script == "" {
			return fmt.Errorf("widget %s: script path cannot be empty", widget.Name)
		}
		if widget.Timeout <= 0 {
			widget.Timeout = 30 // Default 30 seconds
		}
		if widget.Parameters == nil {
			widget.Parameters = make(map[string]interface{})
		}
	}

	return nil
}

// GetEnabledWidgets returns only enabled widgets
func (c *Config) GetEnabledWidgets() []Widget {
	var enabled []Widget
	for _, widget := range c.Widgets {
		if widget.Enabled {
			enabled = append(enabled, widget)
		}
	}
	return enabled
}

// GetRefreshDuration returns refresh interval as time.Duration
func (c *Config) GetRefreshDuration() time.Duration {
	return time.Duration(c.RefreshInterval) * time.Minute
}

// GetServerAddress returns the full server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf(":%d", c.ServerPort)
}
