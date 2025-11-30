package plugGo

import (
	"fmt"
	"log"
	"os"
	"time"
)

// LogLevel defines the log level.
type LogLevel int

const (
	// TraceLevel is the trace level.
	TraceLevel LogLevel = iota
	// DebugLevel is the debug level.
	DebugLevel
	// InfoLevel is the info level.
	InfoLevel
	// WarnLevel is the warn level.
	WarnLevel
	// ErrorLevel is the error level.
	ErrorLevel
)

// String returns the string representation of the log level.
func (l LogLevel) String() string {
	switch l {
	case TraceLevel:
		return "TRACE"
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// StandardLogger is the standard logger implementation.
type StandardLogger struct {
	level  LogLevel
	logger *log.Logger
	prefix string
}

// NewStandardLogger creates a new StandardLogger instance.
func NewStandardLogger(prefix string, level LogLevel) *StandardLogger {
	return &StandardLogger{
		level:  level,
		logger: log.New(os.Stdout, "", 0),
		prefix: prefix,
	}
}

// NewDefaultLogger creates a default Logger instance (INFO level).
func NewDefaultLogger(prefix string) *StandardLogger {
	return NewStandardLogger(prefix, InfoLevel)
}

// formatMessage formats the log message.
func (l *StandardLogger) formatMessage(level LogLevel, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prefix := ""
	if l.prefix != "" {
		prefix = fmt.Sprintf("[%s] ", l.prefix)
	}
	return fmt.Sprintf("%s [%s] %s%s", timestamp, level.String(), prefix, fmt.Sprint(args...))
}

// shouldLog determines whether to output the log.
func (l *StandardLogger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

// SetLevel sets the log level.
func (l *StandardLogger) SetLevel(level LogLevel) {
	l.level = level
}

// Trace outputs trace level log.
func (l *StandardLogger) Trace(args ...interface{}) {
	if l.shouldLog(TraceLevel) {
		l.logger.Println(l.formatMessage(TraceLevel, args...))
	}
}

// Debug outputs debug level log.
func (l *StandardLogger) Debug(args ...interface{}) {
	if l.shouldLog(DebugLevel) {
		l.logger.Println(l.formatMessage(DebugLevel, args...))
	}
}

// Info outputs info level log.
func (l *StandardLogger) Info(args ...interface{}) {
	if l.shouldLog(InfoLevel) {
		l.logger.Println(l.formatMessage(InfoLevel, args...))
	}
}

// Warn outputs warn level log.
func (l *StandardLogger) Warn(args ...interface{}) {
	if l.shouldLog(WarnLevel) {
		l.logger.Println(l.formatMessage(WarnLevel, args...))
	}
}

// Error outputs error level log.
func (l *StandardLogger) Error(args ...interface{}) {
	if l.shouldLog(ErrorLevel) {
		l.logger.Println(l.formatMessage(ErrorLevel, args...))
	}
}
