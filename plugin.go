package plugGo

// Plugin defines the standard interface for plugins.
// Plugin instances focus on business logic, config management is handled by framework and factory.
type Plugin interface {
	Application // Inherits Start, Stop, GetLogger, SetLogger methods

	// ID returns the unique identifier of the plugin instance.
	// Same plugin type can have multiple instances, distinguished by ID.
	ID() string

	// PluginType returns the plugin type name (e.g. "announcement").
	PluginType() string

	// Version returns the plugin version.
	Version() string

	// Reload reloads the plugin with new config.
	// Parameters:
	//   - config: new config object (already validated)
	// Returns:
	//   - error: returns error if reload fails
	// Note: This method may restart internal components to apply new config.
	Reload(config interface{}) error
}
