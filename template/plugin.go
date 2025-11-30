package template

import (
	"fmt"
	"sync"

	"github.com/seencxy/plugGo"
	"github.com/seencxy/plugGo/template/config"
)

// Plugin is the plugin implementation.
type Plugin struct {
	id         string
	pluginType string
	version    string
	cfg        *config.Config
	logger     plugGo.Logger
	running    bool
	stopCh     chan struct{}
	mu         sync.RWMutex
}

// NewPlugin creates a new plugin instance.
func NewPlugin(id string, cfg *config.Config, logger plugGo.Logger) *Plugin {
	return &Plugin{
		id:         id,
		pluginType: "template",
		version:    "1.0.0",
		cfg:        cfg,
		logger:     logger,
		stopCh:     make(chan struct{}),
	}
}

// ID returns the plugin ID.
func (p *Plugin) ID() string {
	return p.id
}

// PluginType returns the plugin type.
func (p *Plugin) PluginType() string {
	return p.pluginType
}

// Version returns the plugin version.
func (p *Plugin) Version() string {
	return p.version
}

// Start starts the plugin.
func (p *Plugin) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("plugin already running")
	}

	if !p.cfg.Enabled {
		p.logger.Info("Plugin is disabled")
		return nil
	}

	p.logger.Info(fmt.Sprintf("Starting plugin, endpoint: %s, interval: %ds", p.cfg.Endpoint, p.cfg.Interval))

	// TODO: Add your startup logic here
	// Example: start background workers, initialize connections, etc.

	p.running = true
	return nil
}

// Stop stops the plugin.
func (p *Plugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return nil
	}

	p.logger.Info("Stopping plugin...")

	// Signal stop
	close(p.stopCh)

	// TODO: Add your cleanup logic here
	// Example: stop workers, close connections, etc.

	p.running = false
	p.logger.Info("Plugin stopped")
	return nil
}

// GetLogger returns the logger.
func (p *Plugin) GetLogger() plugGo.Logger {
	return p.logger
}

// SetLogger sets the logger.
func (p *Plugin) SetLogger(logger plugGo.Logger) {
	p.logger = logger
}

// Reload reloads the plugin with new config.
func (p *Plugin) Reload(newCfg interface{}) error {
	cfg, ok := newCfg.(*config.Config)
	if !ok {
		return fmt.Errorf("invalid config type")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.logger.Info("Reloading plugin config...")
	p.cfg = cfg

	// TODO: Add your reload logic here

	p.logger.Info("Plugin config reloaded")
	return nil
}
