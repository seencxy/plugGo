# PlugGo Plugin Template

A minimal but complete plugin template. Copy and modify to create your own plugin.

## Structure

```
template/
├── entry.go        # Entry implementation and registration
├── plugin.go       # Plugin business logic
├── config.yaml     # Default config (embedded)
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
- `Start()` - startup logic
- `Stop()` - cleanup logic
- `Reload()` - config hot-reload

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
