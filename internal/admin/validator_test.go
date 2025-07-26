package admin

import (
	"testing"

	"github.com/bartosz/homeboard/internal/config"
)

func TestConfigValidator(t *testing.T) {
	validator := NewConfigValidator()

	t.Run("ValidConfig", func(t *testing.T) {
		req := ConfigRequest{
			RefreshInterval: 15,
			ServerPort:      8081,
			Title:           "Test Dashboard",
			Theme: config.Theme{
				FontFamily: "serif",
				FontSize:   "16px",
				Background: "#ffffff",
				Foreground: "#000000",
			},
			Widgets: []config.Widget{
				{
					Name:    "Test Widget",
					Script:  "widgets/test.py",
					Enabled: true,
					Timeout: 10,
					Parameters: map[string]interface{}{
						"test_param": "value",
					},
				},
			},
		}

		result := validator.ValidateConfig(&req)

		if !result.Valid {
			t.Errorf("Expected valid config, but got errors: %v", result.Errors)
		}

		if len(result.Errors) > 0 {
			t.Errorf("Expected no errors, got %d", len(result.Errors))
		}
	})

	t.Run("InvalidRefreshInterval", func(t *testing.T) {
		req := ConfigRequest{
			RefreshInterval: -1, // Invalid
			ServerPort:      8081,
			Title:           "Test Dashboard",
			Theme: config.Theme{
				FontFamily: "serif",
				FontSize:   "16px",
				Background: "#ffffff",
				Foreground: "#000000",
			},
			Widgets: []config.Widget{},
		}

		result := validator.ValidateConfig(&req)

		if result.Valid {
			t.Error("Expected invalid config, but validation passed")
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "refresh_interval" && err.Code == "MIN_VALUE" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected refresh_interval MIN_VALUE error")
		}
	})

	t.Run("InvalidServerPort", func(t *testing.T) {
		req := ConfigRequest{
			RefreshInterval: 15,
			ServerPort:      100, // Invalid (too low)
			Title:           "Test Dashboard",
			Theme: config.Theme{
				FontFamily: "serif",
				FontSize:   "16px",
				Background: "#ffffff",
				Foreground: "#000000",
			},
			Widgets: []config.Widget{},
		}

		result := validator.ValidateConfig(&req)

		if result.Valid {
			t.Error("Expected invalid config, but validation passed")
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "server_port" && err.Code == "MIN_VALUE" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected server_port MIN_VALUE error")
		}
	})

	t.Run("EmptyTitle", func(t *testing.T) {
		req := ConfigRequest{
			RefreshInterval: 15,
			ServerPort:      8081,
			Title:           "", // Invalid (empty)
			Theme: config.Theme{
				FontFamily: "serif",
				FontSize:   "16px",
				Background: "#ffffff",
				Foreground: "#000000",
			},
			Widgets: []config.Widget{},
		}

		result := validator.ValidateConfig(&req)

		if result.Valid {
			t.Error("Expected invalid config, but validation passed")
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "title" && err.Code == "REQUIRED" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected title REQUIRED error")
		}
	})

	t.Run("InvalidTheme", func(t *testing.T) {
		req := ConfigRequest{
			RefreshInterval: 15,
			ServerPort:      8081,
			Title:           "Test Dashboard",
			Theme: config.Theme{
				FontFamily: "",              // Invalid (empty)
				FontSize:   "invalid-size",  // Invalid format
				Background: "not-a-color",   // Invalid color
				Foreground: "#gggggg",       // Invalid hex color
			},
			Widgets: []config.Widget{},
		}

		result := validator.ValidateConfig(&req)

		if result.Valid {
			t.Error("Expected invalid config, but validation passed")
		}

		expectedErrors := []struct {
			field string
			code  string
		}{
			{"theme.font_family", "REQUIRED"},
			{"theme.font_size", "INVALID_FORMAT"},
			{"theme.background", "INVALID_FORMAT"},
			{"theme.foreground", "INVALID_FORMAT"},
		}

		for _, expected := range expectedErrors {
			found := false
			for _, err := range result.Errors {
				if err.Field == expected.field && err.Code == expected.code {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error for field %s with code %s", expected.field, expected.code)
			}
		}
	})

	t.Run("DuplicateWidgetNames", func(t *testing.T) {
		req := ConfigRequest{
			RefreshInterval: 15,
			ServerPort:      8081,
			Title:           "Test Dashboard",
			Theme: config.Theme{
				FontFamily: "serif",
				FontSize:   "16px",
				Background: "#ffffff",
				Foreground: "#000000",
			},
			Widgets: []config.Widget{
				{
					Name:    "Duplicate Widget",
					Script:  "widgets/test1.py",
					Enabled: true,
					Timeout: 10,
				},
				{
					Name:    "Duplicate Widget", // Duplicate name
					Script:  "widgets/test2.py",
					Enabled: true,
					Timeout: 10,
				},
			},
		}

		result := validator.ValidateConfig(&req)

		if result.Valid {
			t.Error("Expected invalid config due to duplicate widget names")
		}

		found := false
		for _, err := range result.Errors {
			if err.Code == "DUPLICATE_NAME" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected DUPLICATE_NAME error for widget names")
		}
	})

	t.Run("InvalidWidgetScript", func(t *testing.T) {
		req := ConfigRequest{
			RefreshInterval: 15,
			ServerPort:      8081,
			Title:           "Test Dashboard",
			Theme: config.Theme{
				FontFamily: "serif",
				FontSize:   "16px",
				Background: "#ffffff",
				Foreground: "#000000",
			},
			Widgets: []config.Widget{
				{
					Name:    "Test Widget",
					Script:  "../../../etc/passwd", // Path traversal attempt
					Enabled: true,
					Timeout: 10,
				},
			},
		}

		result := validator.ValidateConfig(&req)

		if result.Valid {
			t.Error("Expected invalid config due to path traversal in script")
		}

		found := false
		for _, err := range result.Errors {
			if err.Code == "SECURITY_VIOLATION" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected SECURITY_VIOLATION error for path traversal")
		}
	})

	t.Run("InvalidWidgetTimeout", func(t *testing.T) {
		req := ConfigRequest{
			RefreshInterval: 15,
			ServerPort:      8081,
			Title:           "Test Dashboard",
			Theme: config.Theme{
				FontFamily: "serif",
				FontSize:   "16px",
				Background: "#ffffff",
				Foreground: "#000000",
			},
			Widgets: []config.Widget{
				{
					Name:    "Test Widget",
					Script:  "widgets/test.py",
					Enabled: true,
					Timeout: 500, // Too high
				},
			},
		}

		result := validator.ValidateConfig(&req)

		if result.Valid {
			t.Error("Expected invalid config due to high timeout")
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "widgets[0].timeout" && err.Code == "MAX_VALUE" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected timeout MAX_VALUE error")
		}
	})

	t.Run("Warnings", func(t *testing.T) {
		req := ConfigRequest{
			RefreshInterval: 2, // Will generate warning
			ServerPort:      80, // Will generate warning (reserved port)
			Title:           "Test <script>alert('xss')</script> Dashboard", // Will generate warning
			Theme: config.Theme{
				FontFamily: "serif",
				FontSize:   "16px",
				Background: "#cccccc", // Will generate warning (not pure white/black)
				Foreground: "#333333", // Will generate warning (not pure white/black)
			},
			Widgets: []config.Widget{},
		}

		result := validator.ValidateConfig(&req)

		if !result.Valid {
			t.Errorf("Expected valid config with warnings, but got errors: %v", result.Errors)
		}

		if len(result.Warnings) == 0 {
			t.Error("Expected warnings, but got none")
		}

		expectedWarnings := []string{
			"below 5 minutes",
			"reserved for other services",
			"HTML special characters",
			"E-paper displays work best",
		}

		for _, expectedWarning := range expectedWarnings {
			found := false
			for _, warning := range result.Warnings {
				if contains(warning, expectedWarning) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected warning containing '%s'", expectedWarning)
			}
		}
	})
}

func TestWidgetValidator(t *testing.T) {
	validator := NewConfigValidator()

	t.Run("ValidWidget", func(t *testing.T) {
		req := WidgetRequest{
			Name:    "Valid Widget",
			Script:  "widgets/valid.py",
			Enabled: true,
			Timeout: 15,
			Parameters: map[string]interface{}{
				"param1": "value1",
				"param2": 42,
				"param3": true,
			},
		}

		err := validator.ValidateWidget(&req)

		if err != nil {
			t.Errorf("Expected valid widget, but got error: %v", err)
		}
	})

	t.Run("InvalidWidgetName", func(t *testing.T) {
		req := WidgetRequest{
			Name:    "", // Invalid (empty)
			Script:  "widgets/test.py",
			Enabled: true,
			Timeout: 15,
		}

		err := validator.ValidateWidget(&req)

		if err == nil {
			t.Error("Expected validation error for empty widget name")
		}

		if !contains(err.Error(), "name") {
			t.Errorf("Expected error message to mention 'name', got: %v", err)
		}
	})

	t.Run("InvalidWidgetScript", func(t *testing.T) {
		req := WidgetRequest{
			Name:    "Test Widget",
			Script:  "", // Invalid (empty)
			Enabled: true,
			Timeout: 15,
		}

		err := validator.ValidateWidget(&req)

		if err == nil {
			t.Error("Expected validation error for empty script path")
		}

		if !contains(err.Error(), "script") {
			t.Errorf("Expected error message to mention 'script', got: %v", err)
		}
	})

	t.Run("ComplexParameters", func(t *testing.T) {
		req := WidgetRequest{
			Name:    "Complex Widget",
			Script:  "widgets/complex.py",
			Enabled: true,
			Timeout: 30,
			Parameters: map[string]interface{}{
				"simple_string": "value",
				"number":        42,
				"boolean":       true,
				"nested_object": map[string]interface{}{
					"nested_string": "nested_value",
					"nested_number": 123,
				},
				"array": []interface{}{
					"item1",
					"item2",
					map[string]interface{}{
						"nested_in_array": true,
					},
				},
			},
		}

		err := validator.ValidateWidget(&req)

		if err != nil {
			t.Errorf("Expected valid widget with complex parameters, but got error: %v", err)
		}
	})
}

func TestValidationEdgeCases(t *testing.T) {
	validator := NewConfigValidator()

	t.Run("MaxLengthTitle", func(t *testing.T) {
		// Create a title that's exactly at the limit
		longTitle := make([]byte, 100)
		for i := range longTitle {
			longTitle[i] = 'A'
		}

		req := ConfigRequest{
			RefreshInterval: 15,
			ServerPort:      8081,
			Title:           string(longTitle),
			Theme: config.Theme{
				FontFamily: "serif",
				FontSize:   "16px",
				Background: "#ffffff",
				Foreground: "#000000",
			},
			Widgets: []config.Widget{},
		}

		result := validator.ValidateConfig(&req)

		if !result.Valid {
			t.Errorf("Expected valid config with 100-char title, but got errors: %v", result.Errors)
		}

		// Now test with title that's too long
		req.Title = string(longTitle) + "X"
		result = validator.ValidateConfig(&req)

		if result.Valid {
			t.Error("Expected invalid config with 101-char title")
		}
	})

	t.Run("BoundaryValues", func(t *testing.T) {
		testCases := []struct {
			name            string
			refreshInterval int
			serverPort      int
			expectValid     bool
		}{
			{"MinValidValues", 1, 1024, true},
			{"MaxValidValues", 1440, 65535, true},
			{"BelowMinRefresh", 0, 1024, false},
			{"AboveMaxRefresh", 1441, 1024, false},
			{"BelowMinPort", 1, 1023, false},
			{"AboveMaxPort", 1, 65536, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := ConfigRequest{
					RefreshInterval: tc.refreshInterval,
					ServerPort:      tc.serverPort,
					Title:           "Test Dashboard",
					Theme: config.Theme{
						FontFamily: "serif",
						FontSize:   "16px",
						Background: "#ffffff",
						Foreground: "#000000",
					},
					Widgets: []config.Widget{},
				}

				result := validator.ValidateConfig(&req)

				if result.Valid != tc.expectValid {
					t.Errorf("Expected valid=%v, got valid=%v for case %s", tc.expectValid, result.Valid, tc.name)
				}
			})
		}
	})

	t.Run("ColorValidation", func(t *testing.T) {
		validColors := []string{"#ffffff", "#000000", "#123456", "#abc", "#ABC", "#FFF"}
		invalidColors := []string{"ffffff", "#gggggg", "#12345", "#1234567", "white", "rgb(255,255,255)"}

		for _, color := range validColors {
			req := ConfigRequest{
				RefreshInterval: 15,
				ServerPort:      8081,
				Title:           "Test Dashboard",
				Theme: config.Theme{
					FontFamily: "serif",
					FontSize:   "16px",
					Background: color,
					Foreground: "#000000",
				},
				Widgets: []config.Widget{},
			}

			result := validator.ValidateConfig(&req)
			if !result.Valid {
				t.Errorf("Expected valid config with color %s, but got errors: %v", color, result.Errors)
			}
		}

		for _, color := range invalidColors {
			req := ConfigRequest{
				RefreshInterval: 15,
				ServerPort:      8081,
				Title:           "Test Dashboard",
				Theme: config.Theme{
					FontFamily: "serif",
					FontSize:   "16px",
					Background: color,
					Foreground: "#000000",
				},
				Widgets: []config.Widget{},
			}

			result := validator.ValidateConfig(&req)
			if result.Valid {
				t.Errorf("Expected invalid config with color %s", color)
			}
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 indexOf(s, substr) >= 0)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}