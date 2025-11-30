# PlugGo

[中文文档](README_zh.md)

A lightweight Go plugin framework inspired by [rk-boot](https://github.com/rookie-ninja/rk-boot), supporting unified YAML configuration for managing multiple plugins.

## Features

- **Unified Configuration**: Single `boot.yaml` to manage all plugin configs with multi-instance support
- **Entry Interface**: Standardized lifecycle management (`Bootstrap`/`Interrupt`)
- **Auto Registration**: Auto-register plugins via `import _` and `init()`
- **Hook System**: Before/After Bootstrap hooks and shutdown hooks
- **Type Safety**: Compile-time static linking, avoiding runtime loading complexity
- **Framework Agnostic**: No dependency on specific web frameworks

## Quick Start

### 1. Install

```bash
go get plugGo
```

### 2. Create boot.yaml

```yaml
# boot.yaml
announcement:
  - name: "official"
    enabled: true
    logLevel: "info"
    sources:
      - name: "Official Announcements"
        url: "https://example.com/api/announcements"
        interval: 300
    notifications:
      - type: "webhook"
        url: "https://your-webhook-url/notify"
```

### 3. Write main.go

```go
package main

import (
    "context"
    "plugGo"
    
    // Import plugins (triggers init auto-registration)
    _ "plugGo/example/announcement"
)

func main() {
    // Create Boot bootstrapper (reads boot.yaml by default)
    boot := plugGo.NewBoot()
    
    // Bootstrap all Entries
    boot.Bootstrap(context.Background())
    
    // Wait for shutdown signal
    boot.WaitForShutdownSig(context.Background())
}
```

### 4. Run

```bash
cd example/host
go run main.go
```

## Project Structure

```
plugGo/
├── entry.go                  # Entry interface and RegFunc type definition
├── app_ctx.go                # GlobalAppCtx global application context
├── boot.go                   # Boot bootstrapper
├── interface.go              # Application, Logger interfaces
├── logger.go                 # Standard logger implementation
├── config/                   # Config loader
│   └── loader.go
└── example/                  # Example code
    ├── announcement/         # Announcement monitor plugin
    │   ├── entry.go         # Entry implementation
    │   ├── handler.go       # Business logic
    │   └── config/
    │       └── config.go    # Config structure
    └── host/                # Host application
        ├── main.go
        └── boot.yaml        # Unified config file
```

## Core Concepts

### Entry Interface

Entry is the base interface for all bootable components, inspired by rk-boot:

```go
type Entry interface {
    Bootstrap(ctx context.Context)  // Start
    Interrupt(ctx context.Context)  // Stop
    GetName() string                // Instance name
    GetType() string                // Entry type
    GetDescription() string         // Description
    String() string                 // String representation
}
```

### RegFunc Registration Function

Registration function creates Entry instances from raw YAML:

```go
type RegFunc func(raw []byte) map[string]Entry
```

### GlobalAppCtx Global Context

Manages all Entry instances, registration functions and shutdown hooks:

```go
// Register Entry registration function
plugGo.RegisterPluginEntryRegFunc(myRegFunc)

// Get Entry
entry := plugGo.GetEntry("MyEntryType", "instance-name")
```

### Boot Bootstrapper

Unified lifecycle management for all Entries:

```go
boot := plugGo.NewBoot(
    plugGo.WithConfigPath("custom-boot.yaml"),
)

// Add hooks
boot.AddHookFuncBeforeBootstrap("EntryType", "name", func(ctx context.Context) {
    // Execute before bootstrap
})

boot.AddShutdownHookFunc("cleanup", func() {
    // Execute on shutdown
})

boot.Bootstrap(ctx)
boot.WaitForShutdownSig(ctx)
```

## Developing Plugins

### Step 1: Define Config Structure

```go
// config/config.go
package config

type Config struct {
    Name     string   `yaml:"name"`
    Enabled  bool     `yaml:"enabled"`
    LogLevel string   `yaml:"logLevel"`
    // Plugin-specific config...
}
```

### Step 2: Implement Entry

```go
// entry.go
package myplugin

import (
    "context"
    "plugGo"
    plugGoConfig "plugGo/config"
)

type MyEntry struct {
    name   string
    cfg    *Config
    logger plugGo.Logger
}

func (e *MyEntry) Bootstrap(ctx context.Context) {
    if !e.cfg.Enabled {
        return
    }
    e.logger.Info("Starting...")
    // Start logic
}

func (e *MyEntry) Interrupt(ctx context.Context) {
    e.logger.Info("Stopping...")
    // Stop logic
}

func (e *MyEntry) GetName() string        { return e.name }
func (e *MyEntry) GetType() string        { return "MyEntry" }
func (e *MyEntry) GetDescription() string { return "My plugin entry" }
func (e *MyEntry) String() string         { return "MyEntry{}" }
```

### Step 3: Implement Registration Function

```go
// entry.go (continued)

func RegisterMyEntry(raw []byte) map[string]plugGo.Entry {
    result := make(map[string]plugGo.Entry)
    
    // Check config section
    if !plugGoConfig.HasYAMLSection(raw, "myplugin") {
        return result
    }
    
    // Parse config
    type bootConfig struct {
        MyPlugin []Config `yaml:"myplugin"`
    }
    var cfg bootConfig
    plugGoConfig.UnmarshalYAML(raw, &cfg)
    
    // Create Entry
    for i := range cfg.MyPlugin {
        entryCfg := &cfg.MyPlugin[i]
        name := entryCfg.Name
        logger := plugGo.NewDefaultLogger(name)
        entry := &MyEntry{name: name, cfg: entryCfg, logger: logger}
        result[name] = entry
    }
    
    return result
}

// Auto-register in init
func init() {
    plugGo.RegisterPluginEntryRegFunc(RegisterMyEntry)
}
```

### Step 4: Configure boot.yaml

```yaml
# boot.yaml
myplugin:
  - name: "instance1"
    enabled: true
    logLevel: "info"
    
  - name: "instance2"
    enabled: true
    logLevel: "debug"
```

## Hook System

### Before/After Bootstrap Hook

```go
boot.AddHookFuncBeforeBootstrap("EntryType", "entryName", func(ctx context.Context) {
    fmt.Println("Before bootstrap")
})

boot.AddHookFuncAfterBootstrap("EntryType", "entryName", func(ctx context.Context) {
    fmt.Println("After bootstrap")
})
```

### Shutdown Hook

```go
boot.AddShutdownHookFunc("cleanup", func() {
    // Cleanup resources
})
```

## Configuration Options

```go
// Custom config file path
boot := plugGo.NewBoot(plugGo.WithConfigPath("config/boot.yaml"))

// Pass raw config content
boot := plugGo.NewBoot(plugGo.WithConfigRaw([]byte(`
announcement:
  - name: "test"
    enabled: true
`)))
```

## Comparison with rk-boot

| Feature | PlugGo | rk-boot |
|---------|--------|---------|
| Config | Unified boot.yaml | Unified boot.yaml |
| Entry Interface | Bootstrap/Interrupt | Bootstrap/Interrupt |
| Registration | RegFunc | RegFunc |
| Hooks | Before/After/Shutdown | Before/After/Shutdown |
| Web Framework | None built-in | gin/echo/fiber etc. |
| Dependencies | Minimal | More |

## Best Practices

1. **Naming Convention**: Use lowercase for Entry names, e.g., `official`, `community`
2. **Multi-Instance**: Configure multiple instances via array in config
3. **Hooks**: Use for logging, monitoring, resource preparation
4. **Graceful Shutdown**: Properly release resources in `Interrupt`

## License

MIT License

## Contributing

Issues and Pull Requests are welcome!
