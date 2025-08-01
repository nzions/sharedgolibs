package testicle

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/term"
)

// UIController manages the interactive console UI for daemon mode
type UIController struct {
	logger       *Logger
	runner       *Runner
	status       *DaemonStatus
	keyHandler   *KeyHandler
	screenHeight int
	screenWidth  int
	isActive     bool
	testResults  []*TestResultLine
	liveOutput   []string
	maxOutput    int
}

// TestResultLine represents a single test result for display
type TestResultLine struct {
	Name     string
	Status   string // "running", "passed", "failed", "queued"
	Duration string
	Progress int // 0-100 for running tests
}

// DaemonStatus tracks the current state of the daemon
type DaemonStatus struct {
	State        string    `json:"state"` // "watching", "running", "paused", "failed"
	LastRun      time.Time `json:"last_run"`
	TestCount    int       `json:"test_count"`
	PassedCount  int       `json:"passed_count"`
	FailedCount  int       `json:"failed_count"`
	SkippedCount int       `json:"skipped_count"`
	RunningCount int       `json:"running_count"`
	QueuedCount  int       `json:"queued_count"`
	Duration     string    `json:"duration"`
	ElapsedTime  string    `json:"elapsed_time"`
	WatchedFiles int       `json:"watched_files"`
	FileChanges  int       `json:"file_changes"`
	ProgressPct  int       `json:"progress_pct"`
}

// KeyHandler manages keyboard input for interactive controls
type KeyHandler struct {
	enabled      bool
	inputChan    chan rune
	cancelFunc   context.CancelFunc
	originalTerm *term.State
}

// NewUIController creates a new UI controller for daemon mode
func NewUIController(runner *Runner, logger *Logger) *UIController {
	return &UIController{
		logger:       logger,
		runner:       runner,
		screenHeight: 25,
		screenWidth:  80,
		isActive:     false,
		testResults:  make([]*TestResultLine, 0),
		liveOutput:   make([]string, 0),
		maxOutput:    5, // Show last 5 lines of output
		status: &DaemonStatus{
			State:   "initializing",
			LastRun: time.Now(),
		},
		keyHandler: &KeyHandler{
			enabled:   true,
			inputChan: make(chan rune, 10),
		},
	}
}

// Start initializes the interactive UI
func (ui *UIController) Start(ctx context.Context) error {
	ui.initializeScreen()
	ui.isActive = true
	ui.renderFullScreen()

	// Start keyboard input handler
	go ui.handleKeyboardInput(ctx)

	ui.status.State = "watching"
	return nil
}

// initializeScreen sets up the terminal for fixed-position rendering
func (ui *UIController) initializeScreen() {
	// Hide cursor and clear screen
	fmt.Print("\033[?25l\033[2J\033[H")

	// Get terminal size if possible
	ui.getTerminalSize()
}

// getTerminalSize attempts to get the current terminal dimensions
func (ui *UIController) getTerminalSize() {
	// Get actual terminal size using the term package
	fd := int(os.Stdout.Fd())
	if term.IsTerminal(fd) {
		width, height, err := term.GetSize(fd)
		if err == nil {
			ui.screenWidth = width
			ui.screenHeight = height
			return
		}
	}

	// Fallback to reasonable defaults if we can't get terminal size
	ui.screenWidth = 80
	ui.screenHeight = 25
}

// renderFullScreen renders the complete TUI interface
func (ui *UIController) renderFullScreen() {
	if !ui.isActive {
		return
	}

	// Refresh terminal size in case window was resized
	ui.getTerminalSize()

	// Save cursor position and clear screen
	fmt.Print("\033[s\033[2J\033[H")

	ui.renderHeader()
	ui.renderStatus()
	ui.renderTestResults()
	ui.renderLiveOutput()
	ui.renderControls()

	// Restore cursor position
	fmt.Print("\033[u")
} // UpdateStatus updates the daemon status and refreshes the display
func (ui *UIController) UpdateStatus(status *DaemonStatus) {
	ui.status = status
	ui.renderFullScreen()
}

// UpdateTestResults updates the UI with latest test results
func (ui *UIController) UpdateTestResults(results *TestResults) {
	ui.status.TestCount = results.Passed + results.Failed + results.Skipped
	ui.status.PassedCount = results.Passed
	ui.status.FailedCount = results.Failed
	ui.status.SkippedCount = results.Skipped
	ui.status.Duration = results.Duration.String()
	ui.status.LastRun = time.Now()

	if results.Failed > 0 {
		ui.status.State = "failed"
	} else {
		ui.status.State = "watching"
	}

	ui.renderFullScreen()
}

// AddTestResult adds a test result to the display list
func (ui *UIController) AddTestResult(name, status, duration string) {
	result := &TestResultLine{
		Name:     name,
		Status:   status,
		Duration: duration,
		Progress: 100,
	}
	ui.testResults = append(ui.testResults, result)

	// Update status counts
	switch status {
	case "passed":
		ui.status.PassedCount++
	case "failed":
		ui.status.FailedCount++
	case "skipped":
		ui.status.SkippedCount++
	}

	// Keep only the most recent test results (last 10)
	maxResults := 10
	if len(ui.testResults) > maxResults {
		ui.testResults = ui.testResults[len(ui.testResults)-maxResults:]
	}

	// Only render if we're in running state to show progress
	if ui.status.State == "running" {
		ui.renderFullScreen()
	}
}

// AddLiveOutput adds a line to the live output section
func (ui *UIController) AddLiveOutput(message string) {
	ui.liveOutput = append(ui.liveOutput, message)

	// Keep only the most recent output lines
	if len(ui.liveOutput) > ui.maxOutput {
		ui.liveOutput = ui.liveOutput[len(ui.liveOutput)-ui.maxOutput:]
	}

	ui.renderFullScreen()
}

// OnFileChange notifies the UI of a file change event
func (ui *UIController) OnFileChange(filePath string) {
	ui.status.FileChanges++
	ui.status.State = "running"

	// Add to live output instead of scrolling log
	fileName := filepath.Base(filePath)
	ui.AddLiveOutput(fmt.Sprintf("📁 File changed: %s", fileName))
}

// GetKeyInput returns the channel for keyboard input
func (ui *UIController) GetKeyInput() <-chan rune {
	return ui.keyHandler.inputChan
}

// Stop gracefully stops the UI controller
func (ui *UIController) Stop() {
	ui.keyHandler.enabled = false
	ui.isActive = false

	if ui.keyHandler.cancelFunc != nil {
		ui.keyHandler.cancelFunc()
	}

	// Restore terminal - show cursor and move to bottom
	fmt.Print("\033[?25h")
	ui.moveCursor(25, 1)
	fmt.Printf("\n🛑 %sDaemon stopped%s\n", colorGreen(), colorReset())
}

// refreshDisplay updates the entire UI display (legacy - replaced by renderFullScreen)
func (ui *UIController) refreshDisplay() {
	ui.renderFullScreen()
}

// clearScreen clears the terminal screen
func (ui *UIController) clearScreen() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		fmt.Print("\033[2J\033[H")
	}
}

// printHeader displays the testicle header
func (ui *UIController) printHeader() {
	ui.moveCursor(1, 1)
	fmt.Printf("🧪 %sTESTICLE DAEMON%s\n", colorBold(), colorReset())
	fmt.Printf("   %sPlaywright-inspired test runner for Go%s\n", colorDim(), colorReset())
	fmt.Printf("   %s%s%s\n", colorDim(), strings.Repeat("─", 50), colorReset())
	fmt.Println()
}

// printStatus displays the current daemon status
func (ui *UIController) printStatus() {
	ui.moveCursor(5, 1)

	// Status line
	stateColor := ui.getStateColor(ui.status.State)
	fmt.Printf("📊 Status: %s%s%s", stateColor, ui.getStateIcon(ui.status.State), colorReset())
	fmt.Printf(" %s%s%s\n", stateColor, strings.ToUpper(ui.status.State), colorReset())

	// Test results
	if ui.status.TestCount > 0 {
		fmt.Printf("🧪 Tests: %s%d total%s", colorDim(), ui.status.TestCount, colorReset())
		fmt.Printf(" • %s%d passed%s", colorGreen(), ui.status.PassedCount, colorReset())
		if ui.status.FailedCount > 0 {
			fmt.Printf(" • %s%d failed%s", colorRed(), ui.status.FailedCount, colorReset())
		}
		fmt.Printf(" • %s%s%s\n", colorDim(), ui.status.Duration, colorReset())
	} else {
		fmt.Printf("🧪 Tests: %sNo tests run yet%s\n", colorDim(), colorReset())
	}

	// File watching info
	fmt.Printf("👀 Watching: %s%s%s", colorCyan(), ui.runner.config.Dir, colorReset())
	if ui.status.FileChanges > 0 {
		fmt.Printf(" • %s%d changes detected%s", colorYellow(), ui.status.FileChanges, colorReset())
	}
	fmt.Println()

	// Last run time
	if !ui.status.LastRun.IsZero() {
		elapsed := time.Since(ui.status.LastRun)
		fmt.Printf("⏱️  Last run: %s%s ago%s\n", colorDim(), formatDuration(elapsed), colorReset())
	}

	fmt.Println()
}

// printControls displays the interactive controls
func (ui *UIController) printControls() {
	ui.moveCursor(12, 1)
	fmt.Printf("%s┌─ Interactive Controls ─────────────────────────┐%s\n", colorDim(), colorReset())
	fmt.Printf("%s│%s %s[r]%s Re-run tests now  %s[p]%s Pause/Resume  %s[c]%s Clear   %s│%s\n",
		colorDim(), colorReset(), colorGreen(), colorReset(), colorYellow(), colorReset(), colorBlue(), colorReset(), colorDim(), colorReset())
	fmt.Printf("%s│%s %s[d]%s Toggle debug      %s[s]%s Show stats   %s[q]%s Quit    %s│%s\n",
		colorDim(), colorReset(), colorMagenta(), colorReset(), colorCyan(), colorReset(), colorRed(), colorReset(), colorDim(), colorReset())
	fmt.Printf("%s└─────────────────────────────────────────────────┘%s\n", colorDim(), colorReset())
	fmt.Println()

	// Live output area
	fmt.Printf("%sLive Output:%s\n", colorBold(), colorReset())
	fmt.Printf("%s%s%s\n", colorDim(), strings.Repeat("─", 50), colorReset())
}

// handleKeyboardInput processes keyboard input for interactive controls
func (ui *UIController) handleKeyboardInput(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	ui.keyHandler.cancelFunc = cancel
	defer cancel()

	// Enable raw mode for immediate key detection
	if err := ui.enableRawMode(); err != nil {
		ui.logger.Debug("Failed to enable raw mode: %v", err)
		return
	}
	defer ui.disableRawMode()

	for ui.keyHandler.enabled {
		select {
		case <-ctx.Done():
			return
		default:
			// Read single character
			var buf [1]byte
			n, err := os.Stdin.Read(buf[:])
			if err != nil || n == 0 {
				time.Sleep(50 * time.Millisecond)
				continue
			}

			key := rune(buf[0])
			if ui.keyHandler.enabled {
				select {
				case ui.keyHandler.inputChan <- key:
				default:
					// Channel full, skip
				}
			}
		}
	}
}

// getStateColor returns the appropriate color for the current state
func (ui *UIController) getStateColor(state string) string {
	switch state {
	case "running":
		return colorYellow()
	case "failed":
		return colorRed()
	case "watching":
		return colorGreen()
	case "paused":
		return colorMagenta()
	default:
		return colorDim()
	}
}

// getStateIcon returns an icon for the current state
func (ui *UIController) getStateIcon(state string) string {
	switch state {
	case "running":
		return "🔄"
	case "failed":
		return "❌"
	case "watching":
		return "👀"
	case "paused":
		return "⏸️"
	default:
		return "⚪"
	}
}

// moveCursor moves the terminal cursor to the specified line and column
func (ui *UIController) moveCursor(line, col int) {
	fmt.Printf("\033[%d;%dH", line, col)
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}

// renderHeader displays the fixed header section
func (ui *UIController) renderHeader() {
	ui.moveCursor(1, 1)
	fmt.Printf("🧪 %sTESTICLE DAEMON%s", colorBold(), colorReset())

	// Clear to end of line and move to next
	fmt.Print("\033[K\n")

	// Subtitle line
	fmt.Printf("   %sPlaywright-inspired test runner for Go%s", colorDim(), colorReset())
	fmt.Print("\033[K\n")

	// Separator line
	fmt.Printf("   %s%s%s", colorDim(), strings.Repeat("─", 50), colorReset())
	fmt.Print("\033[K\n\n")
}

// renderStatus displays the current daemon status section
func (ui *UIController) renderStatus() {
	ui.moveCursor(6, 1)

	// Status line with state
	stateColor := ui.getStateColor(ui.status.State)
	stateIcon := ui.getStateIcon(ui.status.State)
	stateName := strings.ToUpper(ui.status.State)

	if ui.status.State == "running" && ui.status.TestCount > 0 {
		progress := ui.calculateProgress()
		fmt.Printf("🧪 %sTesticle - Running tests... (%d/%d completed)%s",
			colorBold(), ui.status.PassedCount+ui.status.FailedCount, ui.status.TestCount, colorReset())
		fmt.Print("\033[K\n\n")

		// Progress bar
		fmt.Printf("Progress: %s %d%% (%s elapsed)",
			ui.renderProgressBar(progress), progress, ui.status.ElapsedTime)
		fmt.Print("\033[K\n\n")
	} else {
		fmt.Printf("📊 Status: %s%s %s%s", stateColor, stateIcon, stateName, colorReset())
		fmt.Print("\033[K\n")

		// Test summary line
		if ui.status.TestCount > 0 {
			fmt.Printf("🧪 Tests: %s%d total%s", colorDim(), ui.status.TestCount, colorReset())
			fmt.Printf(" • %s%d passed%s", colorGreen(), ui.status.PassedCount, colorReset())
			if ui.status.FailedCount > 0 {
				fmt.Printf(" • %s%d failed%s", colorRed(), ui.status.FailedCount, colorReset())
			}
			if ui.status.SkippedCount > 0 {
				fmt.Printf(" • %s%d skipped%s", colorYellow(), ui.status.SkippedCount, colorReset())
			}
			if ui.status.Duration != "" {
				fmt.Printf(" • %s%s%s", colorDim(), ui.status.Duration, colorReset())
			}
		} else {
			fmt.Printf("🧪 Tests: %sNo tests run yet%s", colorDim(), colorReset())
		}
		fmt.Print("\033[K\n")

		// File watching info
		fmt.Printf("👀 Watching: %s%s%s", colorCyan(), ui.runner.config.Dir, colorReset())
		if ui.status.FileChanges > 0 {
			fmt.Printf(" • %s%d changes detected%s", colorYellow(), ui.status.FileChanges, colorReset())
		}
		fmt.Print("\033[K\n\n")
	}
}

// renderTestResults displays individual test results
func (ui *UIController) renderTestResults() {
	if ui.status.State != "running" || len(ui.testResults) == 0 {
		return
	}

	// Show last few test results in a compact format
	startLine := 11
	maxDisplay := 5 // Show last 5 test results

	recentResults := ui.testResults
	if len(recentResults) > maxDisplay {
		recentResults = recentResults[len(recentResults)-maxDisplay:]
	}

	for i, result := range recentResults {
		ui.moveCursor(startLine+i, 1)

		var statusIcon, statusColor string
		switch result.Status {
		case "passed":
			statusIcon = "✅"
			statusColor = colorGreen()
		case "failed":
			statusIcon = "❌"
			statusColor = colorRed()
		case "running":
			statusIcon = "🏃"
			statusColor = colorYellow()
		case "queued":
			statusIcon = "⏳"
			statusColor = colorDim()
		default:
			statusIcon = "⚪"
			statusColor = colorDim()
		}

		// Truncate long test names
		displayName := result.Name
		if len(displayName) > 40 {
			displayName = displayName[:37] + "..."
		}

		fmt.Printf("%s %s%-43s%s", statusIcon, statusColor, displayName, colorReset())
		if result.Duration != "" {
			fmt.Printf(" %s(%s)%s", colorDim(), result.Duration, colorReset())
		}
		fmt.Print("\033[K\n")
	}

	// Clear any remaining lines in this section
	for i := len(recentResults); i < maxDisplay; i++ {
		ui.moveCursor(startLine+i, 1)
		fmt.Print("\033[K")
	}
}

// renderLiveOutput displays the live output section
func (ui *UIController) renderLiveOutput() {
	ui.moveCursor(18, 1)

	if len(ui.liveOutput) > 0 {
		for i, line := range ui.liveOutput {
			ui.moveCursor(18+i, 1)
			fmt.Printf("%s", line)
			fmt.Print("\033[K\n")
		}
	}

	// Clear any remaining lines in live output section
	for i := len(ui.liveOutput); i < ui.maxOutput; i++ {
		ui.moveCursor(18+i, 1)
		fmt.Print("\033[K")
	}
}

// renderControls displays the interactive controls panel
func (ui *UIController) renderControls() {
	ui.moveCursor(24, 1)

	fmt.Printf("%s[r] Run Now | [s] Stop | [p] Pause | [c] Clear | [q] Quit%s",
		colorDim(), colorReset())
	fmt.Print("\033[K")
}

// calculateProgress calculates the current test progress percentage
func (ui *UIController) calculateProgress() int {
	if ui.status.TestCount == 0 {
		return 0
	}
	completed := ui.status.PassedCount + ui.status.FailedCount + ui.status.SkippedCount
	return (completed * 100) / ui.status.TestCount
}

// renderProgressBar renders a progress bar for the given percentage
func (ui *UIController) renderProgressBar(percent int) string {
	// Calculate progress bar width based on terminal width
	// Leave space for "Progress: " (10 chars) + " XX% (Xs elapsed)" (~15 chars) + margins (~5 chars)
	reservedSpace := 30
	width := ui.screenWidth - reservedSpace
	if width < 10 {
		width = 10 // Minimum width
	}
	if width > 50 {
		width = 50 // Maximum width for readability
	}

	filled := int((float64(percent)*float64(width))/100.0 + 0.5) // Add 0.5 for proper rounding
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("%s%s%s", colorGreen(), bar, colorReset())
}

// renderMiniProgressBar renders a smaller progress bar for individual tests
func (ui *UIController) renderMiniProgressBar(percent int) string {
	width := 10
	filled := int((float64(percent)*float64(width))/100.0 + 0.5) // Add 0.5 for proper rounding
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s%s%s] %d%%", colorYellow(), bar, colorReset(), percent)
}

// Color functions for terminal output
func colorReset() string   { return "\033[0m" }
func colorBold() string    { return "\033[1m" }
func colorDim() string     { return "\033[2m" }
func colorRed() string     { return "\033[31m" }
func colorGreen() string   { return "\033[32m" }
func colorYellow() string  { return "\033[33m" }
func colorBlue() string    { return "\033[34m" }
func colorMagenta() string { return "\033[35m" }
func colorCyan() string    { return "\033[36m" }

// Platform-specific raw mode functions
func (ui *UIController) enableRawMode() error {
	if runtime.GOOS == "windows" {
		// Windows raw mode - use term package
		fd := int(os.Stdin.Fd())
		if !term.IsTerminal(fd) {
			return fmt.Errorf("stdin is not a terminal")
		}

		state, err := term.MakeRaw(fd)
		if err != nil {
			return err
		}
		ui.keyHandler.originalTerm = state
		return nil
	}

	// Unix-like systems (Linux, macOS, etc.) - use term package
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return fmt.Errorf("stdin is not a terminal")
	}

	state, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	ui.keyHandler.originalTerm = state
	return nil
}

func (ui *UIController) disableRawMode() {
	if ui.keyHandler.originalTerm != nil {
		fd := int(os.Stdin.Fd())
		term.Restore(fd, ui.keyHandler.originalTerm)
	}
}
