package template

import (
	_ "embed"
	"fmt"

	"github.com/seencxy/plugGo"
	plugGoConfig "github.com/seencxy/plugGo/config"
	"github.com/seencxy/plugGo/registry"
	"github.com/seencxy/plugGo/template/config"
)

//go:embed config.yaml
var defaultConfigData []byte

// Factory is the plugin factory.
type Factory struct{}

// NewFactory creates a factory instance.
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
	if err := plugGoConfig.UnmarshalYAML(defaultConfigData, cfg); err != nil {
		return &config.Config{
			Name:     "default",
			Enabled:  true,
			LogLevel: "info",
			Interval: 60,
			Endpoint: "https://api.example.com",
		}
	}
	return cfg
}

// parseLogLevel converts string log level to plugGo.LogLevel.
func parseLogLevel(level string) plugGo.LogLevel {
	switch level {
	case "trace":
		return plugGo.TraceLevel
	case "debug":
		return plugGo.DebugLevel
	case "warn":
		return plugGo.WarnLevel
	case "error":
		return plugGo.ErrorLevel
	default:
		return plugGo.InfoLevel
	}
}

// ValidateConfig validates the config.
func (f *Factory) ValidateConfig(cfg interface{}) error {
	c, ok := cfg.(*config.Config)
	if !ok {
		return fmt.Errorf("invalid config type")
	}

	if c.Name == "" {
		return fmt.Errorf("name is required")
	}

	if c.Interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}

	return nil
}

// Create creates a new plugin instance.
func (f *Factory) Create(id string, cfg interface{}, logger plugGo.Logger) (plugGo.Plugin, error) {
	c, ok := cfg.(*config.Config)
	if !ok {
		return nil, fmt.Errorf("invalid config type")
	}

	if err := f.ValidateConfig(cfg); err != nil {
		return nil, err
	}

	if logger == nil {
		logger = plugGo.NewStandardLogger(fmt.Sprintf(LoggerPrefix, id), plugGo.ParseLogLevel(c.LogLevel))
	}

	return NewPlugin(id, c, logger), nil
}

// Register factory to registry on import
func init() {
	registry.RegisterFactory(NewFactory())
}
