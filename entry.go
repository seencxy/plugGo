package plugGo

import (
	"context"
	"time"
)

// Entry is the base interface for all bootable components.
// Inspired by rk-boot design, Entry represents a service/component that can be bootstrapped from YAML config.
type Entry interface {
	// Bootstrap starts the Entry.
	// ctx is used to pass event ID and other context information.
	Bootstrap(ctx context.Context)

	// Interrupt stops the Entry.
	// Waits for shutdown signal and completes pending operations.
	Interrupt(ctx context.Context)

	// GetName returns the name of the Entry (instance name).
	GetName() string

	// GetType returns the type of the Entry (plugin type).
	GetType() string

	// GetDescription returns the description of the Entry.
	GetDescription() string

	// String returns the string representation of the Entry.
	String() string
}

// RegFunc is the registration function type for Entry.
// Creates Entry instances from raw YAML config.
// Returns map[name]Entry, supporting multiple instances of the same type.
type RegFunc func(raw []byte) map[string]Entry

// EntryType defines Entry type constants.
const (
	// PluginEntryType is the plugin type.
	PluginEntryType = "PluginEntry"
	// UserEntryType is the user-defined type.
	UserEntryType = "UserEntry"
)

// ShutdownHook is the shutdown hook function type.
type ShutdownHook func()

// BootConfig holds bootstrap configuration options.
type BootConfig struct {
	// ConfigPath is the config file path, defaults to boot.yaml.
	ConfigPath string
	// ConfigRaw is the raw config content (takes precedence over file).
	ConfigRaw []byte
	// ShutdownTimeout is the overall shutdown timeout, defaults to 30s.
	ShutdownTimeout time.Duration
	// EntryShutdownTimeout is the timeout for shutting down a single Entry, defaults to 10s.
	EntryShutdownTimeout time.Duration
}

// BootOption is a bootstrap configuration option function.
type BootOption func(*BootConfig)

// WithConfigPath sets the config file path.
func WithConfigPath(path string) BootOption {
	return func(c *BootConfig) {
		c.ConfigPath = path
	}
}

// WithConfigRaw sets the raw config content.
func WithConfigRaw(raw []byte) BootOption {
	return func(c *BootConfig) {
		c.ConfigRaw = raw
	}
}

// WithShutdownTimeout sets the overall shutdown timeout.
func WithShutdownTimeout(timeout time.Duration) BootOption {
	return func(c *BootConfig) {
		c.ShutdownTimeout = timeout
	}
}

// WithEntryShutdownTimeout sets the timeout for shutting down a single Entry.
func WithEntryShutdownTimeout(timeout time.Duration) BootOption {
	return func(c *BootConfig) {
		c.EntryShutdownTimeout = timeout
	}
}
