package config

import (
	"os"
	"strings"
	"time"
)

func ListAvailableVersions() ([]string, error) {
	configDir := "configs/versions"
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return nil, err
	}

	var versions []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			version := strings.TrimSuffix(entry.Name(), ".yaml")
			versions = append(versions, version)
		}
	}

	return versions, nil
}

func LoadConfig(version string) (*AppConfig, error) {
	_, err := LoadVersionConfigByName(version)
	if err != nil {
		return nil, err
	}
	
	// Default values
	return &AppConfig{
		Version:  version,
		Port:     8080,
		Interval: 100 * time.Millisecond,
	}, nil
}