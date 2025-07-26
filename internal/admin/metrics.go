package admin

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

// MetricsCollector collects and manages system and widget metrics
type MetricsCollector struct {
	systemMetrics SystemMetrics
	widgetMetrics map[string]*WidgetMetrics
	logEntries    []LogEntry
	startTime     time.Time
	requestCount  int64
	errorCount    int64
	mutex         sync.RWMutex
	logMutex      sync.RWMutex
	maxLogEntries int
}

// WidgetMetrics represents metrics for a single widget
type WidgetMetrics struct {
	WidgetName      string        `json:"widget_name"`
	Status          string        `json:"status"`
	LastExecution   time.Time     `json:"last_execution"`
	ExecutionTime   time.Duration `json:"execution_time"`
	ExecutionCount  int64         `json:"execution_count"`
	SuccessCount    int64         `json:"success_count"`
	ErrorCount      int64         `json:"error_count"`
	ErrorMessage    string        `json:"error_message,omitempty"`
	AverageTime     time.Duration `json:"average_time"`
	MinTime         time.Duration `json:"min_time"`
	MaxTime         time.Duration `json:"max_time"`
	LastOutput      string        `json:"last_output,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		systemMetrics: SystemMetrics{},
		widgetMetrics: make(map[string]*WidgetMetrics),
		logEntries:    make([]LogEntry, 0),
		startTime:     time.Now(),
		maxLogEntries: 1000,
	}
}

// System Metrics Methods

// UpdateSystemMetrics updates system performance metrics
func (mc *MetricsCollector) UpdateSystemMetrics() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// Get memory statistics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Calculate memory usage as percentage (approximate)
	memoryUsagePercent := float64(m.Sys) / (1024 * 1024 * 1024) * 100 // Rough calculation

	// Update metrics
	mc.systemMetrics.MemoryUsage = memoryUsagePercent
	mc.systemMetrics.RequestCount = mc.requestCount
	mc.systemMetrics.ErrorCount = mc.errorCount
	mc.systemMetrics.TotalUptime = int64(time.Since(mc.startTime).Seconds())

	// Calculate average latency (simplified)
	if mc.requestCount > 0 {
		mc.systemMetrics.AverageLatency = float64(mc.systemMetrics.TotalUptime) / float64(mc.requestCount)
	}
}

// GetSystemStatus returns current system status
func (mc *MetricsCollector) GetSystemStatus() SystemStatus {
	mc.UpdateSystemMetrics()

	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	// Count active widgets
	activeWidgets := 0
	for _, metrics := range mc.widgetMetrics {
		if metrics.Status == "active" {
			activeWidgets++
		}
	}

	status := "running"
	if mc.errorCount > mc.requestCount/2 {
		status = "error"
	}

	return SystemStatus{
		Status:        status,
		Uptime:        time.Since(mc.startTime),
		Version:       "1.0.0", // Should be injected during build
		Metrics:       mc.systemMetrics,
		ActiveWidgets: activeWidgets,
		LastRefresh:   time.Now(),
		ServerPort:    8081, // Should be configurable
	}
}

// GetAllMetrics returns all collected metrics
func (mc *MetricsCollector) GetAllMetrics() map[string]interface{} {
	mc.UpdateSystemMetrics()

	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	return map[string]interface{}{
		"system_metrics": mc.systemMetrics,
		"widget_metrics": mc.widgetMetrics,
		"uptime":         time.Since(mc.startTime),
		"start_time":     mc.startTime,
		"collected_at":   time.Now(),
	}
}

// Widget Metrics Methods

// RecordWidgetExecution records metrics for a widget execution
func (mc *MetricsCollector) RecordWidgetExecution(widgetName string, duration time.Duration, success bool, output string, errorMsg string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// Get or create widget metrics
	metrics, exists := mc.widgetMetrics[widgetName]
	if !exists {
		metrics = &WidgetMetrics{
			WidgetName: widgetName,
			Status:     "idle",
			CreatedAt:  time.Now(),
			MinTime:    duration,
			MaxTime:    duration,
		}
		mc.widgetMetrics[widgetName] = metrics
	}

	// Update metrics
	metrics.LastExecution = time.Now()
	metrics.ExecutionTime = duration
	metrics.ExecutionCount++
	metrics.UpdatedAt = time.Now()

	if success {
		metrics.SuccessCount++
		metrics.Status = "active"
		metrics.ErrorMessage = ""
		metrics.LastOutput = output
	} else {
		metrics.ErrorCount++
		metrics.Status = "error"
		metrics.ErrorMessage = errorMsg
	}

	// Update timing statistics
	if duration < metrics.MinTime || metrics.MinTime == 0 {
		metrics.MinTime = duration
	}
	if duration > metrics.MaxTime {
		metrics.MaxTime = duration
	}

	// Calculate average time
	if metrics.ExecutionCount > 0 {
		totalTime := time.Duration(int64(metrics.AverageTime) * (metrics.ExecutionCount - 1))
		metrics.AverageTime = (totalTime + duration) / time.Duration(metrics.ExecutionCount)
	}
}

// GetWidgetStatus returns status for a specific widget
func (mc *MetricsCollector) GetWidgetStatus(widgetName string) WidgetStatusUpdate {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	metrics, exists := mc.widgetMetrics[widgetName]
	if !exists {
		return WidgetStatusUpdate{
			WidgetName: widgetName,
			Status:     "unknown",
		}
	}

	return WidgetStatusUpdate{
		WidgetID:      widgetName, // Using name as ID for simplicity
		WidgetName:    widgetName,
		Status:        metrics.Status,
		LastExecution: metrics.LastExecution,
		ExecutionTime: metrics.ExecutionTime,
		Error:         metrics.ErrorMessage,
		Output:        metrics.LastOutput,
	}
}

// GetWidgetMetrics returns detailed metrics for a specific widget
func (mc *MetricsCollector) GetWidgetMetrics(widgetName string) *WidgetMetrics {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	metrics, exists := mc.widgetMetrics[widgetName]
	if !exists {
		return nil
	}

	// Return a copy to avoid race conditions
	metricsCopy := *metrics
	return &metricsCopy
}

// GetAllWidgetMetrics returns metrics for all widgets
func (mc *MetricsCollector) GetAllWidgetMetrics() map[string]*WidgetMetrics {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	// Return copies to avoid race conditions
	result := make(map[string]*WidgetMetrics)
	for name, metrics := range mc.widgetMetrics {
		metricsCopy := *metrics
		result[name] = &metricsCopy
	}

	return result
}

// ResetWidgetMetrics resets metrics for a specific widget
func (mc *MetricsCollector) ResetWidgetMetrics(widgetName string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if metrics, exists := mc.widgetMetrics[widgetName]; exists {
		metrics.ExecutionCount = 0
		metrics.SuccessCount = 0
		metrics.ErrorCount = 0
		metrics.AverageTime = 0
		metrics.MinTime = 0
		metrics.MaxTime = 0
		metrics.LastOutput = ""
		metrics.ErrorMessage = ""
		metrics.Status = "idle"
		metrics.UpdatedAt = time.Now()
	}
}

// Request/Response Metrics

// IncrementRequestCount increments the total request count
func (mc *MetricsCollector) IncrementRequestCount() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.requestCount++
}

// IncrementErrorCount increments the total error count
func (mc *MetricsCollector) IncrementErrorCount() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.errorCount++
}

// Logging Methods

// AddLogEntry adds a new log entry
func (mc *MetricsCollector) AddLogEntry(level, message, component string, details map[string]interface{}) {
	mc.logMutex.Lock()
	defer mc.logMutex.Unlock()

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Component: component,
		Details:   details,
	}

	mc.logEntries = append(mc.logEntries, entry)

	// Keep only the most recent entries
	if len(mc.logEntries) > mc.maxLogEntries {
		mc.logEntries = mc.logEntries[len(mc.logEntries)-mc.maxLogEntries:]
	}
}

// GetLogs returns recent log entries with optional filtering
func (mc *MetricsCollector) GetLogs(limit int, level string, since string) ([]LogEntry, error) {
	mc.logMutex.RLock()
	defer mc.logMutex.RUnlock()

	// Parse since parameter if provided
	var sinceTime time.Time
	if since != "" {
		var err error
		sinceTime, err = time.Parse(time.RFC3339, since)
		if err != nil {
			return nil, fmt.Errorf("invalid since parameter: %w", err)
		}
	}

	// Filter logs
	var filtered []LogEntry
	for _, entry := range mc.logEntries {
		// Filter by level
		if level != "" && entry.Level != level {
			continue
		}

		// Filter by time
		if !sinceTime.IsZero() && entry.Timestamp.Before(sinceTime) {
			continue
		}

		filtered = append(filtered, entry)
	}

	// Limit results
	if limit > 0 && limit < len(filtered) {
		filtered = filtered[len(filtered)-limit:]
	}

	return filtered, nil
}

// Performance Monitoring

// StartPerformanceMonitoring starts background performance monitoring
func (mc *MetricsCollector) StartPerformanceMonitoring() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			mc.UpdateSystemMetrics()
			mc.collectPerformanceMetrics()
		}
	}()
}

// collectPerformanceMetrics collects detailed performance metrics
func (mc *MetricsCollector) collectPerformanceMetrics() {
	// Collect CPU usage (simplified - would need a proper system monitoring library)
	mc.mutex.Lock()
	mc.systemMetrics.CPUUsage = mc.estimateCPUUsage()
	mc.mutex.Unlock()

	// Log performance metrics
	mc.AddLogEntry("info", "Performance metrics updated", "metrics_collector", map[string]interface{}{
		"cpu_usage":    mc.systemMetrics.CPUUsage,
		"memory_usage": mc.systemMetrics.MemoryUsage,
		"uptime":       time.Since(mc.startTime).Seconds(),
	})
}

// estimateCPUUsage provides a simple CPU usage estimation
func (mc *MetricsCollector) estimateCPUUsage() float64 {
	// This is a placeholder - in a real implementation, you'd use
	// system monitoring libraries like gopsutil
	return float64(runtime.NumGoroutine()) / 100.0 * 10.0 // Rough estimation
}

// Export/Import Methods

// ExportMetrics exports all metrics to a JSON file
func (mc *MetricsCollector) ExportMetrics(filePath string) error {
	mc.mutex.RLock()
	mc.logMutex.RLock()
	defer mc.mutex.RUnlock()
	defer mc.logMutex.RUnlock()

	data := map[string]interface{}{
		"system_metrics": mc.systemMetrics,
		"widget_metrics": mc.widgetMetrics,
		"log_entries":    mc.logEntries,
		"start_time":     mc.startTime,
		"exported_at":    time.Now(),
		"version":        "1.0",
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write metrics file: %w", err)
	}

	return nil
}

// GetMetricsSummary returns a summary of key metrics
func (mc *MetricsCollector) GetMetricsSummary() map[string]interface{} {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	totalWidgets := len(mc.widgetMetrics)
	activeWidgets := 0
	errorWidgets := 0
	totalExecutions := int64(0)
	totalErrors := int64(0)

	for _, metrics := range mc.widgetMetrics {
		totalExecutions += metrics.ExecutionCount
		totalErrors += metrics.ErrorCount

		switch metrics.Status {
		case "active":
			activeWidgets++
		case "error":
			errorWidgets++
		}
	}

	successRate := float64(100)
	if totalExecutions > 0 {
		successRate = float64(totalExecutions-totalErrors) / float64(totalExecutions) * 100
	}

	return map[string]interface{}{
		"uptime":            time.Since(mc.startTime),
		"total_widgets":     totalWidgets,
		"active_widgets":    activeWidgets,
		"error_widgets":     errorWidgets,
		"total_executions":  totalExecutions,
		"total_errors":      totalErrors,
		"success_rate":      successRate,
		"requests_count":    mc.requestCount,
		"memory_usage":      mc.systemMetrics.MemoryUsage,
		"cpu_usage":         mc.systemMetrics.CPUUsage,
	}
}

// Health Check Methods

// GetHealthStatus returns overall system health status
func (mc *MetricsCollector) GetHealthStatus() map[string]interface{} {
	summary := mc.GetMetricsSummary()
	
	// Determine health status
	health := "healthy"
	if summary["success_rate"].(float64) < 80 {
		health = "degraded"
	}
	if summary["success_rate"].(float64) < 50 {
		health = "unhealthy"
	}

	// Check for recent errors
	recentErrors := 0
	cutoffTime := time.Now().Add(-10 * time.Minute)
	
	mc.logMutex.RLock()
	for _, entry := range mc.logEntries {
		if entry.Level == "error" && entry.Timestamp.After(cutoffTime) {
			recentErrors++
		}
	}
	mc.logMutex.RUnlock()

	if recentErrors > 5 {
		health = "degraded"
	}
	if recentErrors > 20 {
		health = "unhealthy"
	}

	return map[string]interface{}{
		"status":        health,
		"uptime":        summary["uptime"],
		"success_rate":  summary["success_rate"],
		"recent_errors": recentErrors,
		"active_widgets": summary["active_widgets"],
		"memory_usage":  summary["memory_usage"],
		"timestamp":     time.Now(),
	}
}