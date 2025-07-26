package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bartosz/homeboard/internal/config"
	"github.com/bartosz/homeboard/internal/widgets"
)

// WidgetTestRunner handles widget testing in isolation
type WidgetTestRunner struct {
	executor    *widgets.Executor
	pythonPath  string
	defaultTimeout time.Duration
	environment map[string]string
}

// NewWidgetTestRunner creates a new widget test runner
func NewWidgetTestRunner(executor *widgets.Executor) *WidgetTestRunner {
	return &WidgetTestRunner{
		executor:       executor,
		pythonPath:     "python3",
		defaultTimeout: 30 * time.Second,
		environment: map[string]string{
			"WIDGET_TEST_MODE": "true",
			"PYTHONPATH":       ".",
		},
	}
}

// TestWidget executes a widget test with given parameters
func (tr *WidgetTestRunner) TestWidget(widget config.Widget, params map[string]interface{}, timeout int) TestResult {
	start := time.Now()
	
	result := TestResult{
		Parameters: params,
		Timestamp:  start,
		Environment: tr.environment,
	}

	// Use provided timeout or default
	testTimeout := tr.defaultTimeout
	if timeout > 0 {
		testTimeout = time.Duration(timeout) * time.Second
	}

	// Validate widget first
	if err := tr.validateWidget(widget); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Widget validation failed: %v", err)
		result.ValidationErrors = []string{err.Error()}
		result.ExecutionTime = time.Since(start)
		return result
	}

	// Prepare parameters JSON
	parametersJSON, err := json.Marshal(params)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to serialize parameters: %v", err)
		result.ExecutionTime = time.Since(start)
		return result
	}

	// Execute widget
	output, exitCode, err := tr.executeWidget(widget, string(parametersJSON), testTimeout)
	result.ExecutionTime = time.Since(start)
	result.ExitCode = exitCode

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Output = output // May contain partial output
		return result
	}

	// Validate output
	if validationErrors := tr.validateOutput(output); len(validationErrors) > 0 {
		result.Success = false
		result.ValidationErrors = validationErrors
		result.Output = output
		return result
	}

	// Test successful
	result.Success = true
	result.Output = output
	return result
}

// TestWidgetWithVariations tests a widget with multiple parameter sets
func (tr *WidgetTestRunner) TestWidgetWithVariations(widget config.Widget, variations []map[string]interface{}) []TestResult {
	results := make([]TestResult, len(variations))
	
	for i, params := range variations {
		results[i] = tr.TestWidget(widget, params, 0) // Use default timeout
	}
	
	return results
}

// BenchmarkWidget runs performance benchmarks on a widget
func (tr *WidgetTestRunner) BenchmarkWidget(widget config.Widget, params map[string]interface{}, iterations int) BenchmarkResult {
	if iterations <= 0 {
		iterations = 10
	}

	results := make([]TestResult, iterations)
	var totalDuration time.Duration
	successCount := 0

	for i := 0; i < iterations; i++ {
		result := tr.TestWidget(widget, params, 0)
		results[i] = result
		totalDuration += result.ExecutionTime
		
		if result.Success {
			successCount++
		}
	}

	// Calculate statistics
	benchmark := BenchmarkResult{
		Widget:         widget,
		Parameters:     params,
		Iterations:     iterations,
		SuccessCount:   successCount,
		SuccessRate:    float64(successCount) / float64(iterations),
		TotalDuration:  totalDuration,
		AverageDuration: totalDuration / time.Duration(iterations),
		Results:        results,
	}

	// Calculate min/max execution times
	if len(results) > 0 {
		benchmark.MinDuration = results[0].ExecutionTime
		benchmark.MaxDuration = results[0].ExecutionTime
		
		for _, result := range results {
			if result.ExecutionTime < benchmark.MinDuration {
				benchmark.MinDuration = result.ExecutionTime
			}
			if result.ExecutionTime > benchmark.MaxDuration {
				benchmark.MaxDuration = result.ExecutionTime
			}
		}
	}

	return benchmark
}

// validateWidget performs pre-execution validation
func (tr *WidgetTestRunner) validateWidget(widget config.Widget) error {
	// Check if script file exists
	if _, err := os.Stat(widget.Script); os.IsNotExist(err) {
		return fmt.Errorf("widget script not found: %s", widget.Script)
	}

	// Check if script is executable
	info, err := os.Stat(widget.Script)
	if err != nil {
		return fmt.Errorf("cannot access widget script: %v", err)
	}

	mode := info.Mode()
	if mode&0111 == 0 {
		return fmt.Errorf("widget script is not executable: %s", widget.Script)
	}

	// Validate script extension
	if !strings.HasSuffix(widget.Script, ".py") {
		return fmt.Errorf("widget script must be a Python file (.py): %s", widget.Script)
	}

	return nil
}

// executeWidget runs the widget script and captures output
func (tr *WidgetTestRunner) executeWidget(widget config.Widget, parametersJSON string, timeout time.Duration) (string, int, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Prepare command
	cmd := exec.CommandContext(ctx, tr.pythonPath, widget.Script, parametersJSON)

	// Set environment
	cmd.Env = os.Environ()
	for key, value := range tr.environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Execute command
	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	// Get exit code
	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	// Handle different error types
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return outputStr, exitCode, fmt.Errorf("widget execution timed out after %v", timeout)
		}
		
		if exitCode != 0 {
			return outputStr, exitCode, fmt.Errorf("widget script exited with code %d: %s", exitCode, outputStr)
		}
		
		return outputStr, exitCode, fmt.Errorf("widget execution failed: %v", err)
	}

	return outputStr, exitCode, nil
}

// validateOutput validates the widget output format
func (tr *WidgetTestRunner) validateOutput(output string) []string {
	var errors []string

	// Check if output is empty
	if strings.TrimSpace(output) == "" {
		errors = append(errors, "Widget produced no output")
		return errors
	}

	// Check if output contains HTML
	if !tr.isValidHTML(output) {
		errors = append(errors, "Widget output does not appear to be valid HTML")
	}

	// Check for potential security issues
	if tr.containsUnsafeContent(output) {
		errors = append(errors, "Widget output contains potentially unsafe content")
	}

	// Check output length
	if len(output) > 100000 { // 100KB limit
		errors = append(errors, fmt.Sprintf("Widget output is too large (%d bytes, max 100KB)", len(output)))
	}

	return errors
}

// isValidHTML performs basic HTML validation
func (tr *WidgetTestRunner) isValidHTML(output string) bool {
	output = strings.TrimSpace(output)
	
	// Must start with < and end with >
	if !strings.HasPrefix(output, "<") || !strings.HasSuffix(output, ">") {
		return false
	}

	// Check for balanced tags (basic validation)
	openTags := strings.Count(output, "<")
	closeTags := strings.Count(output, "</")
	selfClosing := strings.Count(output, "/>")

	// Self-closing tags count as both open and close
	expectedCloseTags := openTags - selfClosing
	actualCloseTags := closeTags + selfClosing

	return actualCloseTags >= expectedCloseTags/2 // Allow for some flexibility
}

// containsUnsafeContent checks for potentially dangerous content
func (tr *WidgetTestRunner) containsUnsafeContent(output string) bool {
	lowerOutput := strings.ToLower(output)
	
	// Check for script tags
	if strings.Contains(lowerOutput, "<script") {
		return true
	}
	
	// Check for dangerous attributes
	dangerousAttrs := []string{
		"javascript:",
		"onclick=",
		"onload=",
		"onerror=",
		"onmouseover=",
	}
	
	for _, attr := range dangerousAttrs {
		if strings.Contains(lowerOutput, attr) {
			return true
		}
	}
	
	return false
}

// BenchmarkResult represents the result of widget benchmarking
type BenchmarkResult struct {
	Widget          config.Widget     `json:"widget"`
	Parameters      map[string]interface{} `json:"parameters"`
	Iterations      int               `json:"iterations"`
	SuccessCount    int               `json:"success_count"`
	SuccessRate     float64           `json:"success_rate"`
	TotalDuration   time.Duration     `json:"total_duration"`
	AverageDuration time.Duration     `json:"average_duration"`
	MinDuration     time.Duration     `json:"min_duration"`
	MaxDuration     time.Duration     `json:"max_duration"`
	Results         []TestResult      `json:"results"`
}

// TestSuite represents a collection of widget tests
type TestSuite struct {
	Name        string                    `json:"name"`
	Description string                    `json:"description"`
	Tests       []WidgetTestCase         `json:"tests"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}

// WidgetTestCase represents a single test case
type WidgetTestCase struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Widget       config.Widget          `json:"widget"`
	Parameters   map[string]interface{} `json:"parameters"`
	Expected     ExpectedResult         `json:"expected"`
	Timeout      int                    `json:"timeout"`
}

// ExpectedResult represents expected test outcomes
type ExpectedResult struct {
	Success      bool     `json:"success"`
	ContainsText []string `json:"contains_text,omitempty"`
	NotContains  []string `json:"not_contains,omitempty"`
	MaxDuration  int      `json:"max_duration,omitempty"` // in milliseconds
	MinLength    int      `json:"min_length,omitempty"`
	MaxLength    int      `json:"max_length,omitempty"`
}

// RunTestSuite executes a complete test suite
func (tr *WidgetTestRunner) RunTestSuite(suite TestSuite) TestSuiteResult {
	start := time.Now()
	
	results := make([]TestCaseResult, len(suite.Tests))
	passCount := 0
	
	for i, testCase := range suite.Tests {
		testResult := tr.TestWidget(testCase.Widget, testCase.Parameters, testCase.Timeout)
		
		// Validate against expected results
		passed := tr.validateExpectedResult(testResult, testCase.Expected)
		
		results[i] = TestCaseResult{
			TestCase: testCase,
			Result:   testResult,
			Passed:   passed,
		}
		
		if passed {
			passCount++
		}
	}
	
	return TestSuiteResult{
		Suite:        suite,
		Results:      results,
		PassCount:    passCount,
		TotalCount:   len(suite.Tests),
		PassRate:     float64(passCount) / float64(len(suite.Tests)),
		Duration:     time.Since(start),
		ExecutedAt:   start,
	}
}

// validateExpectedResult checks if test result matches expectations
func (tr *WidgetTestRunner) validateExpectedResult(result TestResult, expected ExpectedResult) bool {
	// Check success expectation
	if result.Success != expected.Success {
		return false
	}
	
	// Check content expectations (only if test was successful)
	if result.Success {
		// Check required text
		for _, text := range expected.ContainsText {
			if !strings.Contains(result.Output, text) {
				return false
			}
		}
		
		// Check forbidden text
		for _, text := range expected.NotContains {
			if strings.Contains(result.Output, text) {
				return false
			}
		}
		
		// Check length constraints
		outputLen := len(result.Output)
		if expected.MinLength > 0 && outputLen < expected.MinLength {
			return false
		}
		if expected.MaxLength > 0 && outputLen > expected.MaxLength {
			return false
		}
	}
	
	// Check duration constraint
	if expected.MaxDuration > 0 {
		maxDuration := time.Duration(expected.MaxDuration) * time.Millisecond
		if result.ExecutionTime > maxDuration {
			return false
		}
	}
	
	return true
}

// TestCaseResult represents the result of a single test case
type TestCaseResult struct {
	TestCase WidgetTestCase `json:"test_case"`
	Result   TestResult     `json:"result"`
	Passed   bool           `json:"passed"`
}

// TestSuiteResult represents the result of running a test suite
type TestSuiteResult struct {
	Suite      TestSuite        `json:"suite"`
	Results    []TestCaseResult `json:"results"`
	PassCount  int              `json:"pass_count"`
	TotalCount int              `json:"total_count"`
	PassRate   float64          `json:"pass_rate"`
	Duration   time.Duration    `json:"duration"`
	ExecutedAt time.Time        `json:"executed_at"`
}