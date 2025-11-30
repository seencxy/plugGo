package plugGo

// PluginFactory is the plugin factory interface.
// Responsible for creating plugin instances, providing config templates and validating config.
// Each plugin needs to implement a factory to support multi-instance creation.
type PluginFactory interface {
	// Plugin metadata
	Name() string    // Returns plugin type name (e.g. "announcement")
	Version() string // Returns plugin version

	// Config management (handled by factory, not plugin instance)
	DefaultConfig() interface{}              // Returns default config struct
	ValidateConfig(config interface{}) error // Validates if config is valid

	// Instance creation
	// Parameters:
	//   - instanceID: unique identifier for the instance
	//   - config: config for this instance (already validated)
	//   - logger: logger injected for this instance
	// Returns:
	//   - Plugin: created plugin instance
	//   - error: returns error if creation fails
	Create(instanceID string, config interface{}, logger Logger) (Plugin, error)
}
