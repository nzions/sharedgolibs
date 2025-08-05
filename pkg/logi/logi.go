package logi

// Version of the logi package
const Version = "0.2.0"

// Logger defines the interface for logging in the Logi package.
type Logger interface {
	Info(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	Debug(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
}
