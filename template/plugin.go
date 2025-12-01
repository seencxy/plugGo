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
	cfg        *config.Config
	logger     plugGo.Logger
	running    bool
	stopCh     chan struct{}
	status     plugGo.PluginStatus
	statusCh   chan plugGo.StatusEvent
	mu         sync.RWMutex
}

// NewPlugin creates a new plugin instance.
func NewPlugin(id string, cfg *config.Config, logger plugGo.Logger) *Plugin {
	return &Plugin{
		id:         id,
		pluginType: PluginName,
		cfg:        cfg,
		logger:     logger,
		stopCh:     make(chan struct{}),
		status:     plugGo.StatusIdle,
		statusCh:   make(chan plugGo.StatusEvent, 10), // buffered channel to avoid blocking
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
	//
	// Example error handling:
	// if err := p.initConnection(); err != nil {
	//     p.updateStatus(plugGo.StatusError, err)
	//     return fmt.Errorf("failed to initialize: %w", err)
	// }

	p.running = true
	p.updateStatus(plugGo.StatusRunning, nil)
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
	//
	// Example error handling:
	// if err := p.cleanup(); err != nil {
	//     p.logger.Error("Cleanup failed:", err)
	//     p.updateStatus(plugGo.StatusError, err)
	//     return fmt.Errorf("failed to cleanup: %w", err)
	// }

	p.running = false
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
