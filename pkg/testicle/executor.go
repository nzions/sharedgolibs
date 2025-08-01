package testicle

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// TestResults holds the results of test execution
type TestResults struct {
	Passed   int
	Failed   int
	Skipped  int
	Duration time.Duration
	Tests    []*TestResult
}

// TestResult holds the result of a single test
type TestResult struct {
	Name     string
	Package  string
	Status   TestStatus
	Duration time.Duration
	Output   string
	Error    string
}

// TestStatus represents the status of a test
type TestStatus int

const (
	TestStatusPassed TestStatus = iota
	TestStatusFailed
	TestStatusSkipped
)

// TestResultCallback is called for each individual test result
type TestResultCallback func(result *TestResult)

// Executor handles test execution
type Executor struct {
	logger         *Logger
	resultCallback TestResultCallback
}

// NewExecutor creates a new test executor
func NewExecutor(logger *Logger) *Executor {
	return &Executor{
		logger:         logger,
		resultCallback: nil,
	}
}

// SetResultCallback sets a callback function to be called for each test result
func (e *Executor) SetResultCallback(callback TestResultCallback) {
	e.resultCallback = callback
}

// ExecuteTests executes the discovered tests
func (e *Executor) ExecuteTests(ctx context.Context, tests []*TestInfo) (*TestResults, error) {
	e.logger.Info("🚀 Executing %d test(s)...", len(tests))

	startTime := time.Now()
	results := &TestResults{
		Tests: make([]*TestResult, 0, len(tests)),
	}

	// Group tests by package for efficient execution
	packageTests := e.groupTestsByPackage(tests)

	for packagePath, packageTestList := range packageTests {
		e.logger.Debug("📦 Running tests in package: %s", packagePath)

		packageResults, err := e.executePackageTests(ctx, packagePath, packageTestList)
		if err != nil {
			e.logger.Error("Failed to execute tests in package %s: %v", packagePath, err)
			continue
		}

		// Merge results
		results.Tests = append(results.Tests, packageResults.Tests...)
		results.Passed += packageResults.Passed
		results.Failed += packageResults.Failed
		results.Skipped += packageResults.Skipped
	}

	results.Duration = time.Since(startTime)
	return results, nil
}

// groupTestsByPackage groups tests by their package directory
func (e *Executor) groupTestsByPackage(tests []*TestInfo) map[string][]*TestInfo {
	packages := make(map[string][]*TestInfo)

	for _, test := range tests {
		packageDir := filepath.Dir(test.File)
		packages[packageDir] = append(packages[packageDir], test)
	}

	return packages
}

// executePackageTests executes all tests in a specific package
func (e *Executor) executePackageTests(ctx context.Context, packagePath string, tests []*TestInfo) (*TestResults, error) {
	// For now, we'll run `go test` on the package
	// In the future, we could implement more sophisticated test selection

	cmd := exec.CommandContext(ctx, "go", "test", "-v", packagePath)

	e.logger.Debug("🔧 Executing: %s", cmd.String())

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Parse the go test output to extract individual test results
	results := e.parseGoTestOutput(outputStr, tests)

	if err != nil {
		// Mark tests as failed if the command failed
		for _, result := range results.Tests {
			if result.Status == TestStatusPassed {
				result.Status = TestStatusFailed
				if result.Error == "" {
					result.Error = err.Error()
				}
				results.Failed++
				results.Passed--
			}
		}
	}

	return results, nil
}

// parseGoTestOutput parses the output from `go test -v` and extracts test results
func (e *Executor) parseGoTestOutput(output string, tests []*TestInfo) *TestResults {
	results := &TestResults{
		Tests: make([]*TestResult, 0, len(tests)),
	}

	lines := strings.Split(output, "\n")
	testMap := make(map[string]*TestResult)

	// Initialize results for all tests
	for _, test := range tests {
		result := &TestResult{
			Name:     test.Name,
			Package:  test.Package,
			Status:   TestStatusFailed, // Default to failed, mark as passed if we see success
			Duration: 0,
			Output:   "",
			Error:    "",
		}
		testMap[test.Name] = result
		results.Tests = append(results.Tests, result)
	}

	// Parse output lines
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for test result patterns
		if strings.Contains(line, "PASS:") || strings.Contains(line, "--- PASS:") {
			testName := e.extractTestName(line)
			if result, exists := testMap[testName]; exists {
				result.Status = TestStatusPassed
				result.Duration = e.extractDuration(line)
			}
		} else if strings.Contains(line, "FAIL:") || strings.Contains(line, "--- FAIL:") {
			testName := e.extractTestName(line)
			if result, exists := testMap[testName]; exists {
				result.Status = TestStatusFailed
				result.Error = line
			}
		} else if strings.Contains(line, "SKIP:") || strings.Contains(line, "--- SKIP:") {
			testName := e.extractTestName(line)
			if result, exists := testMap[testName]; exists {
				result.Status = TestStatusSkipped
			}
		}
	}

	// Count results and notify callback
	for _, result := range results.Tests {
		switch result.Status {
		case TestStatusPassed:
			results.Passed++
		case TestStatusFailed:
			results.Failed++
		case TestStatusSkipped:
			results.Skipped++
		}

		// Use callback if available, otherwise log
		if e.resultCallback != nil {
			e.resultCallback(result)
		} else {
			e.logTestResult(result)
		}
	}

	return results
}

// extractTestName extracts the test name from a go test output line
func (e *Executor) extractTestName(line string) string {
	// Look for patterns like "--- PASS: TestName" or "PASS: TestName"
	if idx := strings.Index(line, ":"); idx != -1 {
		remainder := strings.TrimSpace(line[idx+1:])
		if spaceIdx := strings.Index(remainder, " "); spaceIdx != -1 {
			return remainder[:spaceIdx]
		}
		return remainder
	}
	return ""
}

// extractDuration extracts the duration from a go test output line
func (e *Executor) extractDuration(line string) time.Duration {
	// Look for patterns like "(0.00s)"
	if strings.Contains(line, "(") && strings.Contains(line, "s)") {
		start := strings.LastIndex(line, "(")
		end := strings.LastIndex(line, "s)")
		if start != -1 && end != -1 && end > start {
			durationStr := line[start+1 : end+1]
			if duration, err := time.ParseDuration(durationStr); err == nil {
				return duration
			}
		}
	}
	return 0
}

// logTestResult logs the result of an individual test
func (e *Executor) logTestResult(result *TestResult) {
	switch result.Status {
	case TestStatusPassed:
		if result.Duration > 0 {
			e.logger.Info("✅ %s (%s)", result.Name, result.Duration)
		} else {
			e.logger.Info("✅ %s", result.Name)
		}
	case TestStatusFailed:
		e.logger.Info("❌ %s", result.Name)
		if result.Error != "" {
			e.logger.Info("   %s", result.Error)
		}
	case TestStatusSkipped:
		e.logger.Info("⏭️  %s (skipped)", result.Name)
	}
}
