package announcement

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/seencxy/plugGo"
	"github.com/seencxy/plugGo/example/announcement/config"
)

// Plugin is the announcement monitor plugin.
// Plugin instance only handles business logic, config management is handled by factory and framework.
type Plugin struct {
	id         string                  // Instance unique identifier
	pluginType string                  // Plugin type name
	cfg        *config.Config          // Current config
	logger     plugGo.Logger           // Logger
	monitor    *Monitor                // Monitor
	status     plugGo.PluginStatus     // Current status
	statusCh   chan plugGo.StatusEvent // Status notification channel
	mu         sync.RWMutex            // Protects concurrent access
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
	return PluginVersion
}

// Status returns the current plugin status.
func (p *Plugin) Status() plugGo.PluginStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

// StatusNotify returns a read-only channel for receiving status change events.
func (p *Plugin) StatusNotify() <-chan plugGo.StatusEvent {
	return p.statusCh
}

// updateStatus updates the plugin status and sends notification.
// Note: caller must hold the lock (p.mu).
func (p *Plugin) updateStatus(newStatus plugGo.PluginStatus, err error) {
	if p.status != newStatus {
		p.status = newStatus
		// Send status event non-blocking
		select {
		case p.statusCh <- plugGo.StatusEvent{Status: newStatus, Error: err}:
		default:
			// Channel full, skip this event
			p.logger.Warn("Status channel full, event dropped")
		}
	}
}

// Start starts the plugin with context.
func (p *Plugin) Start(ctx context.Context) error {
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
		p.updateStatus(plugGo.StatusError, err)
		return fmt.Errorf("failed to start monitor: %w", err)
	}

	p.updateStatus(plugGo.StatusRunning, nil)
	p.logger.Info("Plugin started successfully")
	return nil
}

// Stop stops the plugin with context.
func (p *Plugin) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.monitor != nil {
		// Extract timeout from context
		timeout := 5 * time.Second
		if deadline, ok := ctx.Deadline(); ok {
			timeout = time.Until(deadline)
			if timeout < 0 {
				timeout = time.Second // At least 1 second
			}
		}

		if err := p.monitor.StopWithTimeout(timeout); err != nil {
			p.logger.Error("Failed to stop monitor:", err)
			p.updateStatus(plugGo.StatusError, err)
			return fmt.Errorf("failed to stop monitor: %w", err)
		}
		p.monitor = nil
	}

	p.updateStatus(plugGo.StatusStopped, nil)
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
		err := fmt.Errorf("invalid config type: expected *config.Config, got %T", newConfig)
		p.updateStatus(plugGo.StatusError, err)
		return err
	}

	// Check if plugin is running
	isRunning := p.monitor != nil

	// If plugin is running, stop Monitor first
	if isRunning {
		p.logger.Info("Stopping monitor for config reload")
		if err := p.monitor.Stop(); err != nil {
			p.logger.Error("Failed to stop monitor during reload:", err)
			p.updateStatus(plugGo.StatusError, err)
			return fmt.Errorf("failed to stop monitor: %w", err)
		}
	}

	// Save old config for rollback
	oldCfg := p.cfg
	oldStatus := p.status
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
			p.updateStatus(plugGo.StatusError, err)
			return fmt.Errorf("failed to restart monitor: %w", err)
		}
		// Restore running status if was running before
		p.updateStatus(plugGo.StatusRunning, nil)
	} else {
		p.monitor = nil
		// Update status based on previous state
		if oldStatus == plugGo.StatusRunning {
			p.updateStatus(plugGo.StatusStopped, nil)
		}
	}

	p.logger.Info("Configuration reloaded successfully")
	return nil
}
