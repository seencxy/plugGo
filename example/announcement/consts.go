package announcement

// ============================================================================
// 插件配置常量 - 创建新插件时只需修改以下常量
// Plugin Configuration Constants - Modify these when creating a new plugin
// ============================================================================

const (
	// PluginName 插件类型名称，用于配置文件中的 YAML key
	// Plugin type name, used as YAML key in config files
	PluginName = "announcement"

	// PluginVersion 插件版本号
	// Plugin version
	PluginVersion = "1.0.0"

	// EntryTypeName Entry 类型名称
	// Entry type name
	EntryTypeName = "AnnouncementEntry"

	// PluginDescription 插件描述模版，%s 会被实例名替换
	// Plugin description template, %s will be replaced with instance name
	PluginDescription = "Announcement monitor entry [%s]"

	// LoggerPrefix 日志前缀模版，%s 会被实例名替换
	// Logger prefix template, %s will be replaced with instance name
	LoggerPrefix = "announcement-%s"
)
