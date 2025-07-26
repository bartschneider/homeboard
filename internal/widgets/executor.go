package widgets

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/bartosz/homeboard/internal/config"
)

// ExecutorResult represents the result of widget execution
type ExecutorResult struct {
	Name    string
	HTML    string
	Error   error
	Elapsed time.Duration
}

// Executor handles concurrent widget execution
type Executor struct {
	pythonPath string
	timeout    time.Duration
}

// NewExecutor creates a new widget executor
func NewExecutor(pythonPath string, defaultTimeout time.Duration) *Executor {
	if pythonPath == "" {
		pythonPath = "python3"
	}
	if defaultTimeout == 0 {
		defaultTimeout = 30 * time.Second
	}

	return &Executor{
		pythonPath: pythonPath,
		timeout:    defaultTimeout,
	}
}

// ExecuteAll executes all enabled widgets concurrently
func (e *Executor) ExecuteAll(widgets []config.Widget) []ExecutorResult {
	if len(widgets) == 0 {
		return []ExecutorResult{}
	}

	results := make([]ExecutorResult, len(widgets))
	var wg sync.WaitGroup

	for i, widget := range widgets {
		wg.Add(1)
		go func(index int, w config.Widget) {
			defer wg.Done()
			results[index] = e.executeWidget(w)
		}(i, widget)
	}

	wg.Wait()
	return results
}

// executeWidget executes a single widget and returns the result
func (e *Executor) executeWidget(widget config.Widget) ExecutorResult {
	start := time.Now()
	result := ExecutorResult{
		Name: widget.Name,
	}

	// Determine timeout for this widget
	timeout := e.timeout
	if widget.Timeout > 0 {
		timeout = time.Duration(widget.Timeout) * time.Second
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Serialize parameters to JSON
	parametersJSON, err := json.Marshal(widget.Parameters)
	if err != nil {
		result.Error = fmt.Errorf("failed to serialize parameters: %w", err)
		result.HTML = e.generateErrorHTML(widget.Name, result.Error)
		result.Elapsed = time.Since(start)
		return result
	}

	// Prepare command
	cmd := exec.CommandContext(ctx, e.pythonPath, widget.Script, string(parametersJSON))

	// Execute command and capture output
	output, err := cmd.Output()
	result.Elapsed = time.Since(start)

	if err != nil {
		// Handle different types of errors
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = fmt.Errorf("widget execution timed out after %v", timeout)
		} else {
			result.Error = fmt.Errorf("widget execution failed: %w", err)
		}
		result.HTML = e.generateErrorHTML(widget.Name, result.Error)
		return result
	}

	// Clean up output and set HTML
	result.HTML = strings.TrimSpace(string(output))
	if result.HTML == "" {
		result.Error = fmt.Errorf("widget produced no output")
		result.HTML = e.generateErrorHTML(widget.Name, result.Error)
	}

	return result
}

// generateErrorHTML creates a user-friendly error display for widgets
func (e *Executor) generateErrorHTML(widgetName string, err error) string {
	return fmt.Sprintf(`
		<div class="widget-error">
			<h3>%s</h3>
			<p class="error-message">⚠️ Error: %s</p>
			<p class="error-hint">Check configuration and try again</p>
		</div>
	`, widgetName, err.Error())
}

// ValidateWidget checks if a widget script exists and is executable
func (e *Executor) ValidateWidget(widget config.Widget) error {
	// Check if script file exists
	cmd := exec.Command(e.pythonPath, "-c", fmt.Sprintf("import os; print(os.path.exists('%s'))", widget.Script))
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check script existence: %w", err)
	}

	if strings.TrimSpace(string(output)) != "True" {
		return fmt.Errorf("script file not found: %s", widget.Script)
	}

	return nil
}

// GetExecutionStats returns execution statistics for debugging
func (e *Executor) GetExecutionStats(results []ExecutorResult) map[string]interface{} {
	stats := map[string]interface{}{
		"total_widgets": len(results),
		"successful":    0,
		"failed":        0,
		"total_time":    time.Duration(0),
		"max_time":      time.Duration(0),
		"min_time":      time.Duration(0),
	}

	if len(results) == 0 {
		return stats
	}

	var totalTime time.Duration
	minTime := results[0].Elapsed
	maxTime := results[0].Elapsed

	for _, result := range results {
		totalTime += result.Elapsed
		if result.Elapsed > maxTime {
			maxTime = result.Elapsed
		}
		if result.Elapsed < minTime {
			minTime = result.Elapsed
		}

		if result.Error == nil {
			stats["successful"] = stats["successful"].(int) + 1
		} else {
			stats["failed"] = stats["failed"].(int) + 1
		}
	}

	stats["total_time"] = totalTime
	stats["max_time"] = maxTime
	stats["min_time"] = minTime
	stats["avg_time"] = totalTime / time.Duration(len(results))

	return stats
}