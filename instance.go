package plugGo

import (
	"fmt"
	"sync"
)

// PluginInstance is the plugin instance wrapper.
// Encapsulates plugin instance, config and metadata.
type PluginInstance struct {
	id         string        // Instance unique identifier
	pluginType string        // Plugin type name
	plugin     Plugin        // Plugin instance
	config     interface{}   // Current config
	factory    PluginFactory // Factory that created this instance
	mu         sync.RWMutex  // Protects concurrent access
}

// NewPluginInstance creates a new plugin instance wrapper.
func NewPluginInstance(id string, pluginType string, plugin Plugin, config interface{}, factory PluginFactory) *PluginInstance {
	return &PluginInstance{
		id:         id,
		pluginType: pluginType,
		plugin:     plugin,
		config:     config,
		factory:    factory,
	}
}

// ID returns the instance ID.
func (pi *PluginInstance) ID() string {
	return pi.id
}

// PluginType returns the plugin type.
func (pi *PluginInstance) PluginType() string {
	return pi.pluginType
}

// Plugin returns the plugin instance.
func (pi *PluginInstance) Plugin() Plugin {
	return pi.plugin
}

// GetConfig returns current config (returns reference, caller handles concurrency).
func (pi *PluginInstance) GetConfig() interface{} {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	return pi.config
}

// UpdateConfig updates config and reloads the plugin.
func (pi *PluginInstance) UpdateConfig(newConfig interface{}) error {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	// Validate new config
	if err := pi.factory.ValidateConfig(newConfig); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Save old config for rollback
	oldConfig := pi.config

	// Update config
	pi.config = newConfig

	// Reload plugin
	if err := pi.plugin.Reload(newConfig); err != nil {
		// Rollback config
		pi.config = oldConfig
		return fmt.Errorf("failed to reload plugin: %w", err)
	}

	return nil
}

// Start starts the plugin.
func (pi *PluginInstance) Start() error {
	return pi.plugin.Start()
}

// Stop stops the plugin.
func (pi *PluginInstance) Stop() error {
	return pi.plugin.Stop()
}

// GetLogger returns the logger.
func (pi *PluginInstance) GetLogger() Logger {
	return pi.plugin.GetLogger()
}

// SetLogger sets the logger.
func (pi *PluginInstance) SetLogger(logger Logger) {
	pi.plugin.SetLogger(logger)
}
