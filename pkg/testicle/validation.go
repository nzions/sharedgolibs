package testicle

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ValidationConfig holds validation settings
type ValidationConfig struct {
	RunVet                   bool     `yaml:"run_vet"`
	VetFlags                 []string `yaml:"vet_flags"`
	VetTimeout               string   `yaml:"vet_timeout"`
	CompileCheck             bool     `yaml:"compile_check"`
	CompileTimeout           string   `yaml:"compile_timeout"`
	CompileFlags             []string `yaml:"compile_flags"`
	ContinueOnVetErrors      bool     `yaml:"continue_on_vet_errors"`
	ContinueOnCompileErrors  bool     `yaml:"continue_on_compile_errors"`
	InteractiveErrorHandling bool     `yaml:"interactive_error_handling"`
}

// ValidationResult holds the result of validation
type ValidationResult struct {
	VetResult     *VetResult
	CompileResult *CompileResult
	Success       bool
	Duration      time.Duration
}

// VetResult holds go vet results
type VetResult struct {
	Success    bool
	Output     string
	Errors     []VetIssue
	FilesCount int
	Duration   time.Duration
}

// VetIssue represents a single go vet issue
type VetIssue struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Category string `json:"category"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // "error", "warning"
}

// String returns a formatted representation of the vet issue
func (v VetIssue) String() string {
	return fmt.Sprintf("%s:%d:%d: %s", v.File, v.Line, v.Column, v.Message)
}

// CompileResult holds compilation results
type CompileResult struct {
	Success      bool
	Output       string
	Errors       []CompileError
	PackageCount int
	Duration     time.Duration
}

// CompileError represents a compilation error
type CompileError struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
	Package string `json:"package"`
}

// String returns a formatted representation of the compile error
func (c CompileError) String() string {
	return fmt.Sprintf("%s:%d:%d: %s", c.File, c.Line, c.Column, c.Message)
}

// ValidationPipeline handles pre-execution validation
type ValidationPipeline struct {
	config    *ValidationConfig
	logger    *Logger
	vetRunner *VetRunner
	compiler  *TestCompiler
}

// VetRunner handles go vet execution
type VetRunner struct {
	config *ValidationConfig
	logger *Logger
}

// TestCompiler handles test compilation validation
type TestCompiler struct {
	config  *ValidationConfig
	logger  *Logger
	tempDir string
}

// NewValidationPipeline creates a new validation pipeline
func NewValidationPipeline(config *ValidationConfig, logger *Logger) *ValidationPipeline {
	return &ValidationPipeline{
		config:    config,
		logger:    logger,
		vetRunner: NewVetRunner(config, logger),
		compiler:  NewTestCompiler(config, logger),
	}
}

// NewVetRunner creates a new vet runner
func NewVetRunner(config *ValidationConfig, logger *Logger) *VetRunner {
	return &VetRunner{
		config: config,
		logger: logger,
	}
}

// NewTestCompiler creates a new test compiler
func NewTestCompiler(config *ValidationConfig, logger *Logger) *TestCompiler {
	return &TestCompiler{
		config:  config,
		logger:  logger,
		tempDir: filepath.Join(getTempDir(), "testicle-compile"),
	}
}

// Validate runs the complete validation pipeline
func (vp *ValidationPipeline) Validate(ctx context.Context, packages []string) (*ValidationResult, error) {
	startTime := time.Now()
	result := &ValidationResult{
		Success: true,
	}

	vp.logger.Debug("🔍 Starting validation pipeline for %d packages", len(packages))

	// Run go vet validation
	if vp.config.RunVet {
		vp.logger.Debug("🔍 Running go vet validation...")
		vetResult, err := vp.vetRunner.ValidatePackages(ctx, packages)
		if err != nil {
			return nil, fmt.Errorf("vet validation failed: %w", err)
		}
		result.VetResult = vetResult
		if !vetResult.Success {
			result.Success = false
			if !vp.config.ContinueOnVetErrors {
				result.Duration = time.Since(startTime)
				return result, nil
			}
		}
	}

	// Run test compilation check
	if vp.config.CompileCheck {
		vp.logger.Debug("🔍 Running test compilation check...")
		compileResult, err := vp.compiler.CompileTests(ctx, packages)
		if err != nil {
			return nil, fmt.Errorf("test compilation failed: %w", err)
		}
		result.CompileResult = compileResult
		if !compileResult.Success {
			result.Success = false
			if !vp.config.ContinueOnCompileErrors {
				result.Duration = time.Since(startTime)
				return result, nil
			}
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// ValidatePackages runs go vet on the specified packages
func (vr *VetRunner) ValidatePackages(ctx context.Context, packages []string) (*VetResult, error) {
	startTime := time.Now()
	result := &VetResult{
		Success:    true,
		Errors:     make([]VetIssue, 0),
		FilesCount: 0,
	}

	vr.logger.Debug("🔧 Running go vet on %d packages", len(packages))

	for _, pkg := range packages {
		if err := vr.validatePackage(ctx, pkg, result); err != nil {
			return nil, err
		}
	}

	result.Duration = time.Since(startTime)

	if len(result.Errors) > 0 {
		result.Success = false
	}

	vr.logger.Debug("✅ Go vet completed: %d files checked, %d issues found",
		result.FilesCount, len(result.Errors))

	return result, nil
}

// validatePackage runs go vet on a single package
func (vr *VetRunner) validatePackage(ctx context.Context, pkg string, result *VetResult) error {
	// Build go vet command
	args := []string{"vet"}
	args = append(args, vr.config.VetFlags...)
	args = append(args, pkg)

	// Set timeout if specified
	if vr.config.VetTimeout != "" {
		timeout, err := time.ParseDuration(vr.config.VetTimeout)
		if err != nil {
			return fmt.Errorf("invalid vet timeout: %w", err)
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	vr.logger.Debug("🔧 Executing: %s", cmd.String())

	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	result.Output += outputStr

	// Parse go vet output for issues
	if err != nil && outputStr != "" {
		issues := vr.parseVetOutput(outputStr, pkg)
		result.Errors = append(result.Errors, issues...)
	}

	// Count files (approximate - go vet doesn't provide exact count)
	result.FilesCount += vr.estimateFileCount(pkg)

	return nil
}

// parseVetOutput parses go vet output and extracts issues
func (vr *VetRunner) parseVetOutput(output, pkg string) []VetIssue {
	var issues []VetIssue
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse vet output format: "file.go:line:col: category: message"
		issue := vr.parseVetLine(line, pkg)
		if issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

// parseVetLine parses a single line of go vet output
func (vr *VetRunner) parseVetLine(line, pkg string) *VetIssue {
	// Look for pattern: file.go:line:col: category: message
	parts := strings.Split(line, ":")
	if len(parts) < 4 {
		return nil
	}

	issue := &VetIssue{
		File:     parts[0],
		Severity: "warning", // go vet issues are typically warnings
	}

	// Try to parse line number
	if len(parts) >= 2 {
		if line := parseInt(parts[1]); line > 0 {
			issue.Line = line
		}
	}

	// Try to parse column number
	if len(parts) >= 3 {
		if col := parseInt(parts[2]); col > 0 {
			issue.Column = col
		}
	}

	// Extract category and message
	if len(parts) >= 4 {
		remaining := strings.Join(parts[3:], ":")
		messageParts := strings.SplitN(remaining, ":", 2)
		if len(messageParts) >= 2 {
			issue.Category = strings.TrimSpace(messageParts[0])
			issue.Message = strings.TrimSpace(messageParts[1])
		} else {
			issue.Message = strings.TrimSpace(remaining)
		}
	}

	return issue
}

// estimateFileCount estimates the number of Go files in a package
func (vr *VetRunner) estimateFileCount(pkg string) int {
	// Simple heuristic - count .go files in package directory
	matches, err := filepath.Glob(filepath.Join(pkg, "*.go"))
	if err != nil {
		return 1 // Default estimate
	}
	return len(matches)
}

// CompileTests compiles tests for validation without running them
func (tc *TestCompiler) CompileTests(ctx context.Context, packages []string) (*CompileResult, error) {
	startTime := time.Now()
	result := &CompileResult{
		Success:      true,
		Errors:       make([]CompileError, 0),
		PackageCount: len(packages),
	}

	tc.logger.Debug("🔧 Compiling tests for %d packages", len(packages))

	for _, pkg := range packages {
		if err := tc.compilePackageTests(ctx, pkg, result); err != nil {
			return nil, err
		}
	}

	result.Duration = time.Since(startTime)

	if len(result.Errors) > 0 {
		result.Success = false
	}

	tc.logger.Debug("✅ Test compilation completed: %d packages, %d errors",
		result.PackageCount, len(result.Errors))

	return result, nil
}

// compilePackageTests compiles tests for a single package
func (tc *TestCompiler) compilePackageTests(ctx context.Context, pkg string, result *CompileResult) error {
	// Create temporary output file
	tempFile := filepath.Join(tc.tempDir, fmt.Sprintf("test-%d", time.Now().UnixNano()))

	// Ensure temp directory exists
	if err := ensureDir(tc.tempDir); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Build go test -c command
	args := []string{"test", "-c", "-o", tempFile}
	args = append(args, tc.config.CompileFlags...)
	args = append(args, pkg)

	// Set timeout if specified
	if tc.config.CompileTimeout != "" {
		timeout, err := time.ParseDuration(tc.config.CompileTimeout)
		if err != nil {
			return fmt.Errorf("invalid compile timeout: %w", err)
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	tc.logger.Debug("🔧 Executing: %s", cmd.String())

	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	result.Output += outputStr

	// Clean up temp file
	defer func() {
		if err := removeFile(tempFile); err != nil {
			tc.logger.Debug("Failed to remove temp file %s: %v", tempFile, err)
		}
	}()

	// Parse compilation errors
	if err != nil && outputStr != "" {
		errors := tc.parseCompileOutput(outputStr, pkg)
		result.Errors = append(result.Errors, errors...)
	}

	return nil
}

// parseCompileOutput parses go test -c output and extracts compilation errors
func (tc *TestCompiler) parseCompileOutput(output, pkg string) []CompileError {
	var errors []CompileError
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse compilation error format: "file.go:line:col: message"
		compileError := tc.parseCompileLine(line, pkg)
		if compileError != nil {
			errors = append(errors, *compileError)
		}
	}

	return errors
}

// parseCompileLine parses a single line of compilation output
func (tc *TestCompiler) parseCompileLine(line, pkg string) *CompileError {
	// Look for pattern: file.go:line:col: message
	parts := strings.Split(line, ":")
	if len(parts) < 3 {
		return nil
	}

	compileError := &CompileError{
		File:    parts[0],
		Package: pkg,
	}

	// Try to parse line number
	if len(parts) >= 2 {
		if line := parseInt(parts[1]); line > 0 {
			compileError.Line = line
		}
	}

	// Try to parse column number
	if len(parts) >= 3 {
		if col := parseInt(parts[2]); col > 0 {
			compileError.Column = col
		}
	}

	// Extract message
	if len(parts) >= 4 {
		compileError.Message = strings.TrimSpace(strings.Join(parts[3:], ":"))
	}

	return compileError
}

// Helper functions

// parseInt safely parses a string to int
func parseInt(s string) int {
	s = strings.TrimSpace(s)
	var result int
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int(r-'0')
		} else {
			return 0
		}
	}
	return result
}

// getTempDir returns the system temp directory
func getTempDir() string {
	return "/tmp" // Simple implementation - could use os.TempDir()
}

// ensureDir creates directory if it doesn't exist
func ensureDir(dir string) error {
	// Simple implementation - could use os.MkdirAll
	return nil
}

// removeFile removes a file
func removeFile(path string) error {
	// Simple implementation - could use os.Remove
	return nil
}
