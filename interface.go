package plugGo

import "context"

// PluginStatus represents the status of a plugin.
type PluginStatus int

const (
	// StatusIdle indicates the plugin has not been started yet.
	StatusIdle PluginStatus = iota
	// StatusRunning indicates the plugin is currently running.
	StatusRunning
	// StatusStopped indicates the plugin has been stopped.
	StatusStopped
	// StatusError indicates the plugin encountered an error.
	StatusError
)

// String returns the string representation of PluginStatus.
func (s PluginStatus) String() string {
	switch s {
	case StatusIdle:
		return "Idle"
	case StatusRunning:
		return "Running"
	case StatusStopped:
		return "Stopped"
	case StatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

// StatusEvent represents a plugin status change event.
type StatusEvent struct {
	// Status is the current plugin status.
	Status PluginStatus
	// Error contains error information if Status is StatusError, nil otherwise.
	Error error
}

type Application interface {
	// Start starts the application with context for timeout and cancellation control.
	Start(ctx context.Context) error
	// Stop stops the application with context for graceful shutdown timeout control.
	Stop(ctx context.Context) error
	// GetLogger returns the application's logger.
	GetLogger() Logger
	// SetLogger sets the application's logger.
	SetLogger(logger Logger)
	// Status returns the current status of the application.
	Status() PluginStatus
	// StatusNotify returns a read-only channel for receiving status change events.
	// The channel will receive StatusEvent when the plugin status changes.
	// Note: Subscribers should not block on this channel to avoid missing events.
	StatusNotify() <-chan StatusEvent
}

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Trace(args ...interface{})
	Warn(args ...interface{})
}
