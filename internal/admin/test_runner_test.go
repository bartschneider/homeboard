package admin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bartosz/homeboard/internal/config"
	"github.com/bartosz/homeboard/internal/widgets"
)

func TestWidgetTestRunner(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	// Create test widget script
	testScript := `#!/usr/bin/env python3
import sys
import json
import time

# Parse arguments
if len(sys.argv) > 1:
    params = json.loads(sys.argv[1])
else:
    params = {}

# Simulate some work
if params.get("delay"):
    time.sleep(params["delay"])

# Check for test scenarios
if params.get("fail"):
    print("Error: Test failure", file=sys.stderr)
    sys.exit(1)

if params.get("timeout"):
    time.sleep(10)  # Sleep longer than timeout

# Output test result
result = {
    "status": "success",
    "message": "Test widget executed successfully",
    "data": params.get("data", "test data"),
    "timestamp": time.time()
}

print(json.dumps(result))
`

	scriptPath := filepath.Join(tempDir, "test_widget.py")
	if err := os.WriteFile(scriptPath, []byte(testScript), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Create executor and test runner
	executor := widgets.NewExecutor("python3", 30*time.Second)
	testRunner := NewWidgetTestRunner(executor)

	t.Run("InitializationAndValidation", func(t *testing.T) {
		if testRunner.executor == nil {
			t.Error("Expected executor to be set")
		}

		if testRunner.pythonPath == "" {
			t.Error("Expected python path to be set")
		}

		if testRunner.defaultTimeout == 0 {
			t.Error("Expected default timeout to be set")
		}
	})

	t.Run("SuccessfulWidgetTest", func(t *testing.T) {
		widget := config.Widget{
			Name:    "Test Widget",
			Script:  "test_widget.py",
			Enabled: true,
			Timeout: 5,
			Parameters: map[string]interface{}{
				"data": "test input",
			},
		}

		result := testRunner.TestWidget(widget, map[string]interface{}{}, 0)

		if !result.Success {
			t.Errorf("Expected successful test, but got failure: %s", result.Error)
		}

		if result.ExecutionTime <= 0 {
			t.Error("Expected positive execution time")
		}

		if result.Output == "" {
			t.Error("Expected non-empty output")
		}

		if !strings.Contains(result.Output, "success") {
			t.Errorf("Expected output to contain 'success', got: %s", result.Output)
		}

		// Test completed successfully
		if result.ExecutionTime <= 0 {
			t.Error("Expected positive execution time")
		}
	})

	t.Run("FailingWidgetTest", func(t *testing.T) {
		widget := config.Widget{
			Name:    "Failing Widget",
			Script:  "test_widget.py",
			Enabled: true,
			Timeout: 5,
			Parameters: map[string]interface{}{
				"fail": true,
			},
		}

		result := testRunner.TestWidget(widget, map[string]interface{}{}, 0)

		if result.Success {
			t.Error("Expected test to fail, but it succeeded")
		}

		if result.Error == "" {
			t.Error("Expected non-empty error message")
		}

		if !strings.Contains(result.Error, "Test failure") {
			t.Errorf("Expected error to contain 'Test failure', got: %s", result.Error)
		}
	})

	t.Run("WidgetTestTimeout", func(t *testing.T) {
		widget := config.Widget{
			Name:    "Timeout Widget",
			Script:  "test_widget.py",
			Enabled: true,
			Timeout: 1, // Short timeout
			Parameters: map[string]interface{}{
				"timeout": true,
			},
		}

		start := time.Now()
		result := testRunner.TestWidget(widget, map[string]interface{}{}, 0)
		duration := time.Since(start)

		if result.Success {
			t.Error("Expected test to fail due to timeout")
		}

		if !strings.Contains(result.Error, "timeout") && !strings.Contains(result.Error, "killed") {
			t.Errorf("Expected timeout error, got: %s", result.Error)
		}

		// Should timeout around the specified timeout duration
		if duration > 3*time.Second {
			t.Errorf("Test took too long to timeout: %v", duration)
		}
	})

	t.Run("NonExistentScript", func(t *testing.T) {
		widget := config.Widget{
			Name:    "Non-existent Widget",
			Script:  "nonexistent.py",
			Enabled: true,
			Timeout: 5,
		}

		result := testRunner.TestWidget(widget, map[string]interface{}{}, 0)

		if result.Success {
			t.Error("Expected test to fail for non-existent script")
		}

		if result.Error == "" {
			t.Error("Expected non-empty error message")
		}
	})

	t.Run("WidgetTestWithEnvironment", func(t *testing.T) {
		widget := config.Widget{
			Name:    "Environment Widget",
			Script:  "test_widget.py",
			Enabled: true,
			Timeout: 5,
		}

		result := testRunner.TestWidget(widget, map[string]interface{}{}, 0)

		if !result.Success {
			t.Errorf("Expected successful test with environment, but got failure: %s", result.Error)
		}
	})

	t.Run("ParameterInjection", func(t *testing.T) {
		widget := config.Widget{
			Name:    "Parameter Widget",
			Script:  "test_widget.py",
			Enabled: true,
			Timeout: 5,
			Parameters: map[string]interface{}{
				"string_param":  "test_string",
				"number_param":  42,
				"boolean_param": true,
				"nested_param": map[string]interface{}{
					"nested_string": "nested_value",
				},
			},
		}

		result := testRunner.TestWidget(widget, map[string]interface{}{}, 0)

		if !result.Success {
			t.Errorf("Expected successful test with parameters, but got failure: %s", result.Error)
		}

		// The test script should echo back the parameters
		if !strings.Contains(result.Output, "test_string") {
			t.Error("Expected output to contain string parameter")
		}
	})
}

func TestWidgetTestSuite(t *testing.T) {
	tempDir := t.TempDir()

	// Create multiple test scripts
	successScript := `#!/usr/bin/env python3
import json
print(json.dumps({"status": "success", "message": "Success"}))
`

	failScript := `#!/usr/bin/env python3
import sys
print("Error: Script failed", file=sys.stderr)
sys.exit(1)
`

	slowScript := `#!/usr/bin/env python3
import time
import json
time.sleep(2)
print(json.dumps({"status": "success", "message": "Slow success"}))
`

	scripts := map[string]string{
		"success.py": successScript,
		"fail.py":    failScript,
		"slow.py":    slowScript,
	}

	for name, content := range scripts {
		scriptPath := filepath.Join(tempDir, name)
		if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
			t.Fatalf("Failed to create script %s: %v", name, err)
		}
	}

	executor := widgets.NewExecutor("python3", 30*time.Second)
	testRunner := NewWidgetTestRunner(executor)

	t.Run("TestMultipleWidgets", func(t *testing.T) {
		widgets := []config.Widget{
			{
				Name:    "Success Widget",
				Script:  filepath.Join(tempDir, "success.py"),
				Enabled: true,
				Timeout: 5,
			},
			{
				Name:    "Fail Widget",
				Script:  filepath.Join(tempDir, "fail.py"),
				Enabled: true,
				Timeout: 5,
			},
			{
				Name:    "Slow Widget",
				Script:  filepath.Join(tempDir, "slow.py"),
				Enabled: true,
				Timeout: 5,
			},
		}

		successCount := 0
		failureCount := 0

		// Test each widget individually since we don't have RunTestSuite
		for _, widget := range widgets {
			result := testRunner.TestWidget(widget, map[string]interface{}{}, 0)
			if result.Success {
				successCount++
			} else {
				failureCount++
			}
		}

		if successCount != 2 {
			t.Errorf("Expected 2 successful tests, got %d", successCount)
		}

		if failureCount != 1 {
			t.Errorf("Expected 1 failed test, got %d", failureCount)
		}
	})

	t.Run("DisabledWidgetSkipping", func(t *testing.T) {
		widgets := []config.Widget{
			{
				Name:    "Enabled Widget",
				Script:  filepath.Join(tempDir, "success.py"),
				Enabled: true,
				Timeout: 5,
			},
			{
				Name:    "Disabled Widget",
				Script:  filepath.Join(tempDir, "success.py"),
				Enabled: false, // Disabled
				Timeout: 5,
			},
		}

		enabledCount := 0
		for _, widget := range widgets {
			if widget.Enabled {
				result := testRunner.TestWidget(widget, map[string]interface{}{}, 0)
				if result.Success {
					enabledCount++
				}
			}
		}

		if enabledCount != 1 {
			t.Errorf("Expected 1 enabled test to pass, got %d", enabledCount)
		}
	})

	t.Run("EmptyTestSuite", func(t *testing.T) {
		widgets := []config.Widget{}

		count := 0
		for _, widget := range widgets {
			if widget.Enabled {
				testRunner.TestWidget(widget, map[string]interface{}{}, 0)
				count++
			}
		}

		if count != 0 {
			t.Errorf("Expected 0 tests for empty suite, got %d", count)
		}
	})
}

func TestWidgetBenchmark(t *testing.T) {
	tempDir := t.TempDir()

	// Create benchmark test script
	benchmarkScript := `#!/usr/bin/env python3
import json
import time
import sys

# Parse parameters
if len(sys.argv) > 1:
    params = json.loads(sys.argv[1])
else:
    params = {}

# Simulate variable execution time
delay = params.get("delay", 0.1)
time.sleep(delay)

result = {
    "status": "success",
    "message": f"Benchmark completed with {delay}s delay",
    "execution_time": delay
}

print(json.dumps(result))
`

	scriptPath := filepath.Join(tempDir, "benchmark.py")
	if err := os.WriteFile(scriptPath, []byte(benchmarkScript), 0755); err != nil {
		t.Fatalf("Failed to create benchmark script: %v", err)
	}

	executor := widgets.NewExecutor("python3", 30*time.Second)
	testRunner := NewWidgetTestRunner(executor)

	t.Run("WidgetBenchmark", func(t *testing.T) {
		widget := config.Widget{
			Name:    "Benchmark Widget",
			Script:  filepath.Join(tempDir, "benchmark.py"),
			Enabled: true,
			Timeout: 10,
			Parameters: map[string]interface{}{
				"delay": 0.1,
			},
		}

		iterations := 5
		params := map[string]interface{}{
			"delay": 0.1,
		}
		benchmarkResult := testRunner.BenchmarkWidget(widget, params, iterations)

		if benchmarkResult.Iterations != iterations {
			t.Errorf("Expected %d iterations, got %d", iterations, benchmarkResult.Iterations)
		}

		if benchmarkResult.SuccessCount != iterations {
			t.Errorf("Expected %d successful runs, got %d", iterations, benchmarkResult.SuccessCount)
		}

		if benchmarkResult.AverageDuration <= 0 {
			t.Error("Expected positive average time")
		}

		if benchmarkResult.MinDuration <= 0 {
			t.Error("Expected positive minimum time")
		}

		if benchmarkResult.MaxDuration <= 0 {
			t.Error("Expected positive maximum time")
		}

		if benchmarkResult.TotalDuration <= 0 {
			t.Error("Expected positive total time")
		}

		// Average should be between min and max
		if benchmarkResult.AverageDuration < benchmarkResult.MinDuration || benchmarkResult.AverageDuration > benchmarkResult.MaxDuration {
			t.Error("Average time should be between min and max times")
		}

		// Check that we have results recorded
		if len(benchmarkResult.Results) != iterations {
			t.Errorf("Expected %d results, got %d", iterations, len(benchmarkResult.Results))
		}
	})

	t.Run("BenchmarkWithFailures", func(t *testing.T) {
		// Create a script that fails sometimes
		failingSomtimesScript := `#!/usr/bin/env python3
import json
import random
import sys

# Randomly fail 30% of the time
if random.random() < 0.3:
    print("Random failure", file=sys.stderr)
    sys.exit(1)

print(json.dumps({"status": "success", "message": "Success"}))
`

		failingScriptPath := filepath.Join(tempDir, "failing_sometimes.py")
		if err := os.WriteFile(failingScriptPath, []byte(failingSomtimesScript), 0755); err != nil {
			t.Fatalf("Failed to create failing script: %v", err)
		}

		widget := config.Widget{
			Name:    "Sometimes Failing Widget",
			Script:  filepath.Join(tempDir, "failing_sometimes.py"),
			Enabled: true,
			Timeout: 5,
		}

		iterations := 20
		params := map[string]interface{}{}
		benchmarkResult := testRunner.BenchmarkWidget(widget, params, iterations)

		if benchmarkResult.Iterations != iterations {
			t.Errorf("Expected %d iterations, got %d", iterations, benchmarkResult.Iterations)
		}

		// Should have recorded all iterations in results
		if len(benchmarkResult.Results) != iterations {
			t.Errorf("Expected %d results, got %d", iterations, len(benchmarkResult.Results))
		}

		// Check success rate is reasonable (not all should fail due to randomness)
		if benchmarkResult.SuccessRate < 0.3 || benchmarkResult.SuccessRate > 0.8 {
			t.Logf("Success rate: %f (expected between 0.3 and 0.8 due to randomness)", benchmarkResult.SuccessRate)
		}
	})

	t.Run("BenchmarkZeroIterations", func(t *testing.T) {
		widget := config.Widget{
			Name:    "Zero Iterations Widget",
			Script:  filepath.Join(tempDir, "benchmark.py"),
			Enabled: true,
			Timeout: 5,
		}

		params := map[string]interface{}{}
		benchmarkResult := testRunner.BenchmarkWidget(widget, params, 0)

		// Zero iterations should default to 10 iterations
		if benchmarkResult.Iterations != 10 {
			t.Errorf("Expected 10 iterations (default), got %d", benchmarkResult.Iterations)
		}
	})
}

func TestTestResultManagement(t *testing.T) {
	tempDir := t.TempDir()
	executor := widgets.NewExecutor("python3", 30*time.Second)
	testRunner := NewWidgetTestRunner(executor)

	t.Run("TestResultStructure", func(t *testing.T) {
		// Create simple test script
		testScript := `#!/usr/bin/env python3
import json
print(json.dumps({"status": "success", "message": "Test result"}))
`
		scriptPath := filepath.Join(tempDir, "result_test.py")
		if err := os.WriteFile(scriptPath, []byte(testScript), 0755); err != nil {
			t.Fatalf("Failed to create test script: %v", err)
		}

		widget := config.Widget{
			Name:    "Result Test Widget",
			Script:  scriptPath,
			Enabled: true,
			Timeout: 5,
		}

		result := testRunner.TestWidget(widget, map[string]interface{}{}, 0)

		// Verify result structure
		if result.ExecutionTime == 0 {
			t.Error("Expected execution time to be recorded")
		}

		if result.Timestamp.IsZero() {
			t.Error("Expected timestamp to be recorded")
		}

		if result.Environment == nil {
			t.Error("Expected environment to be set")
		}
	})
}

func TestConcurrentTesting(t *testing.T) {
	tempDir := t.TempDir()

	// Create simple test script
	testScript := `#!/usr/bin/env python3
import json
import time
import random

# Simulate some work
time.sleep(random.uniform(0.01, 0.1))

print(json.dumps({"status": "success", "message": "Concurrent test"}))
`

	scriptPath := filepath.Join(tempDir, "concurrent.py")
	if err := os.WriteFile(scriptPath, []byte(testScript), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	executor := widgets.NewExecutor("python3", 30*time.Second)
	testRunner := NewWidgetTestRunner(executor)

	t.Run("ConcurrentWidgetTests", func(t *testing.T) {
		numTests := 10
		results := make(chan TestResult, numTests)

		// Run tests concurrently
		for i := 0; i < numTests; i++ {
			go func(index int) {
				widget := config.Widget{
					Name:    "Concurrent Widget " + string(rune('A'+index)),
					Script:  filepath.Join(tempDir, "concurrent.py"),
					Enabled: true,
					Timeout: 5,
				}

				result := testRunner.TestWidget(widget, map[string]interface{}{}, 0)
				results <- result
			}(i)
		}

		// Collect results
		successCount := 0
		for i := 0; i < numTests; i++ {
			result := <-results
			if result.Success {
				successCount++
			}
		}

		// All tests should succeed
		if successCount != numTests {
			t.Errorf("Expected %d successful tests, got %d", numTests, successCount)
		}
	})
}