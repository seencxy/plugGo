package announcement

import (
	"fmt"
	"sync"

	"github.com/seencxy/plugGo"
	"github.com/seencxy/plugGo/example/announcement/config"
)

// Plugin is the announcement monitor plugin.
// Plugin instance only handles business logic, config management is handled by factory and framework.
type Plugin struct {
	id         string         // Instance unique identifier
	pluginType string         // Plugin type name
	version    string         // Plugin version
	cfg        *config.Config // Current config
	logger     plugGo.Logger  // Logger
	monitor    *Monitor       // Monitor
	mu         sync.RWMutex   // Protects concurrent access
}

// ID returns the plugin instance ID.
func (p *Plugin) ID() string {
	return p.id
}

// PluginType returns the plugin type name.
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

	if !p.cfg.Enabled {
		p.logger.Info("Plugin is disabled, skipping start")
		return nil
	}

	// Create and start monitor
	p.monitor = NewMonitor(p.cfg, p.logger)
	if err := p.monitor.Start(); err != nil {
		p.logger.Error("Failed to start monitor:", err)
		return fmt.Errorf("failed to start monitor: %w", err)
	}

	p.logger.Info("Plugin started successfully")
	return nil
}

// Stop stops the plugin.
func (p *Plugin) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.monitor != nil {
		if err := p.monitor.Stop(); err != nil {
			p.logger.Error("Failed to stop monitor:", err)
			return fmt.Errorf("failed to stop monitor: %w", err)
		}
		p.monitor = nil
	}

	p.logger.Info("Plugin stopped")
	return nil
}

// GetLogger returns the logger.
func (p *Plugin) GetLogger() plugGo.Logger {
	return p.logger
}

// SetLogger sets the logger.
func (p *Plugin) SetLogger(logger plugGo.Logger) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.logger = logger

	// If monitor already exists, also update its logger
	if p.monitor != nil {
		// Note: simplified handling, may need to rebuild monitor in practice
		p.logger.Warn("Logger updated, but monitor still uses old logger. Consider reloading.")
	}
}

// Reload reloads the plugin with new config.
func (p *Plugin) Reload(newConfig interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Type assertion
	cfg, ok := newConfig.(*config.Config)
	if !ok {
		return fmt.Errorf("invalid config type: expected *config.Config, got %T", newConfig)
	}

	// Check if plugin is running
	isRunning := p.monitor != nil

	// If plugin is running, stop Monitor first
	if isRunning {
		p.logger.Info("Stopping monitor for config reload")
		if err := p.monitor.Stop(); err != nil {
			p.logger.Error("Failed to stop monitor during reload:", err)
			return fmt.Errorf("failed to stop monitor: %w", err)
		}
	}

	// Save old config for rollback
	oldCfg := p.cfg
	p.cfg = cfg

	// If was running and new config enables plugin, restart Monitor with new config
	if isRunning && p.cfg.Enabled {
		p.logger.Info("Restarting monitor with new configuration")
		p.monitor = NewMonitor(p.cfg, p.logger)
		if err := p.monitor.Start(); err != nil {
			p.logger.Error("Failed to restart monitor:", err)
			// Try to restore old config and restart
			p.cfg = oldCfg
			p.monitor = NewMonitor(p.cfg, p.logger)
			_ = p.monitor.Start()
			return fmt.Errorf("failed to restart monitor: %w", err)
		}
	} else {
		p.monitor = nil
	}

	p.logger.Info("Configuration reloaded successfully")
	return nil
}
