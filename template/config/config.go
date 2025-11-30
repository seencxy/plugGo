package config

// Config is the plugin configuration structure.
type Config struct {
	Name     string `yaml:"name"`     // Instance name
	Enabled  bool   `yaml:"enabled"`  // Whether enabled
	LogLevel string `yaml:"logLevel"` // Log level: debug, info, warn, error

	// Add your custom config fields here
	Interval int    `yaml:"interval"` // Example: polling interval in seconds
	Endpoint string `yaml:"endpoint"` // Example: API endpoint
}
