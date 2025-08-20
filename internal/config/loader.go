package config

import (
	"fmt"
)

// LoadVersionConfigByName loads a built-in version configuration
func LoadVersionConfigByName(version string) (*VersionConfig, error) {
	config, ok := GetVersionConfig(version)
	if !ok {
		return nil, fmt.Errorf("version not found: %s", version)
	}
	return config, nil
}
