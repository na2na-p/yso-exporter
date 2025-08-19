package version

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/na2na/yso-exporter/internal/config"
)

// VersionManager manages game version configurations and provides
// version validation and path resolution capabilities.
// Purpose: Centralize version management for the YSO exporter
type VersionManager struct {
	configDir          string
	availableVersions  []string
}

// NewVersionManager creates a new VersionManager instance.
// Purpose: Initialize version manager with specified config directory
func NewVersionManager(configDir string) (*VersionManager, error) {
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("config directory not found: %s", configDir)
	}

	vm := &VersionManager{
		configDir: configDir,
	}

	if err := vm.scanVersions(); err != nil {
		return nil, fmt.Errorf("failed to scan versions: %w", err)
	}

	return vm, nil
}

// ListAvailableVersions returns a list of all available game versions.
// Purpose: Provide available version information for CLI and validation
func (vm *VersionManager) ListAvailableVersions() ([]string, error) {
	if vm.availableVersions == nil {
		if err := vm.scanVersions(); err != nil {
			return nil, fmt.Errorf("failed to scan versions: %w", err)
		}
	}
	
	// Return a copy to prevent external modification
	result := make([]string, len(vm.availableVersions))
	copy(result, vm.availableVersions)
	return result, nil
}

// ValidateVersion checks if the specified version exists.
// Purpose: Ensure version validity before attempting to use it
func (vm *VersionManager) ValidateVersion(version string) error {
	versions, err := vm.ListAvailableVersions()
	if err != nil {
		return fmt.Errorf("failed to get available versions: %w", err)
	}

	for _, v := range versions {
		if v == version {
			return nil
		}
	}

	return fmt.Errorf("version not found: %s", version)
}

// GetVersionPath returns the file path for the specified version config.
// Purpose: Provide file path for version-specific configuration loading
func (vm *VersionManager) GetVersionPath(version string) (string, error) {
	if err := vm.ValidateVersion(version); err != nil {
		return "", err
	}

	return filepath.Join(vm.configDir, version+".yaml"), nil
}

// LoadVersionConfig loads the configuration for the specified version.
// Purpose: Load and parse version-specific game configuration
func (vm *VersionManager) LoadVersionConfig(version string) (*config.VersionConfig, error) {
	configPath, err := vm.GetVersionPath(version)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var versionConfig config.VersionConfig
	if err := yaml.Unmarshal(data, &versionConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	return &versionConfig, nil
}

// CreateAddressResolver creates an AddressResolver for the specified version.
// Purpose: Provide address resolution capabilities for memory access
func (vm *VersionManager) CreateAddressResolver(version string) (AddressResolver, error) {
	versionConfig, err := vm.LoadVersionConfig(version)
	if err != nil {
		return nil, err
	}

	return &defaultAddressResolver{
		versionConfig: versionConfig,
	}, nil
}

// scanVersions scans the config directory for available version files.
// Purpose: Discover available version configurations from filesystem
func (vm *VersionManager) scanVersions() error {
	var versions []string

	err := filepath.WalkDir(vm.configDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(d.Name(), ".yaml") && !strings.HasSuffix(d.Name(), ".yml") {
			return nil
		}

		// Extract version name from filename (remove extension)
		versionName := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
		versions = append(versions, versionName)

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan config directory: %w", err)
	}

	vm.availableVersions = versions
	return nil
}

// defaultAddressResolver implements AddressResolver interface.
// Purpose: Provide concrete implementation of address resolution for game versions
type defaultAddressResolver struct {
	versionConfig *config.VersionConfig
}

// ResolveAddress resolves a memory address for the given metric name.
// Purpose: Convert metric names to memory address configurations
func (r *defaultAddressResolver) ResolveAddress(metricName string) (*config.AddressConfig, error) {
	if r.versionConfig == nil {
		return nil, fmt.Errorf("version config not initialized")
	}

	addressConfig, exists := r.versionConfig.Addresses[metricName]
	if !exists {
		return nil, fmt.Errorf("address not found for metric: %s", metricName)
	}

	return &addressConfig, nil
}

// GetProcessName returns the process name for the current version.
// Purpose: Provide process information for memory access
func (r *defaultAddressResolver) GetProcessName() string {
	if r.versionConfig == nil {
		return ""
	}
	return r.versionConfig.Metadata.ProcessName
}

// GetVersionName returns the version name being resolved.
// Purpose: Provide version identification for logging and debugging
func (r *defaultAddressResolver) GetVersionName() string {
	if r.versionConfig == nil {
		return ""
	}
	return r.versionConfig.Metadata.Name
}