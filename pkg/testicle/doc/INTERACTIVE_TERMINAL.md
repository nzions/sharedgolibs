# Testicle Interactive Terminal Interface

## 🎮 Interactive Key Handlers

### Live Terminal Controls

When running in daemon mode (`--daemon`) or during test execution, Testicle provides interactive keyboard controls for real-time test management:

```
┌─────────────────────────────────────────────────────────────────┐
│ 🧪 Testicle Daemon v1.0 - Watching /tests for changes...       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│ ✅ TestUserValidation          (89ms)                          │
│ ✅ TestUserSerialization      (142ms)                          │ 
│ 🏃 TestUserAuth               [████████░░] 80%                 │
│ ⏳ TestUserPermissions        (queued)                         │
│                                                                 │
├─────────────────────────────────────────────────────────────────┤
│ [r] Run Now | [s] Stop | [p] Pause | [ESC] Resume | [q] Quit   │
└─────────────────────────────────────────────────────────────────┘
```

### Key Bindings

| Key   | Action              | Description                                      |
| ----- | ------------------- | ------------------------------------------------ |
| `r`   | **Run Tests Now**   | Immediately trigger test discovery and execution |
| `s`   | **Stop Tests**      | Stop currently running tests gracefully          |
| `p`   | **Pause Watching**  | Pause file watching (daemon mode only)           |
| `ESC` | **Resume Watching** | Resume file watching when paused                 |
| `q`   | **Quit**            | Exit Testicle gracefully                         |
| `d`   | **Toggle Debug**    | Enable/disable debug output on the fly           |
| `v`   | **Toggle Verbose**  | Switch between normal and verbose test output    |
| `c`   | **Clear Screen**    | Clear the terminal and refresh display           |
| `h`   | **Help**            | Show key bindings help                           |

### Interactive States

#### 1. Daemon Mode - Watching
```
🧪 Testicle Daemon - Watching /tests for changes...

Status: 👁️  Watching (47 tests discovered)
Last run: 2m 14s ago (42 passed, 3 failed)

Ready - watching for changes...
[r] Run Now | [p] Pause | [q] Quit | [h] Help
```

#### 2. Daemon Mode - Paused
```
🧪 Testicle Daemon - Paused

Status: ⏸️  Paused (watching disabled)
Last run: 2m 14s ago (42 passed, 3 failed)

File watching paused - press ESC to resume
[ESC] Resume | [r] Run Now | [q] Quit | [h] Help
```

#### 3. Running Tests
```
🧪 Testicle - Running tests... (23/47 completed)

Progress: ████████████████░░░░░░░░░░░░ 49% (1m 12s elapsed)

✅ pkg/auth/TestUserLogin          (150ms)
🏃 pkg/api/TestCreateUser          [████████░░] 80%
⏳ pkg/db/TestConnection           (queued)

[s] Stop | [v] Verbose | [c] Clear | [q] Quit
```

#### 4. Tests Stopped
```
🧪 Testicle - Tests stopped by user

Status: ⏹️  Stopped (23/47 tests completed)
Summary: 15 passed, 2 failed, 6 skipped, 24 not run

Tests were stopped gracefully
[r] Run Again | [d] Daemon Mode | [q] Quit | [h] Help
```

## 🔧 Implementation Details

### Terminal Input Handling

```go
// Terminal input handler for interactive controls
type TerminalController struct {
    inputChan   chan rune
    stateChan   chan ControllerState
    app         *App
    keyBindings map[rune]KeyHandler
}

type ControllerState struct {
    Mode        ControllerMode `json:"mode"`
    TestsRunning bool          `json:"tests_running"`
    WatchPaused  bool          `json:"watch_paused"`
    DebugMode    bool          `json:"debug_mode"`
    VerboseMode  bool          `json:"verbose_mode"`
}

type ControllerMode string

const (
    ModeWatching    ControllerMode = "watching"
    ModePaused      ControllerMode = "paused"
    ModeRunning     ControllerMode = "running"
    ModeStopped     ControllerMode = "stopped"
)

type KeyHandler func(*TerminalController, ControllerState) error

func (tc *TerminalController) RegisterKeyBindings() {
    tc.keyBindings = map[rune]KeyHandler{
        'r': tc.handleRunNow,
        's': tc.handleStop,
        'p': tc.handlePause,
        27:  tc.handleResume, // ESC key
        'q': tc.handleQuit,
        'd': tc.handleToggleDebug,
        'v': tc.handleToggleVerbose,
        'c': tc.handleClearScreen,
        'h': tc.handleHelp,
    }
}

func (tc *TerminalController) handleRunNow(state ControllerState) error {
    if state.TestsRunning {
        return fmt.Errorf("tests already running")
    }
    
    tc.app.TriggerTestRun()
    tc.updateStatus("🚀 Running tests now...")
    return nil
}

func (tc *TerminalController) handleStop(state ControllerState) error {
    if !state.TestsRunning {
        return fmt.Errorf("no tests running")
    }
    
    tc.app.StopTestExecution()
    tc.updateStatus("⏹️  Stopping tests...")
    return nil
}

func (tc *TerminalController) handlePause(state ControllerState) error {
    if state.WatchPaused {
        return fmt.Errorf("watching already paused")
    }
    
    tc.app.PauseWatching()
    tc.updateStatus("⏸️  File watching paused")
    return nil
}

func (tc *TerminalController) handleResume(state ControllerState) error {
    if !state.WatchPaused {
        return fmt.Errorf("watching not paused")
    }
    
    tc.app.ResumeWatching()
    tc.updateStatus("👁️  File watching resumed")
    return nil
}
```

### Raw Terminal Mode Setup

```go
// Terminal mode management for capturing key presses
type TerminalMode struct {
    originalState *term.State
    rawMode       bool
}

func (tm *TerminalMode) EnableRawMode() error {
    if tm.rawMode {
        return nil
    }
    
    state, err := term.MakeRaw(int(os.Stdin.Fd()))
    if err != nil {
        return err
    }
    
    tm.originalState = state
    tm.rawMode = true
    return nil
}

func (tm *TerminalMode) DisableRawMode() error {
    if !tm.rawMode {
        return nil
    }
    
    err := term.Restore(int(os.Stdin.Fd()), tm.originalState)
    tm.rawMode = false
    return err
}

func (tm *TerminalMode) ReadKey() (rune, error) {
    reader := bufio.NewReader(os.Stdin)
    char, _, err := reader.ReadRune()
    return char, err
}
```

### Status Bar and Help Display

```go
// Dynamic status bar showing available commands
type StatusBar struct {
    width   int
    state   ControllerState
    visible bool
}

func (sb *StatusBar) Render(state ControllerState) string {
    var commands []string
    
    switch state.Mode {
    case ModeWatching:
        if state.TestsRunning {
            commands = []string{"[s] Stop", "[v] Verbose", "[c] Clear", "[q] Quit"}
        } else {
            commands = []string{"[r] Run Now", "[p] Pause", "[q] Quit", "[h] Help"}
        }
    case ModePaused:
        commands = []string{"[ESC] Resume", "[r] Run Now", "[q] Quit", "[h] Help"}
    case ModeRunning:
        commands = []string{"[s] Stop", "[v] Verbose", "[d] Debug", "[q] Quit"}
    case ModeStopped:
        commands = []string{"[r] Run Again", "[d] Daemon", "[q] Quit", "[h] Help"}
    }
    
    commandStr := strings.Join(commands, " | ")
    
    // Center the commands in the terminal width
    padding := (sb.width - len(commandStr)) / 2
    if padding < 0 {
        padding = 0
    }
    
    return fmt.Sprintf("%s%s%s",
        strings.Repeat(" ", padding),
        commandStr,
        strings.Repeat(" ", padding),
    )
}

func (sb *StatusBar) ShowHelp() string {
    help := []string{
        "🎮 Testicle Interactive Controls",
        "",
        "  r  - Run tests now (trigger immediate execution)",
        "  s  - Stop currently running tests",
        "  p  - Pause file watching (daemon mode)",
        " ESC - Resume file watching when paused",
        "  q  - Quit Testicle",
        "  d  - Toggle debug output",
        "  v  - Toggle verbose test output",
        "  c  - Clear screen and refresh display",
        "  h  - Show this help",
        "",
        "Press any key to continue...",
    }
    
    return strings.Join(help, "\n")
}
```

### Integration with Main Application

```go
// Enhanced App struct with interactive controls
type App struct {
    // ... existing fields ...
    
    controller    *TerminalController
    terminalMode  *TerminalMode
    interactive   bool
    controlChan   chan ControlCommand
}

type ControlCommand struct {
    Type ControlCommandType `json:"type"`
    Data interface{}        `json:"data"`
}

type ControlCommandType string

const (
    CommandRunNow        ControlCommandType = "run_now"
    CommandStop          ControlCommandType = "stop"
    CommandPause         ControlCommandType = "pause"
    CommandResume        ControlCommandType = "resume"
    CommandToggleDebug   ControlCommandType = "toggle_debug"
    CommandToggleVerbose ControlCommandType = "toggle_verbose"
    CommandClearScreen   ControlCommandType = "clear_screen"
    CommandQuit          ControlCommandType = "quit"
)

func (a *App) EnableInteractiveMode() error {
    if !isatty.IsTerminal(os.Stdin.Fd()) {
        return fmt.Errorf("not running in a terminal")
    }
    
    a.interactive = true
    a.controlChan = make(chan ControlCommand, 10)
    
    // Set up terminal controller
    a.controller = NewTerminalController(a)
    a.terminalMode = &TerminalMode{}
    
    if err := a.terminalMode.EnableRawMode(); err != nil {
        return err
    }
    
    // Start input handler goroutine
    go a.inputHandler()
    
    return nil
}

func (a *App) inputHandler() {
    defer a.terminalMode.DisableRawMode()
    
    for {
        key, err := a.terminalMode.ReadKey()
        if err != nil {
            continue
        }
        
        if handler, exists := a.controller.keyBindings[key]; exists {
            state := a.getControllerState()
            if err := handler(state); err != nil {
                a.showError(err.Error())
            }
        }
    }
}
```

## 📱 User Experience Flow

### 1. Starting Daemon Mode
```bash
$ testicle --daemon
🧪 Testicle Daemon v1.0 - Watching /tests for changes...

Discovering tests... ████████████████████████████████ 100%
Found 47 tests in 12 packages

Ready - watching for changes...
[r] Run Now | [p] Pause | [q] Quit | [h] Help
```

### 2. User Presses 'r' (Run Now)
```
🚀 Running tests now...

Running tests... ████████████████████████████░░░░░ 89% (42/47)

✅ pkg/auth/TestUserLogin          (150ms)
✅ pkg/auth/TestPasswordValidation (89ms)
🏃 pkg/api/TestCreateUser          [████████░░] 80%

[s] Stop | [v] Verbose | [c] Clear | [q] Quit
```

### 3. User Presses 's' (Stop)
```
⏹️  Stopping tests...

Tests stopped by user (28/47 completed)
Summary: 25 passed, 2 failed, 1 skipped, 19 not run

[r] Run Again | [d] Daemon Mode | [q] Quit | [h] Help
```

### 4. User Presses 'p' (Pause Watching)
```
⏸️  File watching paused

Status: Paused (watching disabled)
Last run: 30s ago (25 passed, 2 failed)

File watching paused - press ESC to resume
[ESC] Resume | [r] Run Now | [q] Quit | [h] Help
```

This interactive terminal interface provides intuitive, real-time control over test execution, making Testicle feel responsive and developer-friendly like modern interactive tools.
