package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/nzions/sharedgolibs/pkg/testicle/discovery"
)

// TestStatus represents the current status of a test
type TestStatus string

const (
	StatusPending TestStatus = "pending"
	StatusRunning TestStatus = "running"
	StatusPassed  TestStatus = "passed"
	StatusFailed  TestStatus = "failed"
	StatusSkipped TestStatus = "skipped"
)

// TestResult contains the result of a test execution
type TestResult struct {
	TestName  string        `json:"test_name"`
	Package   string        `json:"package"`
	Status    TestStatus    `json:"status"`
	Duration  time.Duration `json:"duration"`
	Output    string        `json:"output"`
	Error     string        `json:"error,omitempty"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Coverage  float64       `json:"coverage,omitempty"`
}

// TestRun represents a single test execution session
type TestRun struct {
	ID        string        `json:"id"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Status    TestStatus    `json:"status"`
	Results   []*TestResult `json:"results"`
	Summary   *TestSummary  `json:"summary"`
	Config    *Config       `json:"config"`
}

// TestSummary provides aggregate statistics for a test run
type TestSummary struct {
	Total    int           `json:"total"`
	Passed   int           `json:"passed"`
	Failed   int           `json:"failed"`
	Skipped  int           `json:"skipped"`
	Duration time.Duration `json:"duration"`
	Coverage float64       `json:"coverage"`
}

// Config represents the configuration for the test runner
type Config struct {
	ProjectRoot  string   `json:"project_root"`
	TestPackages []string `json:"test_packages"`
	TestPattern  string   `json:"test_pattern"`
	Parallel     int      `json:"parallel"`
	Timeout      string   `json:"timeout"`
	Coverage     bool     `json:"coverage"`
	Verbose      bool     `json:"verbose"`
	Tags         []string `json:"tags"`
	Race         bool     `json:"race"`
	FailFast     bool     `json:"fail_fast"`
}

// Runner is the core test execution engine
type Runner struct {
	config     *Config
	discoverer *discovery.Discoverer
	results    []*TestResult
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc

	// Event channels for real-time updates
	statusChan chan TestResult
	doneChan   chan TestRun
}

// NewRunner creates a new test runner with the given configuration
func NewRunner(config *Config) *Runner {
	if config.ProjectRoot == "" {
		config.ProjectRoot = "."
	}
	if config.Parallel == 0 {
		config.Parallel = 1
	}
	if config.TestPattern == "" {
		config.TestPattern = "."
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Runner{
		config:     config,
		discoverer: discovery.NewDiscoverer(config.ProjectRoot),
		results:    make([]*TestResult, 0),
		ctx:        ctx,
		cancel:     cancel,
		statusChan: make(chan TestResult, 100),
		doneChan:   make(chan TestRun, 1),
	}
}

// RunTests executes all discovered tests
func (r *Runner) RunTests() (*TestRun, error) {
	startTime := time.Now()

	// Discover tests
	suites, err := r.discoverer.DiscoverTests()
	if err != nil {
		return nil, fmt.Errorf("discovering tests: %w", err)
	}

	// Create test run
	run := &TestRun{
		ID:        fmt.Sprintf("run-%d", startTime.Unix()),
		StartTime: startTime,
		Status:    StatusRunning,
		Results:   make([]*TestResult, 0),
		Config:    r.config,
	}

	// Execute tests
	for _, suite := range suites {
		if len(suite.Tests) == 0 {
			continue
		}

		// Filter tests by tags if specified
		tests := r.filterTestsByTags(suite.Tests)
		if len(tests) == 0 {
			continue
		}

		// Run tests in the package
		packageResults, err := r.runPackageTests(suite.Package, suite.Path, tests)
		if err != nil {
			return nil, fmt.Errorf("running tests for package %s: %w", suite.Package, err)
		}

		run.Results = append(run.Results, packageResults...)
	}

	// Calculate summary
	run.EndTime = time.Now()
	run.Summary = r.calculateSummary(run.Results)
	run.Status = r.determineOverallStatus(run.Results)

	// Send completion notification
	select {
	case r.doneChan <- *run:
	default:
	}

	return run, nil
}

// runPackageTests executes tests for a specific package
func (r *Runner) runPackageTests(packageName, packagePath string, tests []*discovery.TestFunction) ([]*TestResult, error) {
	var results []*TestResult

	// Build go test command
	cmd := r.buildTestCommand(packagePath, tests)

	startTime := time.Now()
	output, err := cmd.CombinedOutput()
	endTime := time.Now()

	// Parse the output to determine individual test results
	// For now, we'll create a simple result based on command success/failure
	for _, test := range tests {
		result := &TestResult{
			TestName:  test.Name,
			Package:   packageName,
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
			Output:    string(output),
		}

		if err != nil {
			result.Status = StatusFailed
			result.Error = err.Error()
		} else {
			result.Status = StatusPassed
		}

		results = append(results, result)

		// Send real-time update
		select {
		case r.statusChan <- *result:
		default:
		}
	}

	return results, nil
}

// buildTestCommand constructs the go test command with appropriate flags
func (r *Runner) buildTestCommand(packagePath string, tests []*discovery.TestFunction) *exec.Cmd {
	args := []string{"test"}

	if r.config.Verbose {
		args = append(args, "-v")
	}

	if r.config.Coverage {
		args = append(args, "-cover")
	}

	if r.config.Race {
		args = append(args, "-race")
	}

	if r.config.Timeout != "" {
		args = append(args, "-timeout", r.config.Timeout)
	}

	if r.config.Parallel > 1 {
		args = append(args, "-parallel", fmt.Sprintf("%d", r.config.Parallel))
	}

	// Add test pattern if specific tests are requested
	if len(tests) > 0 && len(tests) < 10 { // Only for small sets to avoid command line length issues
		testNames := make([]string, len(tests))
		for i, test := range tests {
			testNames[i] = test.Name
		}
		args = append(args, "-run", fmt.Sprintf("^(%s)$", joinWithOr(testNames)))
	}

	// Add package path
	args = append(args, packagePath)

	cmd := exec.CommandContext(r.ctx, "go", args...)
	cmd.Dir = r.config.ProjectRoot
	cmd.Env = os.Environ()

	return cmd
}

// filterTestsByTags filters tests based on configured tags
func (r *Runner) filterTestsByTags(tests []*discovery.TestFunction) []*discovery.TestFunction {
	if len(r.config.Tags) == 0 {
		return tests
	}

	var filtered []*discovery.TestFunction
	for _, test := range tests {
		if r.hasMatchingTags(test.Tags) {
			filtered = append(filtered, test)
		}
	}

	return filtered
}

// hasMatchingTags checks if a test has any of the configured tags
func (r *Runner) hasMatchingTags(testTags []string) bool {
	for _, configTag := range r.config.Tags {
		for _, testTag := range testTags {
			if configTag == testTag {
				return true
			}
		}
	}
	return len(r.config.Tags) == 0 // If no tags configured, include all tests
}

// calculateSummary computes aggregate statistics for test results
func (r *Runner) calculateSummary(results []*TestResult) *TestSummary {
	summary := &TestSummary{}

	var totalDuration time.Duration
	for _, result := range results {
		summary.Total++
		totalDuration += result.Duration

		switch result.Status {
		case StatusPassed:
			summary.Passed++
		case StatusFailed:
			summary.Failed++
		case StatusSkipped:
			summary.Skipped++
		}
	}

	summary.Duration = totalDuration

	return summary
}

// determineOverallStatus determines the overall status of a test run
func (r *Runner) determineOverallStatus(results []*TestResult) TestStatus {
	for _, result := range results {
		if result.Status == StatusFailed {
			return StatusFailed
		}
	}
	return StatusPassed
}

// StatusChannel returns the channel for real-time test status updates
func (r *Runner) StatusChannel() <-chan TestResult {
	return r.statusChan
}

// DoneChannel returns the channel that signals test run completion
func (r *Runner) DoneChannel() <-chan TestRun {
	return r.doneChan
}

// Stop cancels the current test execution
func (r *Runner) Stop() {
	r.cancel()
}

// joinWithOr joins strings with | for regex alternation
func joinWithOr(strings []string) string {
	if len(strings) == 0 {
		return ""
	}
	if len(strings) == 1 {
		return strings[0]
	}

	result := strings[0]
	for i := 1; i < len(strings); i++ {
		result += "|" + strings[i]
	}
	return result
}
