package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Loader is a config loader that supports loading YAML config from file or embedded data.
type Loader struct{}

// LoadWithFallback loads config with host project config overriding plugin default config.
// Priority: host project config > plugin default config
//
// Parameters:
//   - pluginName: plugin name (used to build config file path)
//   - defaultConfigData: plugin embedded default config data (obtained via go:embed)
//   - target: config struct pointer to store parsed config
//
// Returns:
//   - error: returns error if loading fails
func (l *Loader) LoadWithFallback(
	pluginName string,
	defaultConfigData []byte,
	target interface{},
) error {
	// 1. Try to load config from host project
	// Path format: plugins/{pluginName}/config.yaml
	hostConfigPath := filepath.Join("plugins", pluginName, "config.yaml")
	if data, err := os.ReadFile(hostConfigPath); err == nil {
		return yaml.Unmarshal(data, target)
	}

	// 2. Use plugin embedded default config
	if defaultConfigData != nil {
		return yaml.Unmarshal(defaultConfigData, target)
	}

	return fmt.Errorf("no config found for plugin: %s", pluginName)
}

// Load loads config file directly from specified path.
//
// Parameters:
//   - path: full path to config file
//   - target: config struct pointer to store parsed config
//
// Returns:
//   - error: returns error if loading fails
func (l *Loader) Load(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, target)
}

// UnmarshalYAML parses config from raw YAML data.
// This is the core method of the new architecture for extracting plugin config from unified boot.yaml.
//
// Parameters:
//   - raw: raw YAML data (entire boot.yaml content)
//   - target: config struct pointer
//
// Returns:
//   - error: returns error if parsing fails
func UnmarshalYAML(raw []byte, target interface{}) error {
	return yaml.Unmarshal(raw, target)
}

// UnmarshalYAMLSection parses config for a specific section from raw YAML data.
// Used to extract specific plugin config section from boot.yaml.
//
// Parameters:
//   - raw: raw YAML data
//   - section: config section name (e.g. "announcement")
//   - target: config struct pointer
//
// Returns:
//   - error: returns error if parsing fails
func UnmarshalYAMLSection(raw []byte, section string, target interface{}) error {
	// First parse as map
	var rootMap map[string]interface{}
	if err := yaml.Unmarshal(raw, &rootMap); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Get specified section
	sectionData, ok := rootMap[section]
	if !ok {
		return fmt.Errorf("section '%s' not found in config", section)
	}

	// Re-encode section data to YAML
	sectionYAML, err := yaml.Marshal(sectionData)
	if err != nil {
		return fmt.Errorf("failed to marshal section: %w", err)
	}

	// Parse into target struct
	return yaml.Unmarshal(sectionYAML, target)
}

// GetYAMLSection gets raw data of specified section from raw YAML.
//
// Parameters:
//   - raw: raw YAML data
//   - section: config section name
//
// Returns:
//   - []byte: YAML data of the section
//   - error: returns error if parsing fails
func GetYAMLSection(raw []byte, section string) ([]byte, error) {
	var rootMap map[string]interface{}
	if err := yaml.Unmarshal(raw, &rootMap); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	sectionData, ok := rootMap[section]
	if !ok {
		return nil, fmt.Errorf("section '%s' not found in config", section)
	}

	return yaml.Marshal(sectionData)
}

// HasYAMLSection checks if a specified section exists in YAML.
func HasYAMLSection(raw []byte, section string) bool {
	var rootMap map[string]interface{}
	if err := yaml.Unmarshal(raw, &rootMap); err != nil {
		return false
	}
	_, ok := rootMap[section]
	return ok
}
