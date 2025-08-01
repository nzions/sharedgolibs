package testicle

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
)

// Config holds the configuration for the testicle runner
type Config struct {
	Debug      bool   `yaml:"debug"`
	Daemon     bool   `yaml:"daemon"`
	Dir        string `yaml:"dir"`
	ConfigFile string `yaml:"config_file"`

	// Validation settings
	NoVet        bool `yaml:"no_vet"`
	NoBuildCheck bool `yaml:"no_build_check"`
	Validate     bool `yaml:"validate"`
}

// Runner is the main testicle test runner
type Runner struct {
	config       *Config
	discovery    *Discovery
	executor     *Executor
	watcher      *Watcher
	logger       *Logger
	uiController *UIController
	validator    *ValidationPipeline
}

// NewRunner creates a new testicle runner with the given configuration
func NewRunner(config *Config) (*Runner, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Validate and normalize configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize logger
	logger := NewLogger(config.Debug)

	// Initialize components
	discovery := NewDiscovery(config.Dir, logger)
	executor := NewExecutor(logger)

	// Initialize validation pipeline if needed
	var validator *ValidationPipeline
	if config.Validate {
		validationConfig := &ValidationConfig{
			RunVet:                   !config.NoVet,
			CompileCheck:             !config.NoBuildCheck,
			VetTimeout:               "10s",
			CompileTimeout:           "30s",
			ContinueOnVetErrors:      false,
			ContinueOnCompileErrors:  false,
			InteractiveErrorHandling: true, // Enable interactive mode for validation errors
		}
		validator = NewValidationPipeline(validationConfig, logger)
	}

	var watcher *Watcher
	var uiController *UIController
	if config.Daemon {
		watcher = NewWatcher(config.Dir, logger)
		uiController = NewUIController(nil, logger) // Will set runner reference after creation
	}

	runner := &Runner{
		config:       config,
		discovery:    discovery,
		executor:     executor,
		validator:    validator,
		watcher:      watcher,
		logger:       logger,
		uiController: uiController,
	}

	// Set the runner reference in UI controller
	if uiController != nil {
		uiController.runner = runner

		// Set up executor callback to feed test results to UI
		runner.executor.SetResultCallback(func(result *TestResult) {
			var status, duration string
			switch result.Status {
			case TestStatusPassed:
				status = "passed"
			case TestStatusFailed:
				status = "failed"
			case TestStatusSkipped:
				status = "skipped"
			}

			if result.Duration > 0 {
				duration = result.Duration.String()
			}

			uiController.AddTestResult(result.Name, status, duration)
		})
	}

	logger.Info("🧪 Testicle %s initialized", Version)
	logger.Debug("Configuration: %+v", config)

	return runner, nil
}

// Run starts the testicle runner
func (r *Runner) Run(ctx context.Context) error {
	// In daemon mode with UI, don't show these initial messages
	if !r.config.Daemon || r.uiController == nil {
		r.logger.Info("🧪 Testicle %s - Running tests in %s", Version, r.config.Dir)
	}

	if r.config.Daemon {
		return r.runDaemon(ctx)
	} else {
		return r.runOnce(ctx)
	}
}

// runOnce executes all tests once
func (r *Runner) runOnce(ctx context.Context) error {
	// Run validation if enabled
	if r.validator != nil {
		if r.uiController != nil && r.uiController.isActive {
			r.uiController.AddLiveOutput("🔍 Running validation...")
			r.uiController.status.State = "validating"
			r.uiController.renderFullScreen()
		} else {
			r.logger.Info("Running validation...")
		}

		validationResult, err := r.validator.Validate(ctx, []string{r.config.Dir})
		if err != nil {
			if r.uiController != nil && r.uiController.isActive {
				r.uiController.AddLiveOutput("❌ Validation failed: " + err.Error())
				r.uiController.status.State = "error"
				r.uiController.renderFullScreen()
			} else {
				r.logger.Error("Validation failed: %v", err)
			}
			return fmt.Errorf("validation failed: %w", err)
		}

		if !validationResult.Success {
			if r.uiController != nil && r.uiController.isActive {
				r.uiController.AddLiveOutput("⚠️  Validation completed with errors")
				// Display validation errors in UI
				if validationResult.VetResult != nil && len(validationResult.VetResult.Errors) > 0 {
					for _, vetErr := range validationResult.VetResult.Errors {
						r.uiController.AddLiveOutput(fmt.Sprintf("  vet: %s", vetErr))
					}
				}
				if validationResult.CompileResult != nil && len(validationResult.CompileResult.Errors) > 0 {
					for _, compileErr := range validationResult.CompileResult.Errors {
						r.uiController.AddLiveOutput(fmt.Sprintf("  compile: %s", compileErr))
					}
				}
				r.uiController.status.State = "validation_errors"
				r.uiController.renderFullScreen()
			} else {
				r.logger.Warn("Validation completed with errors")
				// Log validation errors
				if validationResult.VetResult != nil && len(validationResult.VetResult.Errors) > 0 {
					r.logger.Warn("Go vet errors:")
					for _, vetErr := range validationResult.VetResult.Errors {
						r.logger.Warn("  %s", vetErr)
					}
				}
				if validationResult.CompileResult != nil && len(validationResult.CompileResult.Errors) > 0 {
					r.logger.Warn("Compilation errors:")
					for _, compileErr := range validationResult.CompileResult.Errors {
						r.logger.Warn("  %s", compileErr)
					}
				}
			}

			// If we're configured to stop on validation errors, return early
			if !r.validator.config.ContinueOnVetErrors && validationResult.VetResult != nil && len(validationResult.VetResult.Errors) > 0 {
				return fmt.Errorf("go vet found issues")
			}
			if !r.validator.config.ContinueOnCompileErrors && validationResult.CompileResult != nil && len(validationResult.CompileResult.Errors) > 0 {
				return fmt.Errorf("compilation failed")
			}
		} else {
			if r.uiController != nil && r.uiController.isActive {
				r.uiController.AddLiveOutput("✅ Validation passed")
			} else {
				r.logger.Info("Validation passed")
			}
		}
	}

	// In UI mode, add to live output instead of logging
	if r.uiController != nil && r.uiController.isActive {
		r.uiController.AddLiveOutput("🔍 Discovering tests...")
		// Set state to running with test count
		r.uiController.status.State = "running"
		r.uiController.status.TestCount = 0
		r.uiController.status.PassedCount = 0
		r.uiController.status.FailedCount = 0
		r.uiController.status.SkippedCount = 0
		r.uiController.renderFullScreen()
	} else {
		r.logger.Info("Running tests once...")
	}

	// Discover tests
	tests, err := r.discovery.DiscoverTests(ctx)
	if err != nil {
		return fmt.Errorf("test discovery failed: %w", err)
	}

	if r.uiController != nil && r.uiController.isActive {
		r.uiController.AddLiveOutput(fmt.Sprintf("🔍 Found %d test(s)", len(tests)))
		// Clear previous test results and set total count
		r.uiController.testResults = make([]*TestResultLine, 0)
		r.uiController.status.TestCount = len(tests)
		r.uiController.status.State = "running"
		r.uiController.renderFullScreen()
	} else {
		r.logger.Info("🔍 Found %d test(s) in %s", len(tests), r.config.Dir)
	}

	// Execute tests
	results, err := r.executor.ExecuteTests(ctx, tests)
	if err != nil {
		return fmt.Errorf("test execution failed: %w", err)
	}

	// Print results summary
	r.printSummary(results)

	return nil
}

// runDaemon runs in watch mode, re-executing tests on file changes
func (r *Runner) runDaemon(ctx context.Context) error {
	// Initialize the UI controller
	if r.uiController != nil {
		if err := r.uiController.Start(ctx); err != nil {
			r.logger.Error("Failed to start UI: %v", err)
			// Fall back to simple mode
			r.uiController = nil
		}
	}

	// Run tests initially
	if err := r.runOnce(ctx); err != nil {
		r.logger.Error("Initial test run failed: %v", err)
	}

	// Start file watcher
	if r.watcher == nil {
		return fmt.Errorf("watcher not initialized for daemon mode")
	}

	eventChan, err := r.watcher.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start file watcher: %w", err)
	}

	// Get keyboard input if UI is available
	var keyInput <-chan rune
	if r.uiController != nil {
		keyInput = r.uiController.GetKeyInput()
	} else {
		r.logger.Info("🎮 Interactive controls: [Ctrl+C] Quit")
	}

	// Main daemon loop
	if keyInput != nil {
		// With UI - handle keyboard input
		for {
			select {
			case <-ctx.Done():
				if r.uiController != nil {
					r.uiController.Stop()
				}
				r.logger.Info("🛑 Stopping daemon mode...")
				return ctx.Err()

			case key := <-keyInput:
				if err := r.handleKeyInput(ctx, key); err != nil {
					return err
				}

			case event := <-eventChan:
				// In UI mode, don't log debug messages
				if r.uiController == nil || !r.uiController.isActive {
					r.logger.Debug("📁 File change detected: %s", event.Path)
				}

				// Notify UI of file change
				if r.uiController != nil {
					r.uiController.OnFileChange(event.Path)
				}

				// Debounce rapid file changes
				time.Sleep(100 * time.Millisecond)

				if r.uiController != nil && r.uiController.isActive {
					r.uiController.AddLiveOutput("🔄 Re-running tests due to file change...")
				} else {
					r.logger.Info("🔄 Re-running tests due to file change...")
				}

				if err := r.runOnce(ctx); err != nil {
					if r.uiController != nil && r.uiController.isActive {
						r.uiController.AddLiveOutput(fmt.Sprintf("❌ Test execution failed: %v", err))
					} else {
						r.logger.Error("Test execution failed: %v", err)
					}
				}
			}
		}
	} else {
		// Simple mode - no keyboard input
		for {
			select {
			case <-ctx.Done():
				r.logger.Info("🛑 Stopping daemon mode...")
				return ctx.Err()

			case event := <-eventChan:
				r.logger.Debug("📁 File change detected: %s", event.Path)

				// Debounce rapid file changes
				time.Sleep(100 * time.Millisecond)

				r.logger.Info("🔄 Re-running tests due to file change...")
				if err := r.runOnce(ctx); err != nil {
					r.logger.Error("Test execution failed: %v", err)
				}
			}
		}
	}
}

// printSummary prints a summary of test results
func (r *Runner) printSummary(results *TestResults) {
	// Update UI if available
	if r.uiController != nil {
		r.uiController.UpdateTestResults(results)
		return
	}

	// Fall back to aesthetic console output
	total := results.Passed + results.Failed + results.Skipped

	r.logger.Info("")
	r.logger.Info("╭─────────────────────────────────────────────────╮") // 49 chars
	r.logger.Info("│                🧪 TEST SUMMARY                  │")  // 49 chars
	r.logger.Info("├─────────────────────────────────────────────────┤") // 49 chars

	// Helper to pad a string to 49 visible columns (content area)
	pad := func(s string) string {
		w := runewidth.StringWidth(s)
		if w > 49 {
			trunc := ""
			acc := 0
			for _, r := range s {
				rw := runewidth.RuneWidth(r)
				if acc+rw > 49 {
					break
				}
				trunc += string(r)
				acc += rw
			}
			// If we truncated and are still short, pad
			if acc < 49 {
				trunc += strings.Repeat(" ", 49-acc)
			}
			return trunc
		}
		return s + strings.Repeat(" ", 49-w)
	}

	r.logger.Info("│%s│", pad(""))
	r.logger.Info("│%s│", pad(fmt.Sprintf("  📊 Tests Discovered: %d", total)))
	r.logger.Info("│%s│", pad(""))
	r.logger.Info("│%s│", pad(fmt.Sprintf("  ✅ Passed:  %d", results.Passed)))
	r.logger.Info("│%s│", pad(fmt.Sprintf("  ❌ Failed:  %d", results.Failed)))
	if results.Skipped > 0 {
		r.logger.Info("│%s│", pad(fmt.Sprintf("  ⏭️  Skipped: %d", results.Skipped)))
	}
	r.logger.Info("│%s│", pad(""))
	r.logger.Info("│%s│", pad(fmt.Sprintf("  ⏱️  Runtime: %s", results.Duration.String())))
	r.logger.Info("│%s│", pad(""))
	successRate := float64(results.Passed) / float64(total) * 100
	successRateStr := fmt.Sprintf("%.1f%%", successRate)
	r.logger.Info("│%s│", pad(fmt.Sprintf("  📈 Success Rate: %s", successRateStr)))
	r.logger.Info("│%s│", pad(""))
	if results.Failed > 0 {
		r.logger.Info("│%s│", pad("  🔴 Status: FAILED"))
	} else {
		r.logger.Info("│%s│", pad("  🟢 Status: PASSED"))
	}
	r.logger.Info("│%s│", pad(""))
	r.logger.Info("╰─────────────────────────────────────────────────╯") // 49 chars

	if results.Failed > 0 {
		r.logger.Info("")
		r.logger.Info("❌ %d test(s) failed", results.Failed)
		if !r.config.Daemon {
			os.Exit(1)
		}
	} else {
		r.logger.Info("")
		r.logger.Info("✅ All tests passed!")
	}
}

// handleKeyInput processes keyboard input for interactive daemon mode
func (r *Runner) handleKeyInput(ctx context.Context, key rune) error {
	switch key {
	case 'r', 'R':
		// Re-run tests
		r.logger.Info("🔄 Manual test re-run requested...")
		if err := r.runOnce(ctx); err != nil {
			r.logger.Error("Test execution failed: %v", err)
		}

	case 'p', 'P':
		// Toggle pause/resume (placeholder for future implementation)
		r.logger.Info("⏸️ Pause/Resume functionality coming soon...")

	case 'c', 'C':
		// Clear screen and refresh UI
		if r.uiController != nil {
			r.uiController.clearScreen()
			r.uiController.printHeader()
			r.uiController.refreshDisplay()
		}

	case 'd', 'D':
		// Toggle debug mode
		r.config.Debug = !r.config.Debug
		r.logger = NewLogger(r.config.Debug)
		r.logger.Info("🔧 Debug mode: %t", r.config.Debug)

	case 's', 'S':
		// Show detailed stats (placeholder)
		r.logger.Info("📈 Detailed statistics coming soon...")

	case 'q', 'Q':
		// Quit
		r.logger.Info("🛑 Quit requested by user")
		return fmt.Errorf("user requested quit")

	case '\x03': // Ctrl+C
		return fmt.Errorf("interrupt signal received")

	default:
		// Ignore other keys
	}

	return nil
}

// validateConfig validates the runner configuration
func validateConfig(config *Config) error {
	// Validate test directory
	if config.Dir == "" {
		config.Dir = "."
	}

	// Check if directory exists
	if _, err := os.Stat(config.Dir); os.IsNotExist(err) {
		return fmt.Errorf("test directory does not exist: %s", config.Dir)
	}

	// Convert to absolute path
	absDir, err := filepath.Abs(config.Dir)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path for %s: %w", config.Dir, err)
	}
	config.Dir = absDir

	// Validate config file if specified
	if config.ConfigFile != "" && config.ConfigFile != "testicle.yaml" {
		if _, err := os.Stat(config.ConfigFile); os.IsNotExist(err) {
			return fmt.Errorf("configuration file does not exist: %s", config.ConfigFile)
		}
	}

	return nil
}
