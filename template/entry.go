package template

import (
	"context"
	"fmt"
	"sync"

	"github.com/seencxy/plugGo"
	plugGoConfig "github.com/seencxy/plugGo/config"
	"github.com/seencxy/plugGo/template/config"
)

// TemplateEntry is the plugin Entry.
// Implements plugGo.Entry interface, managed by Boot bootstrapper.
type TemplateEntry struct {
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
func (e *TemplateEntry) Bootstrap(ctx context.Context) {
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
func (e *TemplateEntry) Interrupt(ctx context.Context) {
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
func (e *TemplateEntry) GetName() string {
	return e.name
}

// GetType returns the entry type.
func (e *TemplateEntry) GetType() string {
	return e.entryType
}

// GetDescription returns the description.
func (e *TemplateEntry) GetDescription() string {
	return e.description
}

// String returns string representation.
func (e *TemplateEntry) String() string {
	return fmt.Sprintf("TemplateEntry{name=%s, enabled=%v}", e.name, e.enabled)
}

// GetConfig returns current config.
func (e *TemplateEntry) GetConfig() *config.Config {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.cfg
}

// Reload reloads with new config.
func (e *TemplateEntry) Reload(newCfg *config.Config) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.cfg = newCfg
	if e.plugin != nil {
		return e.plugin.Reload(newCfg)
	}
	return nil
}

// RegisterTemplateEntry creates Entry instances from boot.yaml.
// Supports multi-instance: each array element creates one instance.
func RegisterTemplateEntry(raw []byte) map[string]plugGo.Entry {
	result := make(map[string]plugGo.Entry)

	// Change "template" to your plugin config key
	if !plugGoConfig.HasYAMLSection(raw, "template") {
		return result
	}

	type bootConfig struct {
		Entries []config.Config `yaml:"template"`
	}

	var cfg bootConfig
	if err := plugGoConfig.UnmarshalYAML(raw, &cfg); err != nil {
		fmt.Printf("[TemplateEntry] Config parse error: %v\n", err)
		return result
	}

	for i := range cfg.Entries {
		entryCfg := &cfg.Entries[i]
		name := entryCfg.Name

		if name == "" {
			continue
		}

		if _, exists := result[name]; exists {
			fmt.Printf("[TemplateEntry] Duplicate name: %s, skipping\n", name)
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
			fmt.Sprintf("template-%s", name),
			level,
		)

		entry := &TemplateEntry{
			name:        name,
			entryType:   "TemplateEntry",
			description: fmt.Sprintf("Template plugin [%s]", name),
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
	plugGo.RegisterPluginEntryRegFunc(RegisterTemplateEntry)
}
