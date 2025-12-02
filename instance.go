package plugGo

import (
	"context"
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

// Start starts the plugin with context.
func (pi *PluginInstance) Start(ctx context.Context) error {
	return pi.plugin.Start(ctx)
}

// Stop stops the plugin with context.
func (pi *PluginInstance) Stop(ctx context.Context) error {
	return pi.plugin.Stop(ctx)
}

// GetLogger returns the logger.
func (pi *PluginInstance) GetLogger() Logger {
	return pi.plugin.GetLogger()
}

// SetLogger sets the logger.
func (pi *PluginInstance) SetLogger(logger Logger) {
	pi.plugin.SetLogger(logger)
}

// Status returns the current plugin status.
func (pi *PluginInstance) Status() PluginStatus {
	return pi.plugin.Status()
}

// StatusNotify returns a read-only channel for receiving status change events.
func (pi *PluginInstance) StatusNotify() <-chan StatusEvent {
	return pi.plugin.StatusNotify()
}

// GetNotifyChannel returns the plugin's notification channel.
// This delegates to the plugin's GetNotifyChannel() method.
// Returns nil if the plugin doesn't support external notifications.
func (pi *PluginInstance) GetNotifyChannel() chan any {
	return pi.plugin.GetNotifyChannel()
}
