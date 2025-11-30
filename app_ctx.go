package plugGo

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// AppContext is the global application context.
// Manages all Entry instances, registration functions and shutdown hooks.
type AppContext struct {
	// entries stores all registered Entries.
	// Structure: map[entryType]map[entryName]Entry
	entries map[string]map[string]Entry

	// regFuncs stores all Entry registration functions.
	// Structure: map[entryType][]RegFunc
	regFuncs map[string][]RegFunc

	// shutdownHooks stores shutdown hooks.
	shutdownHooks map[string]ShutdownHook

	// shutdownSig is the shutdown signal channel.
	shutdownSig chan os.Signal

	mu sync.RWMutex
}

// GlobalAppCtx is the global application context singleton.
var GlobalAppCtx = &AppContext{
	entries:       make(map[string]map[string]Entry),
	regFuncs:      make(map[string][]RegFunc),
	shutdownHooks: make(map[string]ShutdownHook),
	shutdownSig:   make(chan os.Signal, 1),
}

// RegisterEntry registers an Entry to the global context.
func (ctx *AppContext) RegisterEntry(entry Entry) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	entryType := entry.GetType()
	if ctx.entries[entryType] == nil {
		ctx.entries[entryType] = make(map[string]Entry)
	}
	ctx.entries[entryType][entry.GetName()] = entry
}

// GetEntry returns the Entry with the specified type and name.
func (ctx *AppContext) GetEntry(entryType, entryName string) Entry {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	if byName, ok := ctx.entries[entryType]; ok {
		return byName[entryName]
	}
	return nil
}

// ListEntriesByType lists all Entries of the specified type.
func (ctx *AppContext) ListEntriesByType(entryType string) map[string]Entry {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	result := make(map[string]Entry)
	if byName, ok := ctx.entries[entryType]; ok {
		for k, v := range byName {
			result[k] = v
		}
	}
	return result
}

// ListEntries lists all Entries.
// Returns map[entryType]map[entryName]Entry
func (ctx *AppContext) ListEntries() map[string]map[string]Entry {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	result := make(map[string]map[string]Entry)
	for entryType, byName := range ctx.entries {
		result[entryType] = make(map[string]Entry)
		for name, entry := range byName {
			result[entryType][name] = entry
		}
	}
	return result
}

// RemoveEntry removes an Entry.
func (ctx *AppContext) RemoveEntry(entryType, entryName string) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	if byName, ok := ctx.entries[entryType]; ok {
		delete(byName, entryName)
	}
}

// RegisterPluginEntryRegFunc registers a plugin Entry registration function.
func (ctx *AppContext) RegisterPluginEntryRegFunc(f RegFunc) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.regFuncs[PluginEntryType] = append(ctx.regFuncs[PluginEntryType], f)
}

// RegisterUserEntryRegFunc registers a user-defined Entry registration function.
func (ctx *AppContext) RegisterUserEntryRegFunc(f RegFunc) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.regFuncs[UserEntryType] = append(ctx.regFuncs[UserEntryType], f)
}

// ListPluginEntryRegFunc lists all plugin Entry registration functions.
func (ctx *AppContext) ListPluginEntryRegFunc() []RegFunc {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.regFuncs[PluginEntryType]
}

// ListUserEntryRegFunc lists all user-defined Entry registration functions.
func (ctx *AppContext) ListUserEntryRegFunc() []RegFunc {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.regFuncs[UserEntryType]
}

// AddShutdownHook adds a shutdown hook.
func (ctx *AppContext) AddShutdownHook(name string, hook ShutdownHook) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.shutdownHooks[name] = hook
}

// ListShutdownHooks lists all shutdown hooks.
func (ctx *AppContext) ListShutdownHooks() map[string]ShutdownHook {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	result := make(map[string]ShutdownHook)
	for k, v := range ctx.shutdownHooks {
		result[k] = v
	}
	return result
}

// WaitForShutdownSig waits for shutdown signal.
func (ctx *AppContext) WaitForShutdownSig() {
	signal.Notify(ctx.shutdownSig, syscall.SIGINT, syscall.SIGTERM)
	<-ctx.shutdownSig
}

// ===== Global convenience functions =====

// RegisterPluginEntryRegFunc registers a plugin Entry registration function (global convenience function).
func RegisterPluginEntryRegFunc(f RegFunc) {
	GlobalAppCtx.RegisterPluginEntryRegFunc(f)
}

// RegisterUserEntryRegFunc registers a user-defined Entry registration function (global convenience function).
func RegisterUserEntryRegFunc(f RegFunc) {
	GlobalAppCtx.RegisterUserEntryRegFunc(f)
}

// RegisterEntry registers an Entry (global convenience function).
func RegisterEntry(entry Entry) {
	GlobalAppCtx.RegisterEntry(entry)
}

// GetEntry returns an Entry (global convenience function).
func GetEntry(entryType, entryName string) Entry {
	return GlobalAppCtx.GetEntry(entryType, entryName)
}
