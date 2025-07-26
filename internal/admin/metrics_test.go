package admin

import (
	"testing"
	"time"
)

func TestMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()

	t.Run("InitialState", func(t *testing.T) {
		if collector.startTime.IsZero() {
			t.Error("Expected start time to be set")
		}

		if collector.maxLogEntries != 1000 {
			t.Errorf("Expected max log entries to be 1000, got %d", collector.maxLogEntries)
		}

		status := collector.GetSystemStatus()
		if status.Status != "running" {
			t.Errorf("Expected initial status to be 'running', got '%s'", status.Status)
		}
	})

	t.Run("SystemMetricsUpdate", func(t *testing.T) {
		collector.UpdateSystemMetrics()
		
		metrics := collector.GetAllMetrics()
		if _, exists := metrics["system_metrics"]; !exists {
			t.Error("Expected 'system_metrics' to exist")
		}

		if _, exists := metrics["widget_metrics"]; !exists {
			t.Error("Expected 'widget_metrics' to exist")
		}

		if _, exists := metrics["uptime"]; !exists {
			t.Error("Expected 'uptime' to exist")
		}
	})

	t.Run("RequestCounters", func(t *testing.T) {
		initialRequests := collector.requestCount
		initialErrors := collector.errorCount

		collector.IncrementRequestCount()
		collector.IncrementRequestCount()
		collector.IncrementErrorCount()

		if collector.requestCount != initialRequests+2 {
			t.Errorf("Expected request count to be %d, got %d", initialRequests+2, collector.requestCount)
		}

		if collector.errorCount != initialErrors+1 {
			t.Errorf("Expected error count to be %d, got %d", initialErrors+1, collector.errorCount)
		}
	})
}

func TestWidgetMetrics(t *testing.T) {
	collector := NewMetricsCollector()

	t.Run("RecordSuccessfulExecution", func(t *testing.T) {
		widgetName := "test_widget"
		duration := 150 * time.Millisecond
		output := "Widget executed successfully"

		collector.RecordWidgetExecution(widgetName, duration, true, output, "")

		metrics := collector.GetWidgetMetrics(widgetName)
		if metrics == nil {
			t.Fatal("Expected widget metrics to exist")
		}

		if metrics.WidgetName != widgetName {
			t.Errorf("Expected widget name '%s', got '%s'", widgetName, metrics.WidgetName)
		}

		if metrics.Status != "active" {
			t.Errorf("Expected status 'active', got '%s'", metrics.Status)
		}

		if metrics.ExecutionCount != 1 {
			t.Errorf("Expected execution count 1, got %d", metrics.ExecutionCount)
		}

		if metrics.SuccessCount != 1 {
			t.Errorf("Expected success count 1, got %d", metrics.SuccessCount)
		}

		if metrics.ErrorCount != 0 {
			t.Errorf("Expected error count 0, got %d", metrics.ErrorCount)
		}

		if metrics.LastOutput != output {
			t.Errorf("Expected last output '%s', got '%s'", output, metrics.LastOutput)
		}

		if metrics.ExecutionTime != duration {
			t.Errorf("Expected execution time %v, got %v", duration, metrics.ExecutionTime)
		}
	})

	t.Run("RecordFailedExecution", func(t *testing.T) {
		widgetName := "failing_widget"
		duration := 100 * time.Millisecond
		errorMsg := "Widget execution failed"

		collector.RecordWidgetExecution(widgetName, duration, false, "", errorMsg)

		metrics := collector.GetWidgetMetrics(widgetName)
		if metrics == nil {
			t.Fatal("Expected widget metrics to exist")
		}

		if metrics.Status != "error" {
			t.Errorf("Expected status 'error', got '%s'", metrics.Status)
		}

		if metrics.ErrorCount != 1 {
			t.Errorf("Expected error count 1, got %d", metrics.ErrorCount)
		}

		if metrics.ErrorMessage != errorMsg {
			t.Errorf("Expected error message '%s', got '%s'", errorMsg, metrics.ErrorMessage)
		}
	})

	t.Run("MultipleExecutions", func(t *testing.T) {
		widgetName := "multi_widget"
		
		// Record multiple executions with different durations
		durations := []time.Duration{
			100 * time.Millisecond,
			200 * time.Millisecond,
			150 * time.Millisecond,
		}

		for _, duration := range durations {
			collector.RecordWidgetExecution(widgetName, duration, true, "output", "")
		}

		metrics := collector.GetWidgetMetrics(widgetName)
		if metrics == nil {
			t.Fatal("Expected widget metrics to exist")
		}

		if metrics.ExecutionCount != 3 {
			t.Errorf("Expected execution count 3, got %d", metrics.ExecutionCount)
		}

		if metrics.MinTime != 100*time.Millisecond {
			t.Errorf("Expected min time 100ms, got %v", metrics.MinTime)
		}

		if metrics.MaxTime != 200*time.Millisecond {
			t.Errorf("Expected max time 200ms, got %v", metrics.MaxTime)
		}

		expectedAverage := 150 * time.Millisecond
		if metrics.AverageTime != expectedAverage {
			t.Errorf("Expected average time %v, got %v", expectedAverage, metrics.AverageTime)
		}
	})

	t.Run("ResetWidgetMetrics", func(t *testing.T) {
		widgetName := "reset_widget"
		
		// Record some executions
		collector.RecordWidgetExecution(widgetName, 100*time.Millisecond, true, "output", "")
		collector.RecordWidgetExecution(widgetName, 200*time.Millisecond, false, "", "error")

		// Verify metrics exist
		metrics := collector.GetWidgetMetrics(widgetName)
		if metrics == nil {
			t.Fatal("Expected widget metrics to exist before reset")
		}

		if metrics.ExecutionCount == 0 {
			t.Error("Expected non-zero execution count before reset")
		}

		// Reset metrics
		collector.ResetWidgetMetrics(widgetName)

		// Verify metrics are reset
		metrics = collector.GetWidgetMetrics(widgetName)
		if metrics == nil {
			t.Fatal("Expected widget metrics to exist after reset")
		}

		if metrics.ExecutionCount != 0 {
			t.Errorf("Expected execution count 0 after reset, got %d", metrics.ExecutionCount)
		}

		if metrics.SuccessCount != 0 {
			t.Errorf("Expected success count 0 after reset, got %d", metrics.SuccessCount)
		}

		if metrics.ErrorCount != 0 {
			t.Errorf("Expected error count 0 after reset, got %d", metrics.ErrorCount)
		}

		if metrics.Status != "idle" {
			t.Errorf("Expected status 'idle' after reset, got '%s'", metrics.Status)
		}
	})

	t.Run("GetWidgetStatus", func(t *testing.T) {
		widgetName := "status_widget"
		
		// Record execution
		collector.RecordWidgetExecution(widgetName, 100*time.Millisecond, true, "test output", "")

		status := collector.GetWidgetStatus(widgetName)
		if status.WidgetName != widgetName {
			t.Errorf("Expected widget name '%s', got '%s'", widgetName, status.WidgetName)
		}

		if status.Status != "active" {
			t.Errorf("Expected status 'active', got '%s'", status.Status)
		}

		if status.Output != "test output" {
			t.Errorf("Expected output 'test output', got '%s'", status.Output)
		}

		// Test unknown widget
		unknownStatus := collector.GetWidgetStatus("unknown_widget")
		if unknownStatus.Status != "unknown" {
			t.Errorf("Expected status 'unknown' for non-existent widget, got '%s'", unknownStatus.Status)
		}
	})

	t.Run("GetAllWidgetMetrics", func(t *testing.T) {
		collector := NewMetricsCollector()
		
		// Record metrics for multiple widgets
		widgets := []string{"widget1", "widget2", "widget3"}
		for _, widget := range widgets {
			collector.RecordWidgetExecution(widget, 100*time.Millisecond, true, "output", "")
		}

		allMetrics := collector.GetAllWidgetMetrics()
		if len(allMetrics) != len(widgets) {
			t.Errorf("Expected %d widget metrics, got %d", len(widgets), len(allMetrics))
		}

		for _, widget := range widgets {
			if _, exists := allMetrics[widget]; !exists {
				t.Errorf("Expected metrics for widget '%s'", widget)
			}
		}
	})
}

func TestLogging(t *testing.T) {
	collector := NewMetricsCollector()

	t.Run("AddLogEntry", func(t *testing.T) {
		level := "info"
		message := "Test log message"
		component := "test_component"
		details := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		}

		collector.AddLogEntry(level, message, component, details)

		logs, err := collector.GetLogs(10, "", "")
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		if len(logs) == 0 {
			t.Fatal("Expected at least one log entry")
		}

		lastLog := logs[len(logs)-1]
		if lastLog.Level != level {
			t.Errorf("Expected log level '%s', got '%s'", level, lastLog.Level)
		}

		if lastLog.Message != message {
			t.Errorf("Expected log message '%s', got '%s'", message, lastLog.Message)
		}

		if lastLog.Component != component {
			t.Errorf("Expected log component '%s', got '%s'", component, lastLog.Component)
		}
	})

	t.Run("LogFiltering", func(t *testing.T) {
		collector := NewMetricsCollector()

		// Add logs with different levels
		collector.AddLogEntry("info", "Info message", "test", nil)
		collector.AddLogEntry("warning", "Warning message", "test", nil)
		collector.AddLogEntry("error", "Error message", "test", nil)

		// Filter by level
		infoLogs, err := collector.GetLogs(10, "info", "")
		if err != nil {
			t.Fatalf("Failed to get info logs: %v", err)
		}

		errorLogs, err := collector.GetLogs(10, "error", "")
		if err != nil {
			t.Fatalf("Failed to get error logs: %v", err)
		}

		// Verify filtering
		for _, log := range infoLogs {
			if log.Level != "info" {
				t.Errorf("Expected info level, got '%s'", log.Level)
			}
		}

		for _, log := range errorLogs {
			if log.Level != "error" {
				t.Errorf("Expected error level, got '%s'", log.Level)
			}
		}
	})

	t.Run("LogLimit", func(t *testing.T) {
		collector := NewMetricsCollector()

		// Add more logs than the limit
		for i := 0; i < 5; i++ {
			collector.AddLogEntry("info", "Test message", "test", nil)
		}

		// Request limited number of logs
		logs, err := collector.GetLogs(3, "", "")
		if err != nil {
			t.Fatalf("Failed to get limited logs: %v", err)
		}

		if len(logs) != 3 {
			t.Errorf("Expected 3 logs, got %d", len(logs))
		}
	})

	t.Run("LogRotation", func(t *testing.T) {
		collector := NewMetricsCollector()
		collector.maxLogEntries = 3 // Set low limit for testing

		// Add more logs than the limit
		for i := 0; i < 5; i++ {
			collector.AddLogEntry("info", "Test message", "test", map[string]interface{}{
				"index": i,
			})
		}

		logs, err := collector.GetLogs(10, "", "")
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		// Should only keep the most recent entries
		if len(logs) != 3 {
			t.Errorf("Expected 3 logs after rotation, got %d", len(logs))
		}

		// Verify we kept the most recent logs
		if logs[0].Details["index"].(int) != 2 {
			t.Errorf("Expected first log to have index 2, got %v", logs[0].Details["index"])
		}
	})
}

func TestPerformanceMetrics(t *testing.T) {
	collector := NewMetricsCollector()

	t.Run("SystemStatus", func(t *testing.T) {
		// Add some widget metrics
		collector.RecordWidgetExecution("widget1", 100*time.Millisecond, true, "output", "")
		collector.RecordWidgetExecution("widget2", 200*time.Millisecond, true, "output", "")

		status := collector.GetSystemStatus()
		if status.ActiveWidgets != 2 {
			t.Errorf("Expected 2 active widgets, got %d", status.ActiveWidgets)
		}

		if status.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%s'", status.Version)
		}

		if status.Uptime == 0 {
			t.Error("Expected non-zero uptime")
		}
	})

	t.Run("MetricsSummary", func(t *testing.T) {
		collector := NewMetricsCollector()

		// Add successful executions
		collector.RecordWidgetExecution("widget1", 100*time.Millisecond, true, "output", "")
		collector.RecordWidgetExecution("widget1", 200*time.Millisecond, true, "output", "")
		
		// Add failed execution
		collector.RecordWidgetExecution("widget2", 150*time.Millisecond, false, "", "error")

		summary := collector.GetMetricsSummary()

		if summary["total_widgets"].(int) != 2 {
			t.Errorf("Expected 2 total widgets, got %v", summary["total_widgets"])
		}

		if summary["active_widgets"].(int) != 1 {
			t.Errorf("Expected 1 active widget, got %v", summary["active_widgets"])
		}

		if summary["error_widgets"].(int) != 1 {
			t.Errorf("Expected 1 error widget, got %v", summary["error_widgets"])
		}

		if summary["total_executions"].(int64) != 3 {
			t.Errorf("Expected 3 total executions, got %v", summary["total_executions"])
		}

		if summary["total_errors"].(int64) != 1 {
			t.Errorf("Expected 1 total error, got %v", summary["total_errors"])
		}

		// Check success rate (2 successful out of 3 total = 66.67%)
		successRate := summary["success_rate"].(float64)
		if successRate < 66.0 || successRate > 67.0 {
			t.Errorf("Expected success rate around 66.67%%, got %v", successRate)
		}
	})

	t.Run("HealthStatus", func(t *testing.T) {
		collector := NewMetricsCollector()

		// Test healthy status
		for i := 0; i < 10; i++ {
			collector.RecordWidgetExecution("widget1", 100*time.Millisecond, true, "output", "")
		}

		health := collector.GetHealthStatus()
		if health["status"].(string) != "healthy" {
			t.Errorf("Expected healthy status, got %v", health["status"])
		}

		// Test degraded status with low success rate
		collector = NewMetricsCollector()
		for i := 0; i < 10; i++ {
			success := i < 7 // 70% success rate
			collector.RecordWidgetExecution("widget1", 100*time.Millisecond, success, "output", "error")
		}

		health = collector.GetHealthStatus()
		if health["status"].(string) != "degraded" {
			t.Errorf("Expected degraded status, got %v", health["status"])
		}
	})
}

func TestConcurrency(t *testing.T) {
	collector := NewMetricsCollector()

	t.Run("ConcurrentWidgetMetrics", func(t *testing.T) {
		numGoroutines := 10
		numExecutions := 50

		done := make(chan bool, numGoroutines)

		// Start multiple goroutines recording metrics
		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				for j := 0; j < numExecutions; j++ {
					widgetName := "concurrent_widget"
					duration := time.Duration(j) * time.Millisecond
					success := j%2 == 0
					collector.RecordWidgetExecution(widgetName, duration, success, "output", "error")
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Verify metrics
		metrics := collector.GetWidgetMetrics("concurrent_widget")
		if metrics == nil {
			t.Fatal("Expected widget metrics to exist")
		}

		expectedExecutions := int64(numGoroutines * numExecutions)
		if metrics.ExecutionCount != expectedExecutions {
			t.Errorf("Expected %d executions, got %d", expectedExecutions, metrics.ExecutionCount)
		}
	})

	t.Run("ConcurrentLogging", func(t *testing.T) {
		numGoroutines := 10
		numLogs := 20

		done := make(chan bool, numGoroutines)

		// Start multiple goroutines adding logs
		for i := 0; i < numGoroutines; i++ {
			go func(goroutineID int) {
				for j := 0; j < numLogs; j++ {
					collector.AddLogEntry("info", "Concurrent log", "test", map[string]interface{}{
						"goroutine": goroutineID,
						"log":       j,
					})
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Verify logs were added (may be rotated due to limit)
		logs, err := collector.GetLogs(1000, "", "")
		if err != nil {
			t.Fatalf("Failed to get logs: %v", err)
		}

		// Should have some logs (may not be all due to rotation)
		if len(logs) == 0 {
			t.Error("Expected some log entries")
		}
	})
}