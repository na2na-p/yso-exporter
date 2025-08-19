package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	
	"github.com/goccy/go-yaml"
)

// LoadVersionConfig loads a version configuration from a YAML file
func LoadVersionConfig(path string) (*VersionConfig, error) {
	if path == "" {
		return nil, fmt.Errorf("config path cannot be empty")
	}
	
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}
	
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Parse YAML
	var config VersionConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	
	// Validate required fields
	if config.Metadata.Name == "" {
		return nil, fmt.Errorf("metadata.name is required")
	}
	if config.Metadata.ProcessName == "" {
		return nil, fmt.Errorf("metadata.process_name is required")
	}
	
	return &config, nil
}

// LoadAppConfig loads application configuration from a YAML file
func LoadAppConfig(path string) (*AppConfig, error) {
	if path == "" {
		// Return default config if no path specified
		return &AppConfig{
			Version:  "jp_1.1.0.3",
			Port:     8080,
			Interval: 100 * 1000 * 1000, // 100ms in nanoseconds
		}, nil
	}
	
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", path)
	}
	
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Parse YAML with custom structure for duration handling
	var rawConfig struct {
		Version  string `yaml:"version"`
		Port     int    `yaml:"port"`
		Interval string `yaml:"interval"`
	}
	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	
	// Parse interval duration
	interval, err := parseDuration(rawConfig.Interval)
	if err != nil {
		return nil, fmt.Errorf("invalid interval format: %w", err)
	}
	
	config := &AppConfig{
		Version:  rawConfig.Version,
		Port:     rawConfig.Port,
		Interval: interval,
	}
	
	// Validate
	if config.Port < 1 || config.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", config.Port)
	}
	if config.Interval < 10*1000*1000 { // 10ms in nanoseconds
		return nil, fmt.Errorf("interval too short: %v", config.Interval)
	}
	
	return config, nil
}

// parseDuration handles both string duration format (e.g., "100ms") and nanoseconds as integer
func parseDuration(s string) (time.Duration, error) {
	// First try to parse as a duration string
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}
	
	// Try to parse as nanoseconds integer
	var nanos int64
	if _, err := fmt.Sscanf(s, "%d", &nanos); err == nil {
		return time.Duration(nanos), nil
	}
	
	return 0, fmt.Errorf("cannot parse duration: %s", s)
}

// LoadVersionConfigByName loads a version config from the standard configs directory
func LoadVersionConfigByName(version string) (*VersionConfig, error) {
	configPath := filepath.Join("configs", "versions", version+".yaml")
	return LoadVersionConfig(configPath)
}