package version

import (
	"fmt"

	"github.com/na2na/yso-exporter/internal/config"
)

// ConfigAddressResolver implements AddressResolver interface using built-in configuration
// Purpose: Resolve memory addresses from built-in version configuration
type ConfigAddressResolver struct {
	config *config.VersionConfig
}

// NewConfigAddressResolverByVersion creates a resolver for a specific built-in version
func NewConfigAddressResolverByVersion(version string) (*ConfigAddressResolver, error) {
	if version == "" {
		return nil, fmt.Errorf("version name cannot be empty")
	}

	versionConfig, err := config.LoadVersionConfigByName(version)
	if err != nil {
		return nil, fmt.Errorf("failed to load version config for %s: %w", version, err)
	}

	return &ConfigAddressResolver{
		config: versionConfig,
	}, nil
}

// ResolveAddress resolves a memory address for the given metric name
// Returns the resolved address configuration or error if not found
func (r *ConfigAddressResolver) ResolveAddress(metricName string) (*config.AddressConfig, error) {
	if metricName == "" {
		return nil, fmt.Errorf("metric name cannot be empty")
	}

	if r.config == nil {
		return nil, fmt.Errorf("resolver not properly initialized: config is nil")
	}

	addressConfig, exists := r.config.Addresses[metricName]
	if !exists {
		return nil, fmt.Errorf("address not found for metric: %s", metricName)
	}

	// Return a copy of the address config to prevent external modification
	configCopy := addressConfig
	return &configCopy, nil
}

// GetProcessName returns the process name for the current version
func (r *ConfigAddressResolver) GetProcessName() string {
	if r.config == nil {
		return ""
	}

	return r.config.Metadata.ProcessName
}

// GetVersionName returns the version name being resolved
func (r *ConfigAddressResolver) GetVersionName() string {
	if r.config == nil {
		return ""
	}

	return r.config.Metadata.Name
}
