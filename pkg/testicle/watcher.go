package testicle

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileEvent represents a file system event
type FileEvent struct {
	Path string
	Op   string
}

// Watcher handles file system watching for daemon mode
type Watcher struct {
	dir     string
	logger  *Logger
	watcher *fsnotify.Watcher
}

// NewWatcher creates a new file watcher
func NewWatcher(dir string, logger *Logger) *Watcher {
	return &Watcher{
		dir:    dir,
		logger: logger,
	}
}

// Start begins watching for file changes
func (w *Watcher) Start(ctx context.Context) (<-chan FileEvent, error) {
	var err error
	w.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Walk the directory tree and add all directories to the watcher
	err = filepath.Walk(w.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and files
		if strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip vendor directories
		if strings.Contains(path, "/vendor/") || strings.Contains(path, "\\vendor\\") {
			return filepath.SkipDir
		}

		// Add directories to watcher
		if info.IsDir() {
			w.logger.Debug("👀 Watching directory: %s", path)
			return w.watcher.Add(path)
		}

		return nil
	})

	if err != nil {
		w.watcher.Close()
		return nil, err
	}

	eventChan := make(chan FileEvent, 10)

	// Start the event processing goroutine
	go w.processEvents(ctx, eventChan)

	w.logger.Debug("👀 File watcher started for %s", w.dir)
	return eventChan, nil
}

// processEvents processes file system events and filters relevant ones
func (w *Watcher) processEvents(ctx context.Context, eventChan chan<- FileEvent) {
	defer close(eventChan)
	defer w.watcher.Close()

	// Debouncing mechanism
	eventMap := make(map[string]time.Time)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Debug("👀 Stopping file watcher...")
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Filter relevant file events
			if w.isRelevantFile(event.Name) {
				w.logger.Debug("📁 File event: %s %s", event.Op, event.Name)
				eventMap[event.Name] = time.Now()
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			w.logger.Error("File watcher error: %v", err)

		case <-ticker.C:
			// Process debounced events
			now := time.Now()
			for path, eventTime := range eventMap {
				if now.Sub(eventTime) > 100*time.Millisecond {
					select {
					case eventChan <- FileEvent{Path: path, Op: "modified"}:
					case <-ctx.Done():
						return
					}
					delete(eventMap, path)
				}
			}
		}
	}
}

// isRelevantFile checks if a file change should trigger test re-run
func (w *Watcher) isRelevantFile(filename string) bool {
	// Only watch Go files
	if !strings.HasSuffix(filename, ".go") {
		return false
	}

	// Skip hidden files
	if strings.HasPrefix(filepath.Base(filename), ".") {
		return false
	}

	// Skip vendor directories
	if strings.Contains(filename, "/vendor/") || strings.Contains(filename, "\\vendor\\") {
		return false
	}

	// Skip generated files (common patterns)
	if strings.Contains(filename, ".gen.go") ||
		strings.Contains(filename, "_gen.go") ||
		strings.Contains(filename, ".pb.go") {
		return false
	}

	return true
}

// Stop stops the file watcher
func (w *Watcher) Stop() {
	if w.watcher != nil {
		w.watcher.Close()
	}
}
