package announcement

import (
	_ "embed"
	"fmt"

	"github.com/seencxy/plugGo"
	plugGoConfig "github.com/seencxy/plugGo/config"
	"github.com/seencxy/plugGo/example/announcement/config"
	"github.com/seencxy/plugGo/registry"
)

//go:embed config.yaml
var defaultConfigData []byte

// Factory is the announcement plugin factory.
type Factory struct{}

// NewFactory creates a plugin factory instance.
func NewFactory() *Factory {
	return &Factory{}
}

// Name returns the plugin type name.
func (f *Factory) Name() string {
	return PluginName
}

// Version returns the plugin version.
func (f *Factory) Version() string {
	return PluginVersion
}

// DefaultConfig returns the default config.
func (f *Factory) DefaultConfig() interface{} {
	cfg := &config.Config{}

	// Load default config from embedded config file
	loader := &plugGoConfig.Loader{}
	if err := loader.LoadWithFallback(f.Name(), defaultConfigData, cfg); err != nil {
		// If loading fails, return empty config
		return &config.Config{}
	}

	return cfg
}

// ValidateConfig validates the config.
func (f *Factory) ValidateConfig(cfg interface{}) error {
	announcementCfg, ok := cfg.(*config.Config)
	if !ok {
		return fmt.Errorf("invalid config type: expected *config.Config, got %T", cfg)
	}

	// Validate announcement sources are configured
	if len(announcementCfg.Sources) == 0 {
		return fmt.Errorf("no announcement sources configured")
	}

	// Validate each announcement source config
	for i, source := range announcementCfg.Sources {
		if source.Name == "" {
			return fmt.Errorf("source[%d]: name is required", i)
		}
		if source.URL == "" {
			return fmt.Errorf("source[%d]: url is required", i)
		}
		if source.Interval <= 0 {
			return fmt.Errorf("source[%d]: interval must be positive", i)
		}
	}

	return nil
}

// Create creates a plugin instance.
func (f *Factory) Create(instanceID string, cfg interface{}, logger plugGo.Logger) (plugGo.Plugin, error) {
	announcementCfg, ok := cfg.(*config.Config)
	if !ok {
		return nil, fmt.Errorf("invalid config type: expected *config.Config, got %T", cfg)
	}

	// Create plugin instance
	plugin := &Plugin{
		id:         instanceID,
		pluginType: PluginName,
		cfg:        announcementCfg,
		logger:     logger,
		status:     plugGo.StatusIdle,
		statusCh:   make(chan plugGo.StatusEvent, 10), // buffered channel to avoid blocking
		notifyCh:   make(chan any, 100),               // buffered channel for external notifications
	}

	return plugin, nil
}

// Auto-register plugin factory in init.
func init() {
	registry.RegisterFactory(NewFactory())
}
