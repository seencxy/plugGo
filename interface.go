package plugGo

type Application interface {
	// Start starts the application.
	Start() error
	// Stop stops the application.
	Stop() error
	// GetLogger returns the application's logger.
	GetLogger() Logger
	// SetLogger sets the application's logger.
	SetLogger(logger Logger)
}

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Trace(args ...interface{})
	Warn(args ...interface{})
}
