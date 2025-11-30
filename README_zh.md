# PlugGo - Go 插件框架

[English](README.md)

PlugGo 是一个轻量级的 Go 插件框架，参考 [rk-boot](https://github.com/rookie-ninja/rk-boot) 设计，支持从统一的 YAML 配置文件启动和管理多个插件。

## 核心特性

- **统一配置**: 使用单一 `boot.yaml` 管理所有插件配置，支持多实例
- **Entry 接口**: 标准化的生命周期管理 (`Bootstrap`/`Interrupt`)
- **自动注册**: 通过 `import _` 和 `init()` 自动注册插件
- **Hook 机制**: 支持 Before/After Bootstrap 钩子和关闭钩子
- **类型安全**: 编译时静态链接，避免运行时加载的复杂性
- **框架无关**: 不依赖任何特定的 web 框架

## 快速开始

### 1. 安装框架

```bash
go get plugGo
```

### 2. 创建 boot.yaml

```yaml
# boot.yaml
announcement:
  - name: "official"
    enabled: true
    logLevel: "info"
    sources:
      - name: "官网公告"
        url: "https://example.com/api/announcements"
        interval: 300
    notifications:
      - type: "webhook"
        url: "https://your-webhook-url/notify"
```

### 3. 编写 main.go

```go
package main

import (
    "context"
    "plugGo"
    
    // 导入插件（触发 init 自动注册）
    _ "plugGo/example/announcement"
)

func main() {
    // 创建 Boot 引导器（默认读取 boot.yaml）
    boot := plugGo.NewBoot()
    
    // 启动所有 Entry
    boot.Bootstrap(context.Background())
    
    // 等待关闭信号
    boot.WaitForShutdownSig(context.Background())
}
```

### 4. 运行

```bash
cd example/host
go run main.go
```

## 项目结构

```
plugGo/
├── entry.go                  # Entry 接口和 RegFunc 类型定义
├── app_ctx.go                # GlobalAppCtx 全局应用上下文
├── boot.go                   # Boot 引导器
├── interface.go              # Application, Logger 接口
├── logger.go                 # 标准日志实现
├── config/                   # 配置加载器
│   └── loader.go
└── example/                  # 示例代码
    ├── announcement/         # 公告监控插件
    │   ├── entry.go         # Entry 实现
    │   ├── handler.go       # 业务逻辑
    │   └── config/
    │       └── config.go    # 配置结构
    └── host/                # 宿主应用
        ├── main.go
        └── boot.yaml        # 统一配置文件
```

## 核心概念

### Entry 接口

Entry 是所有可启动组件的基础接口，参考 rk-boot 设计：

```go
type Entry interface {
    Bootstrap(ctx context.Context)  // 启动
    Interrupt(ctx context.Context)  // 停止
    GetName() string                // 实例名称
    GetType() string                // Entry 类型
    GetDescription() string         // 描述
    String() string                 // 字符串表示
}
```

### RegFunc 注册函数

注册函数从 raw YAML 创建 Entry 实例：

```go
type RegFunc func(raw []byte) map[string]Entry
```

### GlobalAppCtx 全局上下文

管理所有 Entry 实例、注册函数和关闭钩子：

```go
// 注册 Entry 注册函数
plugGo.RegisterPluginEntryRegFunc(myRegFunc)

// 获取 Entry
entry := plugGo.GetEntry("MyEntryType", "instance-name")
```

### Boot 引导器

统一管理所有 Entry 的生命周期：

```go
boot := plugGo.NewBoot(
    plugGo.WithConfigPath("custom-boot.yaml"),
)

// 添加钩子
boot.AddHookFuncBeforeBootstrap("EntryType", "name", func(ctx context.Context) {
    // 启动前执行
})

boot.AddShutdownHookFunc("cleanup", func() {
    // 关闭时执行
})

boot.Bootstrap(ctx)
boot.WaitForShutdownSig(ctx)
```

## 开发插件

### 第一步：定义配置结构

```go
// config/config.go
package config

type Config struct {
    Name     string   `yaml:"name"`
    Enabled  bool     `yaml:"enabled"`
    LogLevel string   `yaml:"logLevel"`
    // 插件特定配置...
}
```

### 第二步：实现 Entry

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
    // 启动逻辑
}

func (e *MyEntry) Interrupt(ctx context.Context) {
    e.logger.Info("Stopping...")
    // 停止逻辑
}

func (e *MyEntry) GetName() string        { return e.name }
func (e *MyEntry) GetType() string        { return "MyEntry" }
func (e *MyEntry) GetDescription() string { return "My plugin entry" }
func (e *MyEntry) String() string         { return "MyEntry{}" }
```

### 第三步：实现注册函数

```go
// entry.go (续)

func RegisterMyEntry(raw []byte) map[string]plugGo.Entry {
    result := make(map[string]plugGo.Entry)
    
    // 检查配置节点
    if !plugGoConfig.HasYAMLSection(raw, "myplugin") {
        return result
    }
    
    // 解析配置
    type bootConfig struct {
        MyPlugin []Config `yaml:"myplugin"`
    }
    var cfg bootConfig
    plugGoConfig.UnmarshalYAML(raw, &cfg)
    
    // 创建 Entry
    for i := range cfg.MyPlugin {
        entryCfg := &cfg.MyPlugin[i]
        name := entryCfg.Name
        logger := plugGo.NewDefaultLogger(name)
        entry := &MyEntry{name: name, cfg: entryCfg, logger: logger}
        result[name] = entry
    }
    
    return result
}

// init 自动注册
func init() {
    plugGo.RegisterPluginEntryRegFunc(RegisterMyEntry)
}
```

### 第四步：配置 boot.yaml

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

## Hook 机制

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
    // 清理资源
})
```

## 配置选项

```go
// 自定义配置文件路径
boot := plugGo.NewBoot(plugGo.WithConfigPath("config/boot.yaml"))

// 直接传入配置内容
boot := plugGo.NewBoot(plugGo.WithConfigRaw([]byte(`
announcement:
  - name: "test"
    enabled: true
`)))
```

## 与 rk-boot 的对比

| 特性 | PlugGo | rk-boot |
|------|--------|---------|
| 配置方式 | 统一 boot.yaml | 统一 boot.yaml |
| Entry 接口 | Bootstrap/Interrupt | Bootstrap/Interrupt |
| 注册机制 | RegFunc | RegFunc |
| Hook 支持 | Before/After/Shutdown | Before/After/Shutdown |
| Web 框架 | 无内置 | 支持 gin/echo/fiber 等 |
| 依赖 | 极简 | 较多 |

## 最佳实践

1. **命名规范**: Entry 名称使用小写字母，如 `official`, `community`
2. **多实例**: 同类型插件通过数组配置多实例
3. **Hook 使用**: 用于日志、监控、资源准备等
4. **优雅关闭**: 在 `Interrupt` 中正确释放资源

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
