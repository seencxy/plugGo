# PlugGo Plugin Template

A minimal but complete plugin template. Copy and modify to create your own plugin.

## Structure

```
template/
├── entry.go        # Entry implementation and registration
├── factory.go      # Plugin factory (embed config, create instances)
├── plugin.go       # Plugin business logic
├── config.yaml     # Default config (embedded via go:embed)
└── config/
    └── config.go   # Config structure
```

## Usage

### 1. Copy template

```bash
cp -r template yourplugin
```

### 2. Rename (replace all occurrences)

| From | To |
|------|-----|
| `package template` | `package yourplugin` |
| `TemplateEntry` | `YourPluginEntry` |
| `template` (yaml key) | `yourplugin` |
| `template-` (log prefix) | `yourplugin-` |

### 3. Update import paths

```go
// From
"github.com/seencxy/plugGo/template/config"

// To
"your_project/yourplugin/config"
```

### 4. Add custom config fields

Edit `config/config.go`:

```go
type Config struct {
    Name     string `yaml:"name"`
    Enabled  bool   `yaml:"enabled"`
    LogLevel string `yaml:"logLevel"`
    
    // Your fields
    ApiKey   string `yaml:"apiKey"`
    Timeout  int    `yaml:"timeout"`
}
```

### 5. Implement your logic

Edit `plugin.go`:
- `Start()` - startup logic (remember to call `updateStatus(plugGo.StatusRunning, nil)` on success)
- `Stop()` - cleanup logic (remember to call `updateStatus(plugGo.StatusStopped, nil)`)
- `Reload()` - config hot-reload (update status on errors)

**Status Management**: The plugin automatically tracks status changes:
- `StatusIdle` - initial state after creation
- `StatusRunning` - after successful Start()
- `StatusStopped` - after successful Stop()
- `StatusError` - when operations fail

Access status via:
```go
// Get current status
status := plugin.Status()

// Subscribe to status changes
go func() {
    for event := range plugin.StatusNotify() {
        log.Printf("Status: %s, Error: %v", event.Status, event.Error)
    }
}()
```

### 6. Configure boot.yaml

```yaml
yourplugin:
  - name: "instance1"
    enabled: true
    logLevel: "info"
    apiKey: "xxx"
    timeout: 30
```

### 7. Import in main.go

```go
import _ "your_project/yourplugin"
```
