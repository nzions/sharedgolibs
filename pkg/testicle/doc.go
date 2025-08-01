// Package testicle provides a Playwright-inspired test runner for Go that enhances
// the standard Go testing experience with interactive UI mode, watch mode,
// real-time status updates, and rich reporting capabilities.
//
// Testicle wraps the standard `go test` command while providing additional features:
//   - Interactive web-based UI for test execution and monitoring
//   - Watch mode that automatically re-runs tests when files change
//   - Timeline view showing test execution flow and timing
//   - Rich HTML reports with detailed test traces
//   - Intelligent test discovery and organization
//   - Parallel execution with configurable workers
//
// Example usage:
//
//	runner := testicle.NewRunner(testicle.Config{
//	    ProjectRoot: "/path/to/project",
//	    UIMode:      true,
//	    WatchMode:   true,
//	})
//	runner.Start()
package testicle
