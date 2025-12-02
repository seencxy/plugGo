package plugGo

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// hookFuncM is the hook function map type.
type hookFuncM map[string]map[string]func(ctx context.Context)

func newHookFuncM() hookFuncM {
	return make(hookFuncM)
}

func (m hookFuncM) addFunc(entryType, entryName string, f func(ctx context.Context)) {
	if _, ok := m[entryType]; !ok {
		m[entryType] = make(map[string]func(ctx context.Context))
	}
	m[entryType][entryName] = f
}

func (m hookFuncM) getFunc(entryType, entryName string) func(ctx context.Context) {
	if inner, ok := m[entryType]; ok {
		if f, ok := inner[entryName]; ok {
			return f
		}
	}
	return func(ctx context.Context) {}
}

// Boot is the bootstrapper struct.
type Boot struct {
	configPath    string
	configRaw     []byte
	embedFS       *embed.FS
	beforeHookF   hookFuncM
	afterHookF    hookFuncM
	pluginEntries map[string]map[string]Entry
	userEntries   map[string]map[string]Entry
	logger        Logger
	mu            sync.RWMutex

	// Shutdown timeout configuration
	shutdownTimeout      time.Duration
	entryShutdownTimeout time.Duration
}

// NewBoot creates a new Boot instance.
func NewBoot(opts ...BootOption) *Boot {
	cfg := &BootConfig{
		ConfigPath: "boot.yaml",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	boot := &Boot{
		configPath:           cfg.ConfigPath,
		configRaw:            cfg.ConfigRaw,
		beforeHookF:          newHookFuncM(),
		afterHookF:           newHookFuncM(),
		pluginEntries:        make(map[string]map[string]Entry),
		userEntries:          make(map[string]map[string]Entry),
		logger:               NewDefaultLogger("boot"),
		shutdownTimeout:      cfg.ShutdownTimeout,
		entryShutdownTimeout: cfg.EntryShutdownTimeout,
	}

	// Read config
	raw := boot.readYAML()

	// Call all plugin registration functions to create Entries
	for _, f := range GlobalAppCtx.ListPluginEntryRegFunc() {
		for name, entry := range f(raw) {
			entryType := entry.GetType()
			if boot.pluginEntries[entryType] == nil {
				boot.pluginEntries[entryType] = make(map[string]Entry)
			}
			boot.pluginEntries[entryType][name] = entry
			GlobalAppCtx.RegisterEntry(entry)
		}
	}

	// Call all user registration functions to create Entries
	for _, f := range GlobalAppCtx.ListUserEntryRegFunc() {
		for name, entry := range f(raw) {
			entryType := entry.GetType()
			if boot.userEntries[entryType] == nil {
				boot.userEntries[entryType] = make(map[string]Entry)
			}
			boot.userEntries[entryType][name] = entry
			GlobalAppCtx.RegisterEntry(entry)
		}
	}

	return boot
}

// WithEmbedFS sets the embedded file system.
func WithEmbedFS(fs *embed.FS) BootOption {
	return func(c *BootConfig) {
		// This needs to extend BootConfig
	}
}

// AddHookFuncBeforeBootstrap adds a hook function to run before Bootstrap.
func (b *Boot) AddHookFuncBeforeBootstrap(entryType, entryName string, f func(ctx context.Context)) {
	if f == nil {
		return
	}
	b.beforeHookF.addFunc(entryType, entryName, f)
}

// AddHookFuncAfterBootstrap adds a hook function to run after Bootstrap.
func (b *Boot) AddHookFuncAfterBootstrap(entryType, entryName string, f func(ctx context.Context)) {
	if f == nil {
		return
	}
	b.afterHookF.addFunc(entryType, entryName, f)
}

// Bootstrap starts all Entries.
func (b *Boot) Bootstrap(ctx context.Context) {
	defer b.syncLog()

	// Start plugin Entries
	for entryType, byName := range b.pluginEntries {
		for entryName, entry := range byName {
			b.beforeHookF.getFunc(entryType, entryName)(ctx)
			b.logger.Info(fmt.Sprintf("Bootstrapping [%s] %s", entryType, entryName))
			entry.Bootstrap(ctx)
			b.afterHookF.getFunc(entryType, entryName)(ctx)
		}
	}

	// Start user Entries
	for entryType, byName := range b.userEntries {
		for entryName, entry := range byName {
			b.beforeHookF.getFunc(entryType, entryName)(ctx)
			b.logger.Info(fmt.Sprintf("Bootstrapping [%s] %s", entryType, entryName))
			entry.Bootstrap(ctx)
			b.afterHookF.getFunc(entryType, entryName)(ctx)
		}
	}
}

// WaitForShutdownSig waits for shutdown signal.
func (b *Boot) WaitForShutdownSig(ctx context.Context) {
	GlobalAppCtx.WaitForShutdownSig()
	b.Shutdown(ctx)
}

// Shutdown shuts down all Entries.
func (b *Boot) Shutdown(ctx context.Context) {
	// Create context with timeout (if not already set)
	shutdownTimeout := 30 * time.Second
	if b.shutdownTimeout > 0 {
		shutdownTimeout = b.shutdownTimeout
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	// 1. Execute shutdown hooks
	for name, hook := range GlobalAppCtx.ListShutdownHooks() {
		b.logger.Info(fmt.Sprintf("Running shutdown hook: %s", name))
		hook()
	}

	// 2. Interrupt all Entries concurrently with timeout control
	b.interruptWithContext(shutdownCtx)
}

// AddShutdownHookFunc adds a shutdown hook function.
func (b *Boot) AddShutdownHookFunc(name string, f ShutdownHook) {
	GlobalAppCtx.AddShutdownHook(name, f)
}

// interruptWithContext interrupts all Entries concurrently with timeout control.
func (b *Boot) interruptWithContext(ctx context.Context) {
	defer b.syncLog()

	var wg sync.WaitGroup

	// 1. Shutdown user Entries first (concurrent)
	for entryType, byName := range b.userEntries {
		for entryName, entry := range byName {
			wg.Add(1)
			go func(et, en string, e Entry) {
				defer wg.Done()
				b.interruptSingleEntry(ctx, et, en, e)
			}(entryType, entryName, entry)
		}
	}

	// Wait for user Entries to shutdown
	b.waitWithTimeout(ctx, &wg, "user entries")

	// 2. Then shutdown plugin Entries (concurrent)
	wg = sync.WaitGroup{} // Reset
	for entryType, byName := range b.pluginEntries {
		for entryName, entry := range byName {
			wg.Add(1)
			go func(et, en string, e Entry) {
				defer wg.Done()
				b.interruptSingleEntry(ctx, et, en, e)
			}(entryType, entryName, entry)
		}
	}

	// Wait for plugin Entries to shutdown
	b.waitWithTimeout(ctx, &wg, "plugin entries")
}

// waitWithTimeout waits for all goroutines in WaitGroup or until context timeout.
func (b *Boot) waitWithTimeout(ctx context.Context, wg *sync.WaitGroup, name string) {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		b.logger.Info(fmt.Sprintf("All %s interrupted successfully", name))
	case <-ctx.Done():
		b.logger.Warn(fmt.Sprintf("Shutdown timeout exceeded for %s, forcing exit", name))
	}
}

// interruptSingleEntry interrupts a single Entry with timeout control.
func (b *Boot) interruptSingleEntry(ctx context.Context, entryType, entryName string, entry Entry) {
	b.logger.Info(fmt.Sprintf("Interrupting [%s] %s", entryType, entryName))

	// Create timeout context for this Entry (default 10s)
	entryTimeout := 10 * time.Second
	if b.entryShutdownTimeout > 0 {
		entryTimeout = b.entryShutdownTimeout
	}

	entryCtx, cancel := context.WithTimeout(ctx, entryTimeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		entry.Interrupt(entryCtx)
		close(done)
	}()

	select {
	case <-done:
		b.logger.Info(fmt.Sprintf("Interrupted [%s] %s", entryType, entryName))
	case <-entryCtx.Done():
		b.logger.Warn(fmt.Sprintf("Interrupt timeout [%s] %s", entryType, entryName))
	}
}

// GetEntry returns the specified Entry.
func (b *Boot) GetEntry(entryType, entryName string) Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if byName, ok := b.pluginEntries[entryType]; ok {
		if entry, ok := byName[entryName]; ok {
			return entry
		}
	}
	if byName, ok := b.userEntries[entryType]; ok {
		if entry, ok := byName[entryName]; ok {
			return entry
		}
	}
	return nil
}

// ===== Multi-instance management methods =====

// GetEntriesByType returns all Entry instances of the specified type.
func (b *Boot) GetEntriesByType(entryType string) map[string]Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make(map[string]Entry)
	if byName, ok := b.pluginEntries[entryType]; ok {
		for name, entry := range byName {
			result[name] = entry
		}
	}
	if byName, ok := b.userEntries[entryType]; ok {
		for name, entry := range byName {
			result[name] = entry
		}
	}
	return result
}

// GetAllEntries returns all Entry instances.
func (b *Boot) GetAllEntries() map[string]map[string]Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make(map[string]map[string]Entry)
	for entryType, byName := range b.pluginEntries {
		if result[entryType] == nil {
			result[entryType] = make(map[string]Entry)
		}
		for name, entry := range byName {
			result[entryType][name] = entry
		}
	}
	for entryType, byName := range b.userEntries {
		if result[entryType] == nil {
			result[entryType] = make(map[string]Entry)
		}
		for name, entry := range byName {
			result[entryType][name] = entry
		}
	}
	return result
}

// CountEntries returns the total number of Entries.
func (b *Boot) CountEntries() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	count := 0
	for _, byName := range b.pluginEntries {
		count += len(byName)
	}
	for _, byName := range b.userEntries {
		count += len(byName)
	}
	return count
}

// readYAML reads the YAML config file.
func (b *Boot) readYAML() []byte {
	// Use raw config first
	if len(b.configRaw) > 0 {
		return b.configRaw
	}

	// Read from embedded file system
	if b.embedFS != nil {
		res, err := b.embedFS.ReadFile(b.configPath)
		if err != nil {
			b.logger.Error(fmt.Sprintf("Failed to read config from embed FS: %v", err))
			os.Exit(1)
		}
		return res
	}

	// Read from local file
	if !filepath.IsAbs(b.configPath) {
		wd, _ := os.Getwd()
		b.configPath = filepath.Join(wd, b.configPath)
	}

	res, err := os.ReadFile(b.configPath)
	if err != nil {
		b.logger.Error(fmt.Sprintf("Failed to read config file %s: %v", b.configPath, err))
		os.Exit(1)
	}
	return res
}

// syncLog syncs logs and handles panic.
func (b *Boot) syncLog() {
	if r := recover(); r != nil {
		stackTrace := fmt.Sprintf("Panic occurred, shutting down...\n%s", string(debug.Stack()))
		b.logger.Error(stackTrace)
		b.logger.Error(fmt.Sprintf("Root cause: %v", r))
		os.Exit(1)
	}
}

// ParseYAML parses YAML config into the target struct.
func ParseYAML(raw []byte, target interface{}) error {
	return yaml.Unmarshal(raw, target)
}
