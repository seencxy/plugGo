package announcement

import (
	"context"
	"fmt"
	"sync"

	"github.com/seencxy/plugGo"
	plugGoConfig "github.com/seencxy/plugGo/config"
	"github.com/seencxy/plugGo/example/announcement/config"
)

// AnnouncementEntry is the announcement monitor Entry.
// Implements plugGo.Entry interface, managed by Boot bootstrapper.
type AnnouncementEntry struct {
	name        string         // Instance name
	entryType   string         // Entry type
	description string         // Description
	cfg         *config.Config // Config
	logger      plugGo.Logger  // Logger
	monitor     *Monitor       // Monitor
	mu          sync.RWMutex
}

// NewAnnouncementEntry creates a new announcement Entry.
func NewAnnouncementEntry(name string, cfg *config.Config, logger plugGo.Logger) *AnnouncementEntry {
	return &AnnouncementEntry{
		name:        name,
		entryType:   "AnnouncementEntry",
		description: fmt.Sprintf("Announcement monitor entry [%s]", name),
		cfg:         cfg,
		logger:      logger,
	}
}

// Bootstrap starts the Entry.
func (e *AnnouncementEntry) Bootstrap(ctx context.Context) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.cfg.Enabled {
		e.logger.Info(fmt.Sprintf("[%s] Entry is disabled, skipping bootstrap", e.name))
		return
	}

	e.logger.Info(fmt.Sprintf("[%s] Bootstrapping announcement entry...", e.name))

	// Create and start monitor
	e.monitor = NewMonitor(e.cfg, e.logger)
	if err := e.monitor.Start(); err != nil {
		e.logger.Error(fmt.Sprintf("[%s] Failed to start monitor: %v", e.name, err))
		return
	}

	e.logger.Info(fmt.Sprintf("[%s] Announcement entry bootstrapped successfully", e.name))
}

// Interrupt stops the Entry.
func (e *AnnouncementEntry) Interrupt(ctx context.Context) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.logger.Info(fmt.Sprintf("[%s] Interrupting announcement entry...", e.name))

	if e.monitor != nil {
		if err := e.monitor.Stop(); err != nil {
			e.logger.Error(fmt.Sprintf("[%s] Failed to stop monitor: %v", e.name, err))
		}
		e.monitor = nil
	}

	e.logger.Info(fmt.Sprintf("[%s] Announcement entry interrupted", e.name))
}

// GetName returns the Entry name.
func (e *AnnouncementEntry) GetName() string {
	return e.name
}

// GetType returns the Entry type.
func (e *AnnouncementEntry) GetType() string {
	return e.entryType
}

// GetDescription returns the Entry description.
func (e *AnnouncementEntry) GetDescription() string {
	return e.description
}

// String returns string representation.
func (e *AnnouncementEntry) String() string {
	return fmt.Sprintf("AnnouncementEntry{name=%s, enabled=%v}", e.name, e.cfg.Enabled)
}

// ===== Multi-instance support =====

// GetConfig returns current config (for runtime inspection).
func (e *AnnouncementEntry) GetConfig() *config.Config {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.cfg
}

// Reload hot-reloads the config.
func (e *AnnouncementEntry) Reload(newCfg *config.Config) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	wasRunning := e.monitor != nil

	// Stop existing monitor
	if wasRunning {
		if err := e.monitor.Stop(); err != nil {
			return fmt.Errorf("failed to stop monitor: %w", err)
		}
	}

	// Update config
	e.cfg = newCfg

	// If was running and new config is enabled, restart
	if wasRunning && e.cfg.Enabled {
		e.monitor = NewMonitor(e.cfg, e.logger)
		if err := e.monitor.Start(); err != nil {
			return fmt.Errorf("failed to restart monitor: %w", err)
		}
	}

	e.logger.Info(fmt.Sprintf("[%s] Configuration reloaded", e.name))
	return nil
}

// ===== Registration function (multi-instance core) =====

// RegisterAnnouncementEntry is the registration function that creates multiple Entry instances from YAML config.
// boot.yaml format:
//
//	announcement:
//	  - name: "instance1"
//	    enabled: true
//	    sources: [...]
//	  - name: "instance2"
//	    enabled: true
//	    sources: [...]
//
// Each array element creates an independent Entry instance.
func RegisterAnnouncementEntry(raw []byte) map[string]plugGo.Entry {
	result := make(map[string]plugGo.Entry)

	// Check if config has announcement section
	if !plugGoConfig.HasYAMLSection(raw, "announcement") {
		return result
	}

	// Parse config
	type bootConfig struct {
		Announcement []config.Config `yaml:"announcement"`
	}

	var cfg bootConfig
	if err := plugGoConfig.UnmarshalYAML(raw, &cfg); err != nil {
		return result
	}

	// Multi-instance: create independent Entry for each config item
	for i := range cfg.Announcement {
		entryCfg := &cfg.Announcement[i]

		// Instance name: prefer name from config, otherwise auto-generate
		name := entryCfg.Name
		if name == "" {
			name = fmt.Sprintf("announcement-%d", i)
		}

		// Check name uniqueness
		if _, exists := result[name]; exists {
			// Name conflict, append index
			name = fmt.Sprintf("%s-%d", name, i)
		}

		// Create independent logger for each instance
		logger := plugGo.NewDefaultLogger(fmt.Sprintf("announcement-%s", name))

		// Create Entry instance
		entry := NewAnnouncementEntry(name, entryCfg, logger)
		result[name] = entry
	}

	return result
}

// init auto-registers the Entry registration function.
func init() {
	plugGo.RegisterPluginEntryRegFunc(RegisterAnnouncementEntry)
}
