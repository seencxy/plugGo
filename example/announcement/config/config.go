package config

// Config is the config structure for announcement monitor plugin.
// Supports the new boot.yaml unified config format.
type Config struct {
	// Basic config (directly under entry node)
	Name     string `yaml:"name"`     // Instance name
	Enabled  bool   `yaml:"enabled"`  // Whether enabled
	LogLevel string `yaml:"logLevel"` // Log level

	// Announcement sources config
	Sources []Source `yaml:"sources"`

	// Notification config
	Notifications []Notification `yaml:"notifications"`

	// Filters config
	Filters Filters `yaml:"filters"`
}

// Source is the announcement source config.
type Source struct {
	Name     string `yaml:"name"`     // Source name
	URL      string `yaml:"url"`      // Source URL
	Interval int    `yaml:"interval"` // Polling interval (seconds)
}

// Notification is the notification config.
type Notification struct {
	Type string `yaml:"type"` // Notification type (webhook, email, etc.)
	URL  string `yaml:"url"`  // Webhook URL or other address
}

// Filters is the filters config.
type Filters struct {
	Keywords []string `yaml:"keywords"` // Keywords list
}
