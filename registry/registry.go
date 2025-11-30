package registry

import (
	"fmt"
	"sync"

	"plugGo"
)

// Registry is the plugin registry.
// Manages plugin factories and plugin instances, supports multi-instance creation.
type Registry struct {
	factories map[string]plugGo.PluginFactory   // key: plugin type name
	instances map[string]*plugGo.PluginInstance // key: instance ID
	mu        sync.RWMutex
}

// defaultRegistry is the default global registry instance.
var defaultRegistry = &Registry{
	factories: make(map[string]plugGo.PluginFactory),
	instances: make(map[string]*plugGo.PluginInstance),
}

// RegisterFactory registers a plugin factory.
// Plugin factories typically call this method in their init() function for auto-registration.
//
// Parameters:
//   - factory: the plugin factory to register
//
// Notes:
//   - If plugin type name already exists, new factory will override old factory
//   - This method is thread-safe
func RegisterFactory(factory plugGo.PluginFactory) {
	defaultRegistry.mu.Lock()
	defer defaultRegistry.mu.Unlock()
	defaultRegistry.factories[factory.Name()] = factory
}

// GetFactory returns the factory by plugin type name.
//
// Parameters:
//   - pluginType: plugin type name
//
// Returns:
//   - plugGo.PluginFactory: the found plugin factory
//   - bool: whether the factory was found
func GetFactory(pluginType string) (plugGo.PluginFactory, bool) {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()
	factory, ok := defaultRegistry.factories[pluginType]
	return factory, ok
}

// GetAllFactories returns all registered plugin factories.
//
// Returns:
//   - map[string]plugGo.PluginFactory: mapping from plugin type name to factory
func GetAllFactories() map[string]plugGo.PluginFactory {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()

	result := make(map[string]plugGo.PluginFactory, len(defaultRegistry.factories))
	for name, factory := range defaultRegistry.factories {
		result[name] = factory
	}
	return result
}

// CreateInstance creates a plugin instance.
//
// Parameters:
//   - pluginType: plugin type name
//   - instanceID: unique identifier for the instance
//   - config: plugin config (if nil, uses default config)
//   - logger: logger (if nil, uses default logger)
//
// Returns:
//   - *plugGo.PluginInstance: created plugin instance
//   - error: returns error if creation fails
func CreateInstance(pluginType, instanceID string, config interface{}, logger plugGo.Logger) (*plugGo.PluginInstance, error) {
	defaultRegistry.mu.Lock()
	defer defaultRegistry.mu.Unlock()

	// Check if instance ID already exists
	if _, exists := defaultRegistry.instances[instanceID]; exists {
		return nil, fmt.Errorf("instance ID already exists: %s", instanceID)
	}

	// Get plugin factory
	factory, ok := defaultRegistry.factories[pluginType]
	if !ok {
		return nil, fmt.Errorf("plugin factory not found: %s", pluginType)
	}

	// Use default config (if not provided)
	if config == nil {
		config = factory.DefaultConfig()
	}

	// Validate config
	if err := factory.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Use default Logger (if not provided)
	if logger == nil {
		logger = plugGo.NewDefaultLogger(fmt.Sprintf("%s-%s", pluginType, instanceID))
	}

	// Create plugin instance
	plugin, err := factory.Create(instanceID, config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin instance: %w", err)
	}

	// Wrap as PluginInstance
	instance := plugGo.NewPluginInstance(instanceID, pluginType, plugin, config, factory)

	// Register instance
	defaultRegistry.instances[instanceID] = instance

	return instance, nil
}

// GetInstance returns the plugin instance by instance ID.
//
// Parameters:
//   - instanceID: unique identifier of the instance
//
// Returns:
//   - *plugGo.PluginInstance: the found plugin instance
//   - bool: whether the instance was found
func GetInstance(instanceID string) (*plugGo.PluginInstance, bool) {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()
	instance, ok := defaultRegistry.instances[instanceID]
	return instance, ok
}

// GetInstancesByType returns all plugin instances of the specified type.
//
// Parameters:
//   - pluginType: plugin type name
//
// Returns:
//   - []*plugGo.PluginInstance: all instances of this type
func GetInstancesByType(pluginType string) []*plugGo.PluginInstance {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()

	var result []*plugGo.PluginInstance
	for _, instance := range defaultRegistry.instances {
		if instance.PluginType() == pluginType {
			result = append(result, instance)
		}
	}
	return result
}

// GetAllInstances returns all plugin instances.
//
// Returns:
//   - []*plugGo.PluginInstance: slice of all plugin instances
func GetAllInstances() []*plugGo.PluginInstance {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()

	result := make([]*plugGo.PluginInstance, 0, len(defaultRegistry.instances))
	for _, instance := range defaultRegistry.instances {
		result = append(result, instance)
	}
	return result
}

// RemoveInstance removes a plugin instance.
// Note: caller should stop the plugin before removing.
//
// Parameters:
//   - instanceID: unique identifier of the instance
//
// Returns:
//   - error: returns error if instance does not exist
func RemoveInstance(instanceID string) error {
	defaultRegistry.mu.Lock()
	defer defaultRegistry.mu.Unlock()

	if _, exists := defaultRegistry.instances[instanceID]; !exists {
		return fmt.Errorf("instance not found: %s", instanceID)
	}

	delete(defaultRegistry.instances, instanceID)
	return nil
}

// CountFactories returns the number of registered plugin factories.
func CountFactories() int {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()
	return len(defaultRegistry.factories)
}

// CountInstances returns the number of created plugin instances.
func CountInstances() int {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()
	return len(defaultRegistry.instances)
}

// ===== Legacy API compatibility =====
// The following functions maintain backward compatibility but are deprecated.
// Please use the new APIs instead.

// Register registers a plugin (deprecated, use RegisterFactory).
// Deprecated: Use RegisterFactory instead.
func Register(p plugGo.Plugin) {
	// Kept for backward compatibility
	// Not recommended for use
}

// Get returns a plugin (deprecated).
// Deprecated: Use GetInstance instead.
func Get(name string) (plugGo.Plugin, bool) {
	// Try to find as instance ID
	instance, ok := GetInstance(name)
	if ok {
		return instance.Plugin(), true
	}
	return nil, false
}

// GetAll returns all plugins (deprecated).
// Deprecated: Use GetAllInstances instead.
func GetAll() []plugGo.Plugin {
	instances := GetAllInstances()
	result := make([]plugGo.Plugin, len(instances))
	for i, instance := range instances {
		result[i] = instance.Plugin()
	}
	return result
}

// Count returns the number of plugins (deprecated).
// Deprecated: Use CountInstances instead.
func Count() int {
	return CountInstances()
}
