package template

import (
	"context"
	"fmt"
	"sync"

	"github.com/seencxy/plugGo"
	plugGoConfig "github.com/seencxy/plugGo/config"
	"github.com/seencxy/plugGo/template/config"
)

// Entry is the plugin Entry.
// Implements plugGo.Entry interface, managed by Boot bootstrapper.
type Entry struct {
	name        string
	entryType   string
	description string
	cfg         *config.Config
	logger      plugGo.Logger
	plugin      *Plugin
	enabled     bool
	mu          sync.RWMutex
}

// Bootstrap starts the entry.
func (e *Entry) Bootstrap(ctx context.Context) {
	if !e.enabled {
		e.logger.Info(fmt.Sprintf("[%s] Entry disabled, skipping", e.name))
		return
	}

	e.logger.Info(fmt.Sprintf("[%s] Bootstrapping...", e.name))

	// Create and start plugin
	e.plugin = NewPlugin(e.name, e.cfg, e.logger)
	if err := e.plugin.Start(); err != nil {
		e.logger.Error(fmt.Sprintf("[%s] Failed to start: %v", e.name, err))
		return
	}

	e.logger.Info(fmt.Sprintf("[%s] Bootstrapped successfully", e.name))
}

// Interrupt stops the entry.
func (e *Entry) Interrupt(ctx context.Context) {
	if !e.enabled || e.plugin == nil {
		return
	}

	e.logger.Info(fmt.Sprintf("[%s] Interrupting...", e.name))

	if err := e.plugin.Stop(); err != nil {
		e.logger.Error(fmt.Sprintf("[%s] Failed to stop: %v", e.name, err))
	}

	e.logger.Info(fmt.Sprintf("[%s] Interrupted", e.name))
}

// GetName returns the instance name.
func (e *Entry) GetName() string {
	return e.name
}

// GetType returns the entry type.
func (e *Entry) GetType() string {
	return e.entryType
}

// GetDescription returns the description.
func (e *Entry) GetDescription() string {
	return e.description
}

// String returns string representation.
func (e *Entry) String() string {
	return fmt.Sprintf("Entry{name=%s, enabled=%v}", e.name, e.enabled)
}

// GetConfig returns current config.
func (e *Entry) GetConfig() *config.Config {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.cfg
}

// Reload reloads with new config.
func (e *Entry) Reload(newCfg *config.Config) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.cfg = newCfg
	if e.plugin != nil {
		return e.plugin.Reload(newCfg)
	}
	return nil
}

// RegisterEntry creates Entry instances from boot.yaml.
// Supports multi-instance: each array element creates one instance.
func RegisterEntry(raw []byte) map[string]plugGo.Entry {
	result := make(map[string]plugGo.Entry)

	if !plugGoConfig.HasYAMLSection(raw, PluginName) {
		return result
	}

	var entries []config.Config
	if err := plugGoConfig.UnmarshalYAMLSection(raw, PluginName, &entries); err != nil {
		fmt.Printf("[%s] Config parse error: %v\n", EntryTypeName, err)
		return result
	}

	for i := range entries {
		entryCfg := &entries[i]
		name := entryCfg.Name

		if name == "" {
			continue
		}

		if _, exists := result[name]; exists {
			fmt.Printf("[%s] Duplicate name: %s, skipping\n", EntryTypeName, name)
			continue
		}

		// Parse log level
		level := plugGo.InfoLevel
		switch entryCfg.LogLevel {
		case "debug":
			level = plugGo.DebugLevel
		case "warn":
			level = plugGo.WarnLevel
		case "error":
			level = plugGo.ErrorLevel
		}

		logger := plugGo.NewStandardLogger(
			fmt.Sprintf(LoggerPrefix, name),
			level,
		)

		entry := &Entry{
			name:        name,
			entryType:   EntryTypeName,
			description: fmt.Sprintf(PluginDescription, name),
			cfg:         entryCfg,
			logger:      logger,
			enabled:     entryCfg.Enabled,
		}

		result[name] = entry
		logger.Info(fmt.Sprintf("[%s] Entry registered", name))
	}

	return result
}

// Auto-register on import
func init() {
	plugGo.RegisterPluginEntryRegFunc(RegisterEntry)
}
