// Package log provides a log interface
package logger

var (
	// DefaultLogger logger
	DefaultLogger Logger
)

// Logger is a generic logging interface
type Logger interface {
	// Init initialises options
	Init(options ...Option) error
	// Options The Logger options
	Options() Options
	// Fields set fields to always be logged
	Fields(fields map[string]interface{}) Logger
	// Log writes a log entry
	Log(level Level, v ...interface{})
	// Logf writes a formatted log entry
	Logf(level Level, format string, v ...interface{})
	// String returns the name of logger
	String() string
}

// Init initialises options
func Init(opts ...Option) error {
	return DefaultLogger.Init(opts...)
}

// Fields set fields to always be logged
func Fields(fields map[string]interface{}) Logger {
	return DefaultLogger.Fields(fields)
}

// Log writes a log entry
func Log(level Level, v ...interface{}) {
	DefaultLogger.Log(level, v...)
}

// Logf writes a formatted log entry
func Logf(level Level, format string, v ...interface{}) {
	DefaultLogger.Logf(level, format, v...)
}

// String returns the name of logger
func String() string {
	return DefaultLogger.String()
}
