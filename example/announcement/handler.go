package announcement

import (
	"fmt"
	"sync"
	"time"

	"github.com/seencxy/plugGo"
	"github.com/seencxy/plugGo/example/announcement/config"
)

// Monitor is the announcement monitor.
type Monitor struct {
	cfg     *config.Config
	logger  plugGo.Logger
	stopCh  chan struct{}
	wg      sync.WaitGroup // Wait for all goroutines to exit
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
		m.wg.Add(1) // Count before starting goroutine
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

	// Wait for all goroutines to exit
	m.wg.Wait()

	m.stopped = true
	m.logger.Info("Monitor stopped, all goroutines exited")
	return nil
}

// monitorSource monitors a single announcement source.
func (m *Monitor) monitorSource(source config.Source) {
	defer m.wg.Done() // Decrement counter when goroutine exits

	ticker := time.NewTicker(time.Duration(source.Interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			m.logger.Debug("Monitor goroutine exiting for source:", source.Name)
			return
		case <-ticker.C:
			// Here should implement actual announcement fetching logic
			// Example code only prints logs
			m.checkAnnouncements(source)
		}
	}
}

// StopWithTimeout stops the monitor with a timeout.
func (m *Monitor) StopWithTimeout(timeout time.Duration) error {
	if m.stopped {
		return nil
	}

	m.logger.Info(fmt.Sprintf("Stopping monitor with timeout: %v", timeout))
	close(m.stopCh)

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		m.stopped = true
		m.logger.Info("Monitor stopped gracefully")
		return nil
	case <-time.After(timeout):
		m.stopped = true
		m.logger.Warn("Monitor stop timeout, some goroutines may still be running")
		return fmt.Errorf("stop timeout after %v", timeout)
	}
}

// checkAnnouncements checks for announcement updates.
func (m *Monitor) checkAnnouncements(source config.Source) {
	// TODO: Implement actual announcement fetching, filtering and notification logic
	// This is just an example
	m.logger.Trace(fmt.Sprintf("Checking announcements from source: %s, URL: %s", source.Name, source.URL))
}
