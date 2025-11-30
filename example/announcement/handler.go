package announcement

import (
	"fmt"
	"time"

	"github.com/seencxy/plugGo"
	"github.com/seencxy/plugGo/example/announcement/config"
)

// Monitor is the announcement monitor.
type Monitor struct {
	cfg     *config.Config
	logger  plugGo.Logger
	stopCh  chan struct{}
	stopped bool
}

// NewMonitor creates a new monitor instance.
func NewMonitor(cfg *config.Config, logger plugGo.Logger) *Monitor {
	return &Monitor{
		cfg:    cfg,
		logger: logger,
		stopCh: make(chan struct{}),
	}
}

// Start starts monitoring.
func (m *Monitor) Start() error {
	if !m.cfg.Enabled {
		return fmt.Errorf("plugin is disabled")
	}

	m.logger.Info(fmt.Sprintf("Starting monitor for %d sources", len(m.cfg.Sources)))
	for _, source := range m.cfg.Sources {
		m.logger.Debug("Starting monitor for source:", source.Name)
		go m.monitorSource(source)
	}

	return nil
}

// Stop stops monitoring.
func (m *Monitor) Stop() error {
	if m.stopped {
		return nil
	}

	m.logger.Info("Stopping monitor")
	close(m.stopCh)
	m.stopped = true
	return nil
}

// monitorSource monitors a single announcement source.
func (m *Monitor) monitorSource(source config.Source) {
	ticker := time.NewTicker(time.Duration(source.Interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			// Here should implement actual announcement fetching logic
			// Example code only prints logs
			m.checkAnnouncements(source)
		}
	}
}

// checkAnnouncements checks for announcement updates.
func (m *Monitor) checkAnnouncements(source config.Source) {
	// TODO: Implement actual announcement fetching, filtering and notification logic
	// This is just an example
	m.logger.Trace(fmt.Sprintf("Checking announcements from source: %s, URL: %s", source.Name, source.URL))
}
